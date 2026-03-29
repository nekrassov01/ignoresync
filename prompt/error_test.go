package prompt

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
			name: "state",
			kind: ErrorKindPrompt,
			want: want{
				value: "interaction error",
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
				Kind: ErrorKindPrompt,
				Err:  testutil.NewError(),
			},
			want: want{
				value: "prompt: interaction error: error",
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				Kind: ErrorKindPrompt,
				Err:  nil,
			},
			want: want{
				value: "prompt: interaction error",
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
				Kind: ErrorKindPrompt,
				Err:  testutil.NewError(),
			},
			want: want{
				err: testutil.NewError(),
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				Kind: ErrorKindPrompt,
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

func TestNewPromptError(t *testing.T) {
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
					Kind: ErrorKindPrompt,
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
					Kind: ErrorKindPrompt,
					Err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewPromptError(test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}
