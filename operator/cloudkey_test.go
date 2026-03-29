package operator

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/nekrassov01/ignoresync/manager"
	"github.com/nekrassov01/ignoresync/testutil"
)

func TestOperator_generateCloudKey(t *testing.T) {
	type fields struct {
		kms IKMS
	}
	type args struct {
		ctx   context.Context
		state *manager.State
	}
	type want struct {
		plaintext  []byte
		ciphertext []byte
		isError    bool
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
				kms: &mockOperator{
					generateDataKeyFunc: func(_ context.Context, _ *kms.GenerateDataKeyInput, optFns ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
						opts := &kms.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &kms.GenerateDataKeyOutput{
							Plaintext:      []byte("plaintext"),
							CiphertextBlob: []byte("ciphertext"),
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				state: manager.NewMockState(func(s *manager.State) {
					s.Region = "ap-northeast-1"
					s.CSEKey.ARN.Region = "ap-northeast-1"
					s.CSEKey.ARN.Resource = "alias/custom-cse"
					s.CSEKey.EncryptionContext = map[string]string{"bucket/key": "custom-context"}
				}),
			},
			want: want{
				plaintext:  []byte("plaintext"),
				ciphertext: []byte("ciphertext"),
				isError:    false,
			},
		},
		{
			name: "generate data key error",
			fields: fields{
				kms: &mockOperator{
					generateDataKeyFunc: func(_ context.Context, _ *kms.GenerateDataKeyInput, _ ...func(*kms.Options)) (*kms.GenerateDataKeyOutput, error) {
						return nil, testutil.NewError()
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				state: manager.NewMockState(),
			},
			want: want{
				plaintext:  nil,
				ciphertext: nil,
				isError:    true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				kms: test.fields.kms,
			}
			plaintext, ciphertext, err := o.generateCloudKey(test.args.ctx, test.args.state)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, plaintext, test.want.plaintext)
			testutil.CheckValue(t, ciphertext, test.want.ciphertext)
		})
	}
}

func TestOperator_decryptCloudKey(t *testing.T) {
	type fields struct {
		kms IKMS
	}
	type args struct {
		ctx        context.Context
		state      *manager.State
		ciphertext []byte
	}
	type want struct {
		value   []byte
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
				kms: &mockOperator{
					decryptFunc: func(_ context.Context, _ *kms.DecryptInput, optFns ...func(*kms.Options)) (*kms.DecryptOutput, error) {
						opts := &kms.Options{}
						for _, optFn := range optFns {
							optFn(opts)
						}
						return &kms.DecryptOutput{Plaintext: []byte("plaintext")}, nil
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				ciphertext: []byte("ciphertext"),
				state: manager.NewMockState(func(s *manager.State) {
					s.Region = "eu-west-1"
					s.CSEKey.EncryptionContext = map[string]string{"bucket/key": "custom-context"}
				}),
			},
			want: want{
				value:   []byte("plaintext"),
				isError: false,
			},
		},
		{
			name: "decrypt error",
			fields: fields{
				kms: &mockOperator{
					decryptFunc: func(_ context.Context, _ *kms.DecryptInput, _ ...func(*kms.Options)) (*kms.DecryptOutput, error) {
						return nil, testutil.NewError()
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				state:      manager.NewMockState(),
				ciphertext: []byte("ciphertext"),
			},
			want: want{
				value:   nil,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				kms: test.fields.kms,
			}
			got, err := o.decryptCloudKey(test.args.ctx, test.args.state, test.args.ciphertext)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
