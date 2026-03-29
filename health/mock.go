package health

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	_ ICFN = (*mockChecker)(nil)
	_ IS3  = (*mockChecker)(nil)
	_ IKMS = (*mockChecker)(nil)
)

// mockChecker is a mock implementation of the IS3 and IKMS interfaces for testing.
type mockChecker struct {
	describeStacksFunc func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	headBucketFunc     func(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	getObjectFunc      func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	putObjectFunc      func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	deleteObjectFunc   func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	encryptFunc        func(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error)
	decryptFunc        func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
	w                  io.Writer
}

// DescribeStacks calls the mocked DescribeStacks function.
func (o *mockChecker) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return o.describeStacksFunc(ctx, params, optFns...)
}

// HeadBucket calls the mocked HeadBucket function.
func (o *mockChecker) HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
	return o.headBucketFunc(ctx, params, optFns...)
}

// GetObject calls the mocked GetObject function.
func (o *mockChecker) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return o.getObjectFunc(ctx, params, optFns...)
}

// PutObject calls the mocked PutObject function.
func (o *mockChecker) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return o.putObjectFunc(ctx, params, optFns...)
}

// DeleteObject calls the mocked DeleteObject function.
func (o *mockChecker) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return o.deleteObjectFunc(ctx, params, optFns...)
}

// Encrypt calls the mocked Encrypt function.
func (o *mockChecker) Encrypt(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error) {
	return o.encryptFunc(ctx, params, optFns...)
}

// Decrypt calls the mocked Decrypt function.
func (o *mockChecker) Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
	return o.decryptFunc(ctx, params, optFns...)
}
