package operator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3m "github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3c "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/nekrassov01/ignoresync/manager"
)

// upload uploads an encrypted object with the given metadata.
func (o *Operator) upload(ctx context.Context, state *manager.State, body io.Reader, prefix string, metadata map[string]string, etag *string) error {
	encryptionContext, err := json.Marshal(state.SSEKey.EncryptionContext)
	if err != nil {
		return fmt.Errorf("failed to marshal sse encryption context: %w", err)
	}
	opt := func(u *s3m.Uploader) {
		u.PartSize = uploadPartSize
		u.Concurrency = uploadConcurrency
		u.ClientOptions = []func(o *s3.Options){
			func(o *s3.Options) {
				o.Region = state.Region
				o.Retryer = retryer
			},
		}
	}
	in := &s3c.PutObjectInput{
		Bucket:                  aws.String(state.Bucket.ARN.Resource),
		Key:                     aws.String(prefix),
		Body:                    body,
		ExpectedBucketOwner:     aws.String(state.Account),
		ServerSideEncryption:    types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:             aws.String(state.SSEKey.ARN.String()),
		SSEKMSEncryptionContext: aws.String(base64.StdEncoding.EncodeToString(encryptionContext)),
		ChecksumAlgorithm:       types.ChecksumAlgorithmSha256,
		Metadata:                metadata,
		IfMatch:                 etag,
	}
	uploader := o.s3.NewUploader(opt)
	if _, err = uploader.Upload(ctx, in); err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}
	return o.waitUpload(ctx, state, prefix)
}

// download downloads an object from S3 with the given prefix and returns its body and parsed metadata.
func (o *Operator) download(ctx context.Context, state *manager.State, prefix string) (io.ReadCloser, *metadata, *string, error) {
	opt := func(o *s3.Options) {
		o.Region = state.Region
		o.Retryer = retryer
	}
	in := &s3.GetObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
		ChecksumMode:        types.ChecksumModeEnabled,
	}
	out, err := o.s3.GetObject(ctx, in, opt)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get object: %w", err)
	}
	m, err := parseMetadata(out.Metadata)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return out.Body, m, out.ETag, nil
}

// // head retrieves metadata for objects with the specified prefix from S3, parses it, and returns the results.
func (o *Operator) head(ctx context.Context, state *manager.State, prefix string) (*metadata, *string, error) {
	opt := func(o *s3.Options) {
		o.Region = state.Region
		o.Retryer = retryer
	}
	in := &s3.HeadObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
		ChecksumMode:        types.ChecksumModeEnabled,
	}
	out, err := o.s3.HeadObject(ctx, in, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to head object: %w", err)
	}
	m, err := parseMetadata(out.Metadata)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse metadata: %w", err)
	}
	return m, out.ETag, nil
}

// delete deletes an object from S3 with the given prefix.
func (o *Operator) delete(ctx context.Context, state *manager.State, prefix string) error {
	opt := func(o *s3.Options) {
		o.Region = state.Region
		o.Retryer = retryer
	}
	in := &s3.DeleteObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	if _, err := o.s3.DeleteObject(ctx, in, opt); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// waitUpload waits until the uploaded object is available in S3, to ensure subsequent download can succeed.
func (o *Operator) waitUpload(ctx context.Context, state *manager.State, prefix string) error {
	opt := func(opt *s3.ObjectExistsWaiterOptions) {
		opt.ClientOptions = []func(o *s3.Options){
			func(o *s3.Options) {
				o.Region = state.Region
				o.Retryer = retryer
			},
		}
	}
	in := &s3c.HeadObjectInput{
		Bucket:              aws.String(state.Bucket.ARN.Resource),
		Key:                 aws.String(prefix),
		ExpectedBucketOwner: aws.String(state.Account),
	}
	waiter := o.s3.NewObjectExistsWaiter(opt)
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
