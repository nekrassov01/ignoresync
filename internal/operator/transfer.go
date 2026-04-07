package operator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	transfertypes "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/nekrassov01/ignoresync/internal/manager"
)

// upload uploads an encrypted object with the given metadata.
func (o *Operator) upload(ctx context.Context, state *manager.State, body io.Reader, prefix string, metadata map[string]string, etag *string) error {
	encryptionContext, err := json.Marshal(state.SSEKey.EncryptionContext)
	if err != nil {
		return fmt.Errorf("failed to marshal sse encryption context: %w", err)
	}
	opt := func(t *transfermanager.Options) {
		t.PartSizeBytes = uploadPartSize
		t.Concurrency = uploadConcurrency
	}
	in := &transfermanager.UploadObjectInput{
		Bucket:                  aws.String(state.Bucket.ARN.Resource),
		Key:                     aws.String(prefix),
		Body:                    body,
		ExpectedBucketOwner:     aws.String(state.Account),
		ServerSideEncryption:    transfertypes.ServerSideEncryptionAwsKms,
		SSEKMSKeyID:             aws.String(state.SSEKey.ARN.String()),
		SSEKMSEncryptionContext: aws.String(base64.StdEncoding.EncodeToString(encryptionContext)),
		ChecksumAlgorithm:       transfertypes.ChecksumAlgorithmSha256,
		Metadata:                metadata,
		IfMatch:                 etag,
	}
	transfer := o.s3.NewTransferManager()
	if _, err := transfer.UploadObject(ctx, in, opt); err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return o.waitUpload(ctx, state, prefix)
}

// download downloads an object from S3 with the given prefix and returns its body and parsed metadata.
func (o *Operator) download(ctx context.Context, state *manager.State, prefix string) (io.ReadCloser, *metadata, *string, error) {
	in := &s3.GetObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
		ChecksumMode:        s3types.ChecksumModeEnabled,
	}
	out, err := o.s3.GetObject(ctx, in)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get object: %w", err)
	}
	meta, err := parseMetadata(out.Metadata)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return out.Body, meta, out.ETag, nil
}

// // head retrieves metadata for objects with the specified prefix from S3, parses it, and returns the results.
func (o *Operator) head(ctx context.Context, state *manager.State, prefix string) (*metadata, *string, error) {
	in := &s3.HeadObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
		ChecksumMode:        s3types.ChecksumModeEnabled,
	}
	out, err := o.s3.HeadObject(ctx, in)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to head object: %w", err)
	}
	meta, err := parseMetadata(out.Metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return meta, out.ETag, nil
}

// delete deletes an object from S3 with the given prefix.
func (o *Operator) delete(ctx context.Context, state *manager.State, prefix string) error {
	in := &s3.DeleteObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	if _, err := o.s3.DeleteObject(ctx, in); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// waitUpload waits until the uploaded object is available in S3, to ensure subsequent download can succeed.
func (o *Operator) waitUpload(ctx context.Context, state *manager.State, prefix string) error {
	in := &s3.HeadObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	waiter := o.s3.NewObjectExistsWaiter()
	if err := waiter.Wait(ctx, in, uploadMaxWaitDur); err != nil {
		return fmt.Errorf("failed to wait upload: %w", err)
	}
	return nil
}

// isObjectNotFound checks if the error is a "not found" error from S3.
func isObjectNotFound(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		code := apiErr.ErrorCode()
		return code == "NoSuchKey" || code == "NotFound"
	}
	return false
}

// isConditionalError checks if the error is a "PreconditionFailed" or "ConditionalRequestConflict" error from S3.
// see: https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html#ErrorCodeList
func isConditionalError(err error) bool {
	if err == nil {
		return false
	}
	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		code := apiErr.ErrorCode()
		return code == "PreconditionFailed" || code == "ConditionalRequestConflict"
	}
	return false
}
