package health

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	_ IS3  = (*S3)(nil)
	_ IKMS = (*KMS)(nil)
)

// IS3 defines the interface for S3 client.
type IS3 interface {
	HeadBucket(ctx context.Context, params *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

// IKMS defines the interface for KMS client.
type IKMS interface {
	Encrypt(ctx context.Context, params *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error)
	Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

// ICFN defines the interface for CloudFormation client.
type ICFN interface {
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
}

// S3 defines a struct for S3 client.
type S3 struct {
	*s3.Client
}

// KMS defines a struct for KMS client.
type KMS struct {
	*kms.Client
}

// CFN defines a struct for CloudFormation client.
type CFN struct {
	*cloudformation.Client
}

// Checker is a client for health checks.
type Checker struct {
	s3  IS3
	kms IKMS
	cfn ICFN
	w   io.Writer
}

// NewObjectExistsWaiter creates a new ObjectExistsWaiter.
func (o *S3) NewObjectExistsWaiter(opts ...func(*s3.ObjectExistsWaiterOptions)) *s3.ObjectExistsWaiter {
	return s3.NewObjectExistsWaiter(o.Client, opts...)
}

// New creates a new Checker.
func New(w io.Writer, cfg aws.Config) *Checker {
	return &Checker{
		s3:  &S3{Client: s3.NewFromConfig(cfg)},
		kms: &KMS{Client: kms.NewFromConfig(cfg)},
		cfn: &CFN{Client: cloudformation.NewFromConfig(cfg)},
		w:   w,
	}
}
