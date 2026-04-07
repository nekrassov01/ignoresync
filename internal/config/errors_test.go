package config

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
			name: "config",
			kind: ErrorKindConfig,
			want: want{
				value: "configuration error",
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
				Kind: ErrorKindConfig,
				Err:  testutil.NewError(),
			},
			want: want{
				value: "config: configuration error: error",
			},
		},
		{
			name: "unwrapped",
			fields: fields{
				Kind: ErrorKindConfig,
				Err:  nil,
			},
			want: want{
				value: "config: configuration error",
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
				Kind: ErrorKindConfig,
				Err:  testutil.NewError(),
			},
			want: want{
				err: testutil.NewError(),
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				Kind: ErrorKindConfig,
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
			got := e.Unwrap()
			testutil.CheckValue(t, got, test.want.err)
		})
	}
}

func TestNewConfigError(t *testing.T) {
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
					Kind: ErrorKindConfig,
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
					Kind: ErrorKindConfig,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewConfigError(test.args.err)
			testutil.CheckValue(t, got, test.want.err)
		})
	}
}
