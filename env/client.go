package env

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

var _ ICFN = (*CFN)(nil)

// ICFN defines the interface for CloudFormation client.
type ICFN interface {
	CreateStack(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	NewStackCreateCompleteWaiter(opts ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter
	DescribeStackEvents(ctx context.Context, params *cloudformation.DescribeStackEventsInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error)
}

// CFN defines a struct for CloudFormation client.
type CFN struct {
	*cloudformation.Client
}

// Deployer is a client for deploying environments.
type Deployer struct {
	cfn ICFN
	w   io.Writer
}

// NewStackCreateCompleteWaiter creates a new StackCreateCompleteWaiter.
func (o *CFN) NewStackCreateCompleteWaiter(opts ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
	return cloudformation.NewStackCreateCompleteWaiter(o.Client, opts...)
}

// New creates a new Deployer.
func New(w io.Writer, cfg aws.Config) *Deployer {
	return &Deployer{
		cfn: &CFN{Client: cloudformation.NewFromConfig(cfg)},
		w:   w,
	}
}
