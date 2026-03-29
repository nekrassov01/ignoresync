package health

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfntypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nekrassov01/ignoresync/manager"
	"github.com/nekrassov01/ignoresync/testutil"
)

func TestChecker_Check(t *testing.T) {
	type fields struct {
		cfn ICFN
		s3  IS3
		kms IKMS
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
				cfn: &mockChecker{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return &cloudformation.DescribeStacksOutput{
							Stacks: []cfntypes.Stack{
								{
									StackStatus: cfntypes.StackStatusCreateComplete,
								},
							},
						}, nil
					},
				},
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{}, nil
					},
					putObjectFunc: func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						return &s3.PutObjectOutput{}, nil
					},
					deleteObjectFunc: func(_ context.Context, _ *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
						return &s3.DeleteObjectOutput{}, nil
					},
				},
				kms: &mockChecker{
					encryptFunc: func(_ context.Context, _ *kms.EncryptInput, _ ...func(*kms.Options)) (*kms.EncryptOutput, error) {
						return &kms.EncryptOutput{}, nil
					},
					decryptFunc: func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
						return &kms.DecryptOutput{}, nil
					},
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Checker{
				cfn: test.fields.cfn,
				s3:  test.fields.s3,
				kms: test.fields.kms,
				w:   test.fields.w,
			}
			err := o.Check(test.args.ctx, test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func TestChecker_CheckState(t *testing.T) {
	type fields struct {
		s3  IS3
		kms IKMS
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
			name:   "success",
			fields: fields{},
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
			name:   "invalid master key",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(func(s *manager.State) {
					s.MasterKeys = nil
				}),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
		{
			name:   "invalid account",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(func(s *manager.State) {
					s.Account = ""
				}),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
		{
			name:   "invalid region",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(func(s *manager.State) {
					s.Region = ""
				}),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
		{
			name:   "invalid bucket",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(func(s *manager.State) {
					s.Bucket.ARN.Resource = ""
				}),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
		{
			name:   "invalid sse key",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(func(s *manager.State) {
					s.SSEKey = nil
				}),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
		{
			name:   "invalid cse key",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(func(s *manager.State) {
					s.CSEKey = nil
				}),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Checker{
				s3:  test.fields.s3,
				kms: test.fields.kms,
				cfn: test.fields.cfn,
				w:   test.fields.w,
			}
			got, err := o.CheckState(test.args.ctx, test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestChecker_CheckStack(t *testing.T) {
	type fields struct {
		s3  IS3
		kms IKMS
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
			name: "success",
			fields: fields{
				cfn: &mockChecker{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						opts := &cloudformation.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &cloudformation.DescribeStacksOutput{
							Stacks: []cfntypes.Stack{
								{
									StackStatus: cfntypes.StackStatusCreateComplete,
								},
							},
						}, nil
					},
				},
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
			name: "error",
			fields: fields{
				cfn: &mockChecker{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return nil, testutil.NewError()
					},
				},
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
		{
			name: "deleted",
			fields: fields{
				cfn: &mockChecker{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return &cloudformation.DescribeStacksOutput{
							Stacks: []cfntypes.Stack{
								{
									StackStatus: cfntypes.StackStatusDeleteComplete,
								},
							},
						}, nil
					},
				},
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
		{
			name: "not completed",
			fields: fields{
				cfn: &mockChecker{
					describeStacksFunc: func(_ context.Context, _ *cloudformation.DescribeStacksInput, _ ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
						return &cloudformation.DescribeStacksOutput{
							Stacks: []cfntypes.Stack{
								{
									StackStatus: cfntypes.StackStatusCreateInProgress,
								},
							},
						}, nil
					},
				},
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
			o := &Checker{
				s3:  test.fields.s3,
				kms: test.fields.kms,
				cfn: test.fields.cfn,
				w:   test.fields.w,
			}
			got, err := o.CheckStack(test.args.ctx, test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestChecker_CheckBucket(t *testing.T) {
	type fields struct {
		s3  IS3
		kms IKMS
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
			name: "success",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, optFns ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						opts := &s3.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func() func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						// Fail the first time, succeed the second time
						n := 0
						return func(_ context.Context, _ *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							opts := &s3.Options{}
							for _, optFn := range optFns {
								optFn(opts)
							}
							n++
							if n == 1 {
								return nil, testutil.NewError()
							}
							return &s3.PutObjectOutput{}, nil
						}
					}(),
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						opts := &s3.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("healthcheck")))}, nil
					},
					deleteObjectFunc: func(_ context.Context, _ *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
						opts := &s3.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &s3.DeleteObjectOutput{}, nil
					},
				},
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
			name: "error at heading bucket",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return nil, testutil.NewError()
					},
					putObjectFunc: func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						return &s3.PutObjectOutput{}, nil
					},
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{}, nil
					},
					deleteObjectFunc: nil,
				},
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
		{
			name: "error at putting object",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						return nil, testutil.NewError()
					},
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{}, nil
					},
					deleteObjectFunc: nil,
				},
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
		{
			name: "error at putting object because of sse key is nil",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func() func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						// Fail the first time, succeed the second time
						n := 0
						return func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							n++
							if n == 1 {
								return nil, testutil.NewError()
							}
							return &s3.PutObjectOutput{}, nil
						}
					}(),
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{}, nil
					},
					deleteObjectFunc: nil,
				},
			},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(
					func(s *manager.State) {
						s.SSEKey = nil
					},
				),
			},
			want: want{
				value:   false,
				isError: true,
			},
		},
		{
			name: "error at getting object",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func() func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						// Fail the first time, succeed the second time
						n := 0
						return func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							n++
							if n == 1 {
								return nil, testutil.NewError()
							}
							return &s3.PutObjectOutput{}, nil
						}
					}(),
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return nil, testutil.NewError()
					},
					deleteObjectFunc: nil,
				},
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
		{
			name: "error at deleting object",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func() func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						// Fail the first time, succeed the second time
						n := 0
						return func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							n++
							if n == 1 {
								return nil, testutil.NewError()
							}
							return &s3.PutObjectOutput{}, nil
						}
					}(),
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte("healthcheck")))}, nil
					},
					deleteObjectFunc: func(_ context.Context, _ *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
						return nil, testutil.NewError()
					},
				},
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
		{
			name: "error at reading object",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func() func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						// Fail the first time, succeed the second time
						n := 0
						return func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							n++
							if n == 1 {
								return nil, testutil.NewError()
							}
							return &s3.PutObjectOutput{}, nil
						}
					}(),
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body: testutil.ReadErrorReadCloser{},
						}, nil
					},
				},
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
		{
			name: "error at closing object",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func() func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						// Fail the first time, succeed the second time
						n := 0
						return func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							n++
							if n == 1 {
								return nil, testutil.NewError()
							}
							return &s3.PutObjectOutput{}, nil
						}
					}(),
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body: testutil.CloseErrorReadCloser{Reader: bytes.NewReader([]byte("healthcheck"))},
						}, nil
					},
				},
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
		{
			name: "content mismatch",
			fields: fields{
				s3: &mockChecker{
					headBucketFunc: func(_ context.Context, _ *s3.HeadBucketInput, _ ...func(*s3.Options)) (*s3.HeadBucketOutput, error) {
						return &s3.HeadBucketOutput{}, nil
					},
					putObjectFunc: func() func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						// Fail the first time, succeed the second time
						n := 0
						return func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
							n++
							if n == 1 {
								return nil, testutil.NewError()
							}
							return &s3.PutObjectOutput{}, nil
						}
					}(),
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body: io.NopCloser(bytes.NewReader([]byte("mismatched"))),
						}, nil
					},
				},
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
			o := &Checker{
				s3:  test.fields.s3,
				kms: test.fields.kms,
				cfn: test.fields.cfn,
				w:   test.fields.w,
			}
			got, err := o.CheckBucket(test.args.ctx, test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestChecker_checkKey(t *testing.T) {
	type fields struct {
		s3  IS3
		kms IKMS
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
			name: "success",
			fields: fields{
				kms: &mockChecker{
					encryptFunc: func(_ context.Context, _ *kms.EncryptInput, optFns ...func(*kms.Options)) (*kms.EncryptOutput, error) {
						opts := &kms.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &kms.EncryptOutput{CiphertextBlob: []byte("encrypted")}, nil
					},
					decryptFunc: func() func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
						n := 0
						return func(_ context.Context, _ *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
							opts := &kms.Options{}
							for _, optFn := range optFns {
								optFn(opts)
							}
							n++
							switch n {
							case 1: // step1: decrypt OK
								return &kms.DecryptOutput{Plaintext: []byte("healthcheck")}, nil
							case 2: // step2: invalid context
								return nil, testutil.NewError()
							case 3: // step3: tampered data
								return nil, testutil.NewError()
							default:
								t.Fatalf("count exceeded: n=%d", n)
								return nil, testutil.NewError()
							}
						}
					}(),
				},
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
			name: "error at encrypt",
			fields: fields{
				kms: &mockChecker{
					encryptFunc: func(_ context.Context, _ *kms.EncryptInput, _ ...func(*kms.Options)) (*kms.EncryptOutput, error) {
						return nil, testutil.NewError()
					},
					decryptFunc: nil,
				},
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
		{
			name: "error at decrypt (1st call)",
			fields: fields{
				kms: &mockChecker{
					encryptFunc: func(_ context.Context, _ *kms.EncryptInput, _ ...func(*kms.Options)) (*kms.EncryptOutput, error) {
						return &kms.EncryptOutput{CiphertextBlob: []byte("encrypted")}, nil
					},
					decryptFunc: func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
						return nil, testutil.NewError()
					},
				},
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
		{
			name: "security flaw: decrypted with invalid context",
			fields: fields{
				kms: &mockChecker{
					encryptFunc: func(_ context.Context, _ *kms.EncryptInput, _ ...func(*kms.Options)) (*kms.EncryptOutput, error) {
						return &kms.EncryptOutput{
							CiphertextBlob: []byte("encrypted"),
						}, nil
					},
					decryptFunc: func() func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
						n := 0
						return func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
							n++
							switch n {
							case 1: // case1: decrypt OK
								return &kms.DecryptOutput{
									Plaintext: []byte("healthcheck"),
								}, nil
							case 2: // case2: invalid context - SHOULD FAIL but succeeds
								return &kms.DecryptOutput{
									Plaintext: []byte("healthcheck"),
								}, nil
							case 3: // case3: tampered data
								return nil, testutil.NewError()
							default:
								t.Fatalf("count exceeded: n=%d", n)
								return nil, testutil.NewError()
							}
						}
					}(),
				},
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
		{
			name: "security flaw: decrypted tampered ciphertext",
			fields: fields{
				kms: &mockChecker{
					encryptFunc: func(_ context.Context, _ *kms.EncryptInput, _ ...func(*kms.Options)) (*kms.EncryptOutput, error) {
						return &kms.EncryptOutput{
							CiphertextBlob: []byte("encrypted"),
						}, nil
					},
					decryptFunc: func() func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
						n := 0
						return func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
							n++
							switch n {
							case 1: // case1: decrypt OK
								return &kms.DecryptOutput{
									Plaintext: []byte("healthcheck"),
								}, nil
							case 2: // case2: invalid context
								return nil, testutil.NewError()
							case 3: // case3: tampered data - SHOULD FAIL but succeeds
								return &kms.DecryptOutput{
									Plaintext: []byte("healthcheck"),
								}, nil
							default:
								t.Fatalf("count exceeded: n=%d", n)
								return nil, testutil.NewError()
							}
						}
					}(),
				},
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
			o := &Checker{
				s3:  test.fields.s3,
				kms: test.fields.kms,
				cfn: test.fields.cfn,
				w:   test.fields.w,
			}
			got, err := o.checkKey(test.args.ctx, test.args.state.SSEKey, test.args.state.Region)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
