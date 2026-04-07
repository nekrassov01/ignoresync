package manager

import (
	"testing"

	"github.com/nekrassov01/ignoresync/internal/testutil"
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
			name: "auth",
			kind: ErrorKindAuth,
			want: want{
				value: "authentication error",
			},
		},
		{
			name: "state",
			kind: ErrorKindState,
			want: want{
				value: "state error",
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
				Kind: ErrorKindAuth,
				Err:  testutil.NewError(),
			},
			want: want{
				value: "manager: authentication error: error",
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				Kind: ErrorKindAuth,
				Err:  nil,
			},
			want: want{
				value: "manager: authentication error",
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
				Kind: ErrorKindAuth,
				Err:  testutil.NewError(),
			},
			want: want{
				err: testutil.NewError(),
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				Kind: ErrorKindAuth,
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

func TestNewAuthError(t *testing.T) {
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
					Kind: ErrorKindAuth,
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
					Kind: ErrorKindAuth,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewAuthError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewCredentialError(t *testing.T) {
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
					Kind: ErrorKindCredential,
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
					Kind: ErrorKindCredential,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewCredentialError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}

func TestNewStateError(t *testing.T) {
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
					Kind: ErrorKindState,
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
					Kind: ErrorKindState,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewStateError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}
