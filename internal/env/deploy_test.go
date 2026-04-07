package env

import (
	"bytes"
	"context"
	_ "embed"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
	"github.com/nekrassov01/ignoresync/internal/manager"
	"github.com/nekrassov01/ignoresync/internal/testutil"
)

func TestDeployer_Deploy(t *testing.T) {
	type fields struct {
		cfn ICFN
		w   io.Writer
	}
	type args struct {
		ctx   context.Context
		state *manager.State
	}
	type want struct {
		isError bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "success",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						opts := &cloudformation.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(optFns ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						outer := &cloudformation.StackCreateCompleteWaiterOptions{
							ClientOptions: nil,
						}
						for _, optFn := range optFns {
							optFn(outer)
						}
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									inner := &cloudformation.Options{}
									for _, opt := range outer.ClientOptions {
										opt(inner)
									}
									return &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackStatus: types.StackStatusCreateComplete,
											},
										},
									}, nil
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: nil,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: false,
			},
		},
		{
			name: "error at creating stack",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return nil, testutil.NewError()
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackStatus: types.StackStatusCreateComplete,
											},
										},
									}, nil
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: nil,
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return error",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return nil, testutil.NewError()
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						opts := &cloudformation.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &cloudformation.DescribeStackEventsOutput{}, nil
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return failure status",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackStatus: types.StackStatusCreateFailed,
											},
										},
									}, nil
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						return &cloudformation.DescribeStackEventsOutput{}, nil
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return error and failed to get reason",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return nil, testutil.NewError()
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						return nil, testutil.NewError()
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return failure status and failed to get reason",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackStatus: types.StackStatusCreateFailed,
											},
										},
									}, nil
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						return nil, testutil.NewError()
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return error and get reason",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return nil, testutil.NewError()
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						return &cloudformation.DescribeStackEventsOutput{
							StackEvents: []types.StackEvent{
								{
									ResourceStatus:       types.ResourceStatusCreateFailed,
									ResourceStatusReason: aws.String("reason"),
								},
							},
						}, nil
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return failure status and get reason",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackStatus: types.StackStatusCreateFailed,
											},
										},
									}, nil
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						return &cloudformation.DescribeStackEventsOutput{
							StackEvents: []types.StackEvent{
								{
									ResourceStatus:       types.ResourceStatusCreateFailed,
									ResourceStatusReason: aws.String("reason"),
								},
							},
						}, nil
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return error and get empty reason",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return nil, testutil.NewError()
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						return &cloudformation.DescribeStackEventsOutput{
							StackEvents: []types.StackEvent{
								{
									ResourceStatus:       types.ResourceStatusCreateFailed,
									ResourceStatusReason: nil,
								},
							},
						}, nil
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at waiting for stack creation: waiter return failure status and get empty reason",
			fields: fields{
				cfn: &mockDeployer{
					createStackFunc: func(_ context.Context, _ *cloudformation.CreateStackInput, _ ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
						return &cloudformation.CreateStackOutput{}, nil
					},
					newStackCreateCompleteWaiterFunc: func(_ ...func(*cloudformation.StackCreateCompleteWaiterOptions)) *cloudformation.StackCreateCompleteWaiter {
						return cloudformation.NewStackCreateCompleteWaiter(
							&mockStackCreateCompleteWaiter{
								describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
									return &cloudformation.DescribeStacksOutput{
										Stacks: []types.Stack{
											{
												StackStatus: types.StackStatusCreateFailed,
											},
										},
									}, nil
								},
							},
						)
					},
					// Called only if failure occurs while waiting
					describeStackEventsFunc: func(_ context.Context, _ *cloudformation.DescribeStackEventsInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStackEventsOutput, error) {
						return &cloudformation.DescribeStackEventsOutput{
							StackEvents: []types.StackEvent{
								{
									ResourceStatus:       types.ResourceStatusCreateFailed,
									ResourceStatusReason: nil,
								},
							},
						}, nil
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Deployer{
				cfn: test.fields.cfn,
				w:   test.fields.w,
			}
			err := o.Deploy(test.args.ctx, test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func TestDeployer_CheckDeployed(t *testing.T) {
	type fields struct {
		cfn ICFN
		w   io.Writer
	}
	type args struct {
		ctx   context.Context
		state *manager.State
	}
	type want struct {
		value   bool
		isError bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "deployed",
			fields: fields{
				cfn: &mockDeployer{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return &cloudformation.DescribeStacksOutput{
							Stacks: []types.Stack{
								{
									StackStatus: types.StackStatusCreateComplete,
								},
							},
						}, nil
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				value:   true,
				isError: false,
			},
		},
		{
			name: "not deployed",
			fields: fields{
				cfn: &mockDeployer{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return nil, &types.StackNotFoundException{
							Message: aws.String("does not exist"),
						}
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				value:   false,
				isError: false,
			},
		},
		{
			name: "not deployed by code only",
			fields: fields{
				cfn: &mockDeployer{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return nil, &smithy.GenericAPIError{
							Code:    "StackNotFoundException",
							Message: "other message",
						}
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				value:   false,
				isError: false,
			},
		},
		{
			name: "not deployed by message only",
			fields: fields{
				cfn: &mockDeployer{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return nil, &smithy.GenericAPIError{
							Code:    "ValidationError",
							Message: "stack does not exist",
						}
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				value:   false,
				isError: false,
			},
		},
		{
			name: "error",
			fields: fields{
				cfn: &mockDeployer{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return nil, testutil.NewError()
					},
				},
				w: &bytes.Buffer{},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Deployer{
				cfn: test.fields.cfn,
				w:   test.fields.w,
			}
			got, err := o.CheckDeployed(test.args.ctx, test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
