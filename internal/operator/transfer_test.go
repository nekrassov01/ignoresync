package operator

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"io"
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/nekrassov01/ignoresync/internal/manager"
	"github.com/nekrassov01/ignoresync/internal/testutil"
)

func TestOperator_upload(t *testing.T) {
	var (
		ctx    = context.Background()
		state  = manager.NewMockState()
		prefix = "prefix"
		body   = bytes.NewReader([]byte("body"))
		etag   = new("etag")
	)
	type fields struct {
		s3 IS3
	}
	type args struct {
		ctx      context.Context
		state    *manager.State
		body     io.Reader
		prefix   string
		metadata map[string]string
		etag     *string
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
				s3: &mockOperator{
					newTransferManagerFunc: func() *transfermanager.Client {
						return transfermanager.New(
							&mockOperator{
								putObjectFunc: func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
									return &s3.PutObjectOutput{}, nil
								},
							},
						)
					},
					newObjectExistsWaiterFunc: func() *s3.ObjectExistsWaiter {
						return s3.NewObjectExistsWaiter(
							&mockOperator{
								headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
									return &s3.HeadObjectOutput{}, nil
								},
							},
						)
					},
				},
			},
			args: args{
				ctx:      ctx,
				state:    state,
				body:     body,
				metadata: map[string]string{},
				prefix:   prefix,
				etag:     etag,
			},
			want: want{
				isError: false,
			},
		},
		{
			name: "error at upload",
			fields: fields{
				s3: &mockOperator{
					newTransferManagerFunc: func() *transfermanager.Client {
						return transfermanager.New(
							&mockOperator{
								putObjectFunc: func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
									return nil, testutil.NewError()
								},
							},
						)
					},
					newObjectExistsWaiterFunc: func() *s3.ObjectExistsWaiter {
						return s3.NewObjectExistsWaiter(
							&mockOperator{
								headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
									return &s3.HeadObjectOutput{}, nil
								},
							},
						)
					},
				},
			},
			args: args{
				ctx:      ctx,
				state:    state,
				body:     body,
				metadata: map[string]string{},
				prefix:   prefix,
				etag:     etag,
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "error at wait upload",
			fields: fields{
				s3: &mockOperator{
					newTransferManagerFunc: func() *transfermanager.Client {
						return transfermanager.New(
							&mockOperator{
								putObjectFunc: func(_ context.Context, _ *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
									return &s3.PutObjectOutput{}, nil
								},
							},
						)
					},
					newObjectExistsWaiterFunc: func() *s3.ObjectExistsWaiter {
						return s3.NewObjectExistsWaiter(
							&mockOperator{
								headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
									return nil, testutil.NewError()
								},
							},
						)
					},
				},
			},
			args: args{
				ctx:      ctx,
				state:    state,
				body:     body,
				metadata: map[string]string{},
				prefix:   prefix,
				etag:     etag,
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				s3: test.fields.s3,
			}
			err := o.upload(test.args.ctx, test.args.state, test.args.body, test.args.prefix, test.args.metadata, test.args.etag)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func TestOperator_download(t *testing.T) {
	var (
		ctx    = context.Background()
		state  = manager.NewMockState()
		prefix = "prefix"
	)
	type fields struct {
		s3 IS3
	}
	type args struct {
		ctx    context.Context
		state  *manager.State
		prefix string
	}
	type want struct {
		body     []byte
		metadata *metadata
		etag     *string
		isError  bool
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
				s3: &mockOperator{
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body:     io.NopCloser(bytes.NewReader([]byte("body"))),
							Metadata: testMetadataInput,
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				body:     []byte("body"),
				metadata: testMetadataOutput,
				etag:     nil,
				isError:  false,
			},
		},
		{
			name: "error at download",
			fields: fields{
				s3: &mockOperator{
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return nil, testutil.NewError()
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				body:     nil,
				metadata: nil,
				etag:     nil,
				isError:  true,
			},
		},
		{
			name: "missing wrapped data key",
			fields: fields{
				s3: &mockOperator{
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body: io.NopCloser(bytes.NewReader([]byte("body"))),
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								delete(m, metadataKeyDataKey)
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				body:     nil,
				metadata: nil,
				etag:     nil,
				isError:  true,
			},
		},
		{
			name: "invalid wrapped data key",
			fields: fields{
				s3: &mockOperator{
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body: io.NopCloser(bytes.NewReader([]byte("body"))),
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								m[metadataKeyDataKey] = "invalid"
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				body:     nil,
				metadata: nil,
				etag:     nil,
				isError:  true,
			},
		},
		{
			name: "missing base nonce",
			fields: fields{
				s3: &mockOperator{
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body: io.NopCloser(bytes.NewReader([]byte("body"))),
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								delete(m, metadataKeyBaseNonce)
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				body:     nil,
				metadata: nil,
				etag:     nil,
				isError:  true,
			},
		},
		{
			name: "invalid base nonce",
			fields: fields{
				s3: &mockOperator{
					getObjectFunc: func(_ context.Context, _ *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
						return &s3.GetObjectOutput{
							Body: io.NopCloser(bytes.NewReader([]byte("body"))),
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								m[metadataKeyBaseNonce] = "invalid"
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				body:     nil,
				metadata: nil,
				etag:     nil,
				isError:  true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				s3: test.fields.s3,
			}
			body, metadata, etag, err := o.download(test.args.ctx, test.args.state, test.args.prefix)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, testutil.ReadBody(t, body), test.want.body)
			testutil.CheckValue(t, metadata, test.want.metadata)
			testutil.CheckValue(t, etag, test.want.etag)
		})
	}
}

func TestOperator_head(t *testing.T) {
	var (
		ctx    = context.Background()
		state  = manager.NewMockState()
		prefix = "prefix"
	)
	type fields struct {
		s3 IS3
	}
	type args struct {
		ctx    context.Context
		state  *manager.State
		prefix string
	}
	type want struct {
		metadata *metadata
		etag     *string
		isError  bool
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
				s3: &mockOperator{
					headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
						return &s3.HeadObjectOutput{
							Metadata: testMetadataInput,
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				metadata: testMetadataOutput,
				isError:  false,
			},
		},
		{
			name: "error at download",
			fields: fields{
				s3: &mockOperator{
					headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
						return nil, testutil.NewError()
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing wrapped data key",
			fields: fields{
				s3: &mockOperator{
					headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
						return &s3.HeadObjectOutput{
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								delete(m, metadataKeyDataKey)
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid wrapped data key",
			fields: fields{
				s3: &mockOperator{
					headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
						return &s3.HeadObjectOutput{
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								m[metadataKeyDataKey] = ""
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "missing base nonce",
			fields: fields{
				s3: &mockOperator{
					headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
						return &s3.HeadObjectOutput{
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								delete(m, metadataKeyBaseNonce)
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
		{
			name: "invalid base nonce",
			fields: fields{
				s3: &mockOperator{
					headObjectFunc: func(_ context.Context, _ *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
						return &s3.HeadObjectOutput{
							Metadata: func() map[string]string {
								m := make(map[string]string, len(testMetadataInput))
								maps.Copy(m, testMetadataInput)
								m[metadataKeyDataKey] = base64.StdEncoding.EncodeToString([]byte("wdk"))
								m[metadataKeyBaseNonce] = "%%%"
								return m
							}(),
						}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				metadata: nil,
				isError:  true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				s3: test.fields.s3,
			}
			metadata, etag, err := o.head(test.args.ctx, test.args.state, test.args.prefix)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, metadata, test.want.metadata)
			testutil.CheckValue(t, etag, test.want.etag)
		})
	}
}

func TestOperator_delete(t *testing.T) {
	var (
		ctx    = context.Background()
		state  = manager.NewMockState()
		prefix = "files/prefix"
	)
	type fields struct {
		s3 IS3
	}
	type args struct {
		ctx    context.Context
		state  *manager.State
		prefix string
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
				s3: &mockOperator{
					deleteObjectFunc: func(_ context.Context, _ *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
						return &s3.DeleteObjectOutput{}, nil
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				isError: false,
			},
		},
		{
			name: "error",
			fields: fields{
				s3: &mockOperator{
					deleteObjectFunc: func(_ context.Context, _ *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
						return nil, testutil.NewError()
					},
				},
			},
			args: args{
				ctx:    ctx,
				state:  state,
				prefix: prefix,
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				s3: test.fields.s3,
			}
			err := o.delete(test.args.ctx, test.args.state, test.args.prefix)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}

func Test_isObjectNotFound(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		value bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "nil",
			args: args{
				err: nil,
			},
			want: want{
				value: false,
			},
		},
		{
			name: "not found",
			args: args{
				err: &smithy.GenericAPIError{
					Code: "NotFound",
				},
			},
			want: want{
				value: true,
			},
		},
		{
			name: "no such key",
			args: args{
				err: &smithy.GenericAPIError{
					Code: "NoSuchKey",
				},
			},
			want: want{
				value: true,
			},
		},
		{
			name: "other api error",
			args: args{
				err: &smithy.GenericAPIError{
					Code: "AccessDenied",
				},
			},
			want: want{
				value: false,
			},
		},
		{
			name: "non api error",
			args: args{
				err: errors.New("other"),
			},
			want: want{
				value: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isObjectNotFound(test.args.err)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func Test_isPreconditionFailed(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		value bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "nil",
			args: args{
				err: nil,
			},
			want: want{
				value: false,
			},
		},
		{
			name: "precondition failed",
			args: args{
				err: &smithy.GenericAPIError{
					Code: "PreconditionFailed",
				},
			},
			want: want{
				value: true,
			},
		},
		{
			name: "conditional request conflict",
			args: args{
				err: &smithy.GenericAPIError{
					Code: "ConditionalRequestConflict",
				},
			},
			want: want{
				value: true,
			},
		},
		{
			name: "other api error",
			args: args{
				err: &smithy.GenericAPIError{
					Code: "AccessDenied",
				},
			},
			want: want{
				value: false,
			},
		},
		{
			name: "non api error",
			args: args{
				err: errors.New("other"),
			},
			want: want{
				value: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isConditionalError(test.args.err)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
