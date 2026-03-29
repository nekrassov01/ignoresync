package health

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nekrassov01/ignoresync"
	"github.com/nekrassov01/ignoresync/color"
	"github.com/nekrassov01/ignoresync/manager"
	"golang.org/x/sync/errgroup"
)

// Check performs validation checks on the given state and account.
func (o *Checker) Check(ctx context.Context, state *manager.State) error {
	entries := []struct {
		label string
		fn    func(context.Context, *manager.State) (bool, error)
	}{
		{
			label: "local state validity",
			fn:    o.CheckState,
		},
		{
			label: "cloudformation stack completeness",
			fn:    o.CheckStack,
		},
		{
			label: "bucket accessibility round-trip check",
			fn:    o.CheckBucket,
		},
		{
			label: "kms sse key round-trip check",
			fn:    o.CheckSSEKey,
		},
		{
			label: "kms cse key round-trip check",
			fn:    o.CheckCSEKey,
		},
	}
	eg := errgroup.Group{}
	msgCh := make(chan string, len(entries))
	for _, entry := range entries {
		eg.Go(func() error {
			ok, err := entry.fn(ctx, state)
			if ok {
				msgCh <- fmt.Sprintf("%s %s", color.Success("OK"), entry.label)
			} else {
				msgCh <- fmt.Sprintf("%s %s: %v", color.Error("NG"), entry.label, err)
			}
			return nil
		})
	}
	_ = eg.Wait()
	close(msgCh)
	for msg := range msgCh {
		_, _ = fmt.Fprintln(o.w, msg)
	}
	return nil
}

// CheckState checks if the given state is complete and valid.
func (o *Checker) CheckState(_ context.Context, state *manager.State) (bool, error) {
	if state == nil ||
		len(state.MasterKeys[state.KeyID]) != ignoresync.MasterKeySize ||
		state.Account == "" ||
		state.Region == "" ||
		state.Bucket == nil ||
		state.Bucket.ARN.Resource == "" ||
		state.SSEKey == nil ||
		state.SSEKey.ARN.Resource == "" ||
		state.CSEKey == nil ||
		state.CSEKey.ARN.Resource == "" {
		return false, NewStateError(errors.New("invalid state"))
	}
	return true, nil
}

// CheckStack checks if the CloudFormation stack exists and is in a complete state.
func (o *Checker) CheckStack(ctx context.Context, state *manager.State) (bool, error) {
	opt := func(opt *cloudformation.Options) {
		opt.Region = state.Region
	}
	in := &cloudformation.DescribeStacksInput{
		StackName: aws.String(ignoresync.CanonicalName),
	}
	out, err := o.cfn.DescribeStacks(ctx, in, opt)
	if err != nil {
		return false, NewStackError(errors.New("failed to describe stacks"))
	}
	status := (out.Stacks[0].StackStatus)
	if status == (cfntypes.StackStatusDeleteComplete) {
		return false, NewStackError(errors.New("stack already deleted"))
	}
	if !strings.HasSuffix(string(status), "_COMPLETE") {
		return false, NewStackError(errors.New("stack not completed"))
	}
	return true, nil
}

// CheckBucket performs a head, put, get and delete round-trip test.
func (o *Checker) CheckBucket(ctx context.Context, state *manager.State) (bool, error) {
	opt := func(opt *s3.Options) {
		opt.Region = state.Region
	}
	const prefix = "healthcheck"
	body := []byte("healthcheck")

	// Step 1: Head
	in1 := &s3.HeadBucketInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	if _, err := o.s3.HeadBucket(ctx, in1, opt); err != nil {
		return false, NewBucketError(errors.New("failed to head bucket"))
	}

	// Step 2: Put denied
	in2 := &s3.PutObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		Body:                bytes.NewReader(body),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	if _, err := o.s3.PutObject(ctx, in2, opt); err == nil {
		return false, NewBucketError(errors.New("put object succeeded without expected permissions"))
	}

	// Step 3: Put succeeded
	if state.SSEKey == nil {
		return false, NewBucketError(errors.New("sse key not set in state"))
	}
	b, err := json.Marshal(state.SSEKey.EncryptionContext)
	if err != nil {
		return false, NewBucketError(errors.New("failed to marshal sse key encryption context"))
	}
	in3 := &s3.PutObjectInput{
		Bucket:                  aws.String(state.Bucket.ARN.Resource),
		Key:                     aws.String(prefix),
		Body:                    bytes.NewReader(body),
		ExpectedBucketOwner:     aws.String(state.Account),
		ServerSideEncryption:    s3types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:             aws.String(state.SSEKey.ARN.String()),
		SSEKMSEncryptionContext: aws.String(base64.StdEncoding.EncodeToString(b)),
		ChecksumAlgorithm:       s3types.ChecksumAlgorithmSha256,
	}
	if _, err := o.s3.PutObject(ctx, in3, opt); err != nil {
		return false, NewBucketError(errors.New("failed to put object"))
	}

	// Step 4: Get
	in4 := &s3.GetObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	out, err := o.s3.GetObject(ctx, in4, opt)
	if err != nil {
		return false, NewBucketError(errors.New("failed to get object"))
	}
	got, err := io.ReadAll(out.Body)
	if err != nil {
		return false, NewBucketError(errors.New("failed to read object"))
	}
	defer func() {
		closeErr := out.Body.Close()
		if err == nil && closeErr != nil {
			err = NewBucketError(errors.New("failed to close object"))
		}
	}()
	if !bytes.Equal(got, body) {
		return false, NewBucketError(errors.New("content mismatch"))
	}

	// Step 5: Delete
	in5 := &s3.DeleteObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	if _, err := o.s3.DeleteObject(ctx, in5, opt); err != nil {
		return false, NewBucketError(errors.New("failed to delete object"))
	}

	return true, nil
}

// CheckSSEKey performs a round-trip encryption/decryption test using the SSE KMS key to verify the existence and validity of the key.
func (o *Checker) CheckSSEKey(ctx context.Context, state *manager.State) (bool, error) {
	ok, err := o.checkKey(ctx, state.SSEKey, state.Region)
	if err != nil {
		return false, NewKeyError(err)
	}
	return ok, nil
}

// CheckCSEKey performs a round-trip encryption/decryption test using the CSE KMS key to verify the existence and validity of the key.
func (o *Checker) CheckCSEKey(ctx context.Context, state *manager.State) (bool, error) {
	ok, err := o.checkKey(ctx, state.CSEKey, state.Region)
	if err != nil {
		return false, NewKeyError(err)
	}
	return ok, nil
}

// checkKey performs a round-trip encryption/decryption test.
func (o *Checker) checkKey(ctx context.Context, target *manager.KMSKeyInfo, region string) (bool, error) {
	opt := func(opt *kms.Options) {
		opt.Region = region
	}
	plaintext := []byte("healthcheck")

	// Step 1: Decrypt OK
	in1 := &kms.EncryptInput{
		KeyId:             aws.String(target.ARN.String()),
		Plaintext:         plaintext,
		EncryptionContext: target.EncryptionContext,
	}
	out1, err := o.kms.Encrypt(ctx, in1, opt)
	if err != nil {
		return false, errors.New("failed to encrypt")
	}
	in2 := &kms.DecryptInput{
		KeyId:             aws.String(target.ARN.String()),
		CiphertextBlob:    out1.CiphertextBlob,
		EncryptionContext: target.EncryptionContext,
	}
	out2, err := o.kms.Decrypt(ctx, in2, opt)
	if err != nil {
		return false, (errors.New("failed to decrypt"))
	}
	if !bytes.Equal(out2.Plaintext, plaintext) {
		return false, errors.New("decrypted plaintext mismatch")
	}

	// Step 2: Invalid context
	in3 := &kms.DecryptInput{
		KeyId:             aws.String(target.ARN.String()),
		CiphertextBlob:    out1.CiphertextBlob,
		EncryptionContext: nil,
	}
	if _, err := o.kms.Decrypt(ctx, in3, opt); err == nil {
		return false, errors.New("security flaw: decrypted with invalid context")
	}

	// Step 3: Tampered data
	blob := make([]byte, len(out1.CiphertextBlob))
	copy(blob, out1.CiphertextBlob)
	if len(blob) > 0 {
		// Flip the last byte of the data to simulate tampering
		blob[len(blob)-1] ^= 0xFF
	}
	in4 := &kms.DecryptInput{
		KeyId:             aws.String(target.ARN.String()),
		CiphertextBlob:    blob,
		EncryptionContext: target.EncryptionContext,
	}
	if _, err := o.kms.Decrypt(ctx, in4, opt); err == nil {
		return false, errors.New("security flaw: decrypted tampered ciphertext")
	}
	return true, nil
}
