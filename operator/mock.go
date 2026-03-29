package operator

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	_ IS3  = (*mockOperator)(nil)
	_ IKMS = (*mockOperator)(nil)
)

// mockOperator is a mock implementation of the IS3 and IKMS interfaces for testing.
type mockOperator struct {
	getObjectFunc             func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	putObjectFunc             func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	deleteObjectFunc          func(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	headObjectFunc            func(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
	newObjectExistsWaiterFunc func(opts ...func(*s3.ObjectExistsWaiterOptions)) *s3.ObjectExistsWaiter
	newUploaderFunc           func(opts ...func(*manager.Uploader)) *manager.Uploader
	createMultipartUploadFunc func(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error)
	uploadPartFunc            func(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error)
	completeMultipartFunc     func(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error)
	abortMultipartFunc        func(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error)
	generateDataKeyFunc       func(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error)
	decryptFunc               func(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
	w                         io.Writer
}

// GetObject calls the mocked GetObject function.
func (o *mockOperator) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return o.getObjectFunc(ctx, params, optFns...)
}

// PutObject calls the mocked PutObject function.
func (o *mockOperator) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return o.putObjectFunc(ctx, params, optFns...)
}

// DeleteObject calls the mocked DeleteObject function.
func (o *mockOperator) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	return o.deleteObjectFunc(ctx, params, optFns...)
}

// HeadObject calls the mocked HeadObject function.
func (o *mockOperator) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return o.headObjectFunc(ctx, params, optFns...)
}

// NewObjectExistsWaiter calls the mocked NewObjectExistsWaiter function.
func (o *mockOperator) NewObjectExistsWaiter(opts ...func(*s3.ObjectExistsWaiterOptions)) *s3.ObjectExistsWaiter {
	return o.newObjectExistsWaiterFunc(opts...)
}

// NewUploader calls the mocked NewUploader function.
func (o *mockOperator) NewUploader(opts ...func(*manager.Uploader)) *manager.Uploader {
	return o.newUploaderFunc(opts...)
}

// GenerateDataKey calls the mocked GenerateDataKey function.
func (o *mockOperator) GenerateDataKey(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
	return o.generateDataKeyFunc(ctx, params, optFns...)
}

// Decrypt calls the mocked Decrypt function.
func (o *mockOperator) Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
	return o.decryptFunc(ctx, params, optFns...)
}

// CreateMultipartUpload calls the mocked CreateMultipartUpload function.
func (o *mockOperator) CreateMultipartUpload(ctx context.Context, params *s3.CreateMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CreateMultipartUploadOutput, error) {
	return o.createMultipartUploadFunc(ctx, params, optFns...)
}

// UploadPart calls the mocked UploadPart function.
func (o *mockOperator) UploadPart(ctx context.Context, params *s3.UploadPartInput, optFns ...func(*s3.Options)) (*s3.UploadPartOutput, error) {
	return o.uploadPartFunc(ctx, params, optFns...)
}

// CompleteMultipartUpload calls the mocked CompleteMultipartUpload function.
func (o *mockOperator) CompleteMultipartUpload(ctx context.Context, params *s3.CompleteMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.CompleteMultipartUploadOutput, error) {
	return o.completeMultipartFunc(ctx, params, optFns...)
}

// AbortMultipartUpload calls the mocked AbortMultipartUpload function.
func (o *mockOperator) AbortMultipartUpload(ctx context.Context, params *s3.AbortMultipartUploadInput, optFns ...func(*s3.Options)) (*s3.AbortMultipartUploadOutput, error) {
	return o.abortMultipartFunc(ctx, params, optFns...)
}
