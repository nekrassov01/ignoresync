package env

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

var _ ICFN = (*mockDeployer)(nil)

// mockDeployer is a mock implementation of the ICFN interface for testing.
type mockDeployer struct {
	createStackFunc                  func(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	describeStacksFunc               func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	newStackCreateCompleteWaiterFunc func(optFns ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter
	describeStackEventsFunc          func(ctx context.Context, params *cloudformation.DescribeStackEventsInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error)
	w                                io.Writer
}

// CreateStack calls the mocked CreateStack function.
func (o *mockDeployer) CreateStack(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	return o.createStackFunc(ctx, params, optFns...)
}

// DescribeStacks calls the mocked DescribeStacks function.
func (o *mockDeployer) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return o.describeStacksFunc(ctx, params, optFns...)
}

// NewStackCreateCompleteWaiter creates a new StackCreateCompleteWaiter.
func (o *mockDeployer) NewStackCreateCompleteWaiter(optFns ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
	return o.newStackCreateCompleteWaiterFunc(optFns...)
}

// DescribeStackEvents calls the mocked DescribeStackEvents function.
func (o *mockDeployer) DescribeStackEvents(ctx context.Context, params *cloudformation.DescribeStackEventsInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
	return o.describeStackEventsFunc(ctx, params, optFns...)
}

// mockStackCreateCompleteWaiter is a mock implementation of the StackCreateCompleteWaiter for testing.
type mockStackCreateCompleteWaiter struct {
	describeStacksFunc func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
}

// Wait calls the mocked Wait function.
func (o *mockStackCreateCompleteWaiter) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return o.describeStacksFunc(ctx, params, optFns...)
}
