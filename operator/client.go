package operator

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	_ IS3  = (*S3)(nil)
	_ IKMS = (*KMS)(nil)
)

// IS3 defines the interface for S3 client.
type IS3 interface {
	transfermanager.S3APIClient
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	NewObjectExistsWaiter(opts ...func(*s3.ObjectExistsWaiterOptions)) *s3.ObjectExistsWaiter
	NewTransferManager() *transfermanager.Client
}

// IKMS defines the interface for KMS client.
type IKMS interface {
	GenerateDataKey(ctx context.Context, params *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error)
	Decrypt(ctx context.Context, params *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error)
}

// S3 defines a struct for S3 client.
type S3 struct {
	*s3.Client
}

// KMS defines a struct for KMS client.
type KMS struct {
	*kms.Client
}

// Operator is a client for operator.
type Operator struct {
	s3             IS3       // s3 client for upload/download
	kms            IKMS      // kms client for client side encryption
	repo           *RepoInfo // repository information
	prefixFiles    string    // s3 prefix for files
	prefixPatterns string    // s3 prefix for patterns
	dryrun         bool      // whether to perform a dry run
	overwrite      bool      // whether to force download even if the same version exists
	w              io.Writer // writer for output
}

// NewObjectExistsWaiter creates a new ObjectExistsWaiter.
func (o *S3) NewObjectExistsWaiter(opts ...func(*s3.ObjectExistsWaiterOptions)) *s3.ObjectExistsWaiter {
	return s3.NewObjectExistsWaiter(o.Client, opts...)
}

// NewTransferManager creates a new S3 transfer manager.
func (o *S3) NewTransferManager() *transfermanager.Client {
	return transfermanager.New(o.Client)
}

// New creates a new Operator.
func New(w io.Writer, repoPath, remoteName string, cfg aws.Config) (*Operator, error) {
	// Initialize client
	o := &Operator{
		s3:  &S3{Client: s3.NewFromConfig(cfg)},
		kms: &KMS{Client: kms.NewFromConfig(cfg)},
		w:   w,
	}

	// Get repository information
	repo, err := NewRepoInfo(repoPath, remoteName)
	if err != nil {
		return nil, err
	}
	o.repo = repo

	// Build S3 prefixes based on repository name
	b := newPrefixBuilder(o.repo.Hash)
	o.prefixFiles = b.build("files.tar.gz")
	o.prefixPatterns = b.build("patterns.gz")

	return o, nil
}

// Repo returns the repository information.
func (o *Operator) Repo() *RepoInfo {
	return o.repo
}

// SetDryrun sets the dry run mode for the repository.
func (o *Operator) SetDryrun(b bool) {
	o.dryrun = b
}

// SetOverwrite sets the overwrite mode for the repository.
func (o *Operator) SetOverwrite(b bool) {
	o.overwrite = b
}

// SetPatterns sets the target patterns for the repository.
func (o *Operator) SetPatterns(patterns []string) {
	o.repo.targetPatterns = getPatterns(patterns)
}
