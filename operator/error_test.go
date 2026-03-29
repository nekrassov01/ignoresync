package operator

import (
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

func TestErrorKind_String(t *testing.T) {
	type want struct {
		value string
	}
	tests := []struct {
		name string
		kind ErrorKind
		want want
	}{
		{
			name: "none",
			kind: ErrorKindNone,
			want: want{
				value: "no error",
			},
		},
		{
			name: "repository",
			kind: ErrorKindRepository,
			want: want{
				value: "repository error",
			},
		},
		{
			name: "generation",
			kind: ErrorKindGenerate,
			want: want{
				value: "generation error",
			},
		},
		{
			name: "derivation",
			kind: ErrorKindDerive,
			want: want{
				value: "derivation error",
			},
		},
		{
			name: "validate",
			kind: ErrorKindValidate,
			want: want{
				value: "validation error",
			},
		},
		{
			name: "archive",
			kind: ErrorKindArchive,
			want: want{
				value: "archive error",
			},
		},
		{
			name: "encrypt",
			kind: ErrorKindEncrypt,
			want: want{
				value: "encrypt error",
			},
		},
		{
			name: "upload",
			kind: ErrorKindUpload,
			want: want{
				value: "upload error",
			},
		},
		{
			name: "download",
			kind: ErrorKindDownload,
			want: want{
				value: "download error",
			},
		},
		{
			name: "decrypt",
			kind: ErrorKindDecrypt,
			want: want{
				value: "decrypt error",
			},
		},
		{
			name: "restore",
			kind: ErrorKindRestore,
			want: want{
				value: "restore error",
			},
		},
		{
			name: "delete",
			kind: ErrorKindDelete,
			want: want{
				value: "deletion error",
			},
		},
		{
			name: "unknown",
			kind: ErrorKind(256),
			want: want{
				value: "unknown error",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.kind.String()
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestError_Error(t *testing.T) {
	type fields struct {
		Kind ErrorKind
		Err  error
	}
	type want struct {
		value string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "wrapped error",
			fields: fields{
				Kind: ErrorKindRepository,
				Err:  testutil.NewError(),
			},
			want: want{
				value: "operator: repository error: error",
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				Kind: ErrorKindRepository,
				Err:  nil,
			},
			want: want{
				value: "operator: repository error",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := &Error{
				Kind: test.fields.Kind,
				Err:  test.fields.Err,
			}
			got := e.Error()
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	type fields struct {
		Kind ErrorKind
		Err  error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "wrapped error",
			fields: fields{
				Kind: ErrorKindRepository,
				Err:  testutil.NewError(),
			},
			want: want{
				err: testutil.NewError(),
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				Kind: ErrorKindRepository,
				Err:  nil,
			},
			want: want{
				err: nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := &Error{
				Kind: test.fields.Kind,
				Err:  test.fields.Err,
			}
			err := e.Unwrap()
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewRepositoryError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindRepository,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindRepository,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewRepositoryError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewGenerateError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindGenerate,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindGenerate,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewGenerateError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewDeriveError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDerive,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDerive,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewDeriveError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewValidateError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindValidate,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindValidate,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewValidateError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewArchiveError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindArchive,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindArchive,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewArchiveError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewEncryptError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindEncrypt,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindEncrypt,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewEncryptError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewUploadError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindUpload,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindUpload,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewUploadError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewDownloadError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDownload,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDownload,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewDownloadError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewDecryptError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDecrypt,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDecrypt,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewDecryptError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewRestoreError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindRestore,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindRestore,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewRestoreError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewDeleteError(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "wrapped error",
			args: args{
				err: testutil.NewError(),
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDelete,
					Err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				err: nil,
			},
			want: want{
				err: &Error{
					Kind: ErrorKindDelete,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewDeleteError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}
