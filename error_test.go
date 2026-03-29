package ignoresync

import (
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

func Test_baseError_Error(t *testing.T) {
	type fields struct {
		pkg  string
		kind string
		err  error
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
				pkg:  "pkg",
				kind: "kind",
				err:  testutil.NewError(),
			},
			want: want{
				value: "pkg: kind: error",
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				pkg:  "pkg",
				kind: "kind",
				err:  nil,
			},
			want: want{
				value: "pkg: kind",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := &baseError{
				pkg:  test.fields.pkg,
				kind: test.fields.kind,
				err:  test.fields.err,
			}
			got := e.Error()
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func Test_baseError_Unwrap(t *testing.T) {
	type fields struct {
		pkg  string
		kind string
		err  error
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
				pkg:  "pkg",
				kind: "kind",
				err:  testutil.NewError(),
			},
			want: want{
				err: testutil.NewError(),
			},
		},
		{
			name: "unwrapped error",
			fields: fields{
				pkg:  "pkg",
				kind: "kind",
				err:  nil,
			},
			want: want{
				err: nil,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := &baseError{
				kind: test.fields.kind,
				pkg:  test.fields.pkg,
				err:  test.fields.err,
			}
			got := e.Unwrap()
			testutil.CheckValue(t, got, test.want.err)
		})
	}
}

func TestNewError(t *testing.T) {
	type args struct {
		pkg  string
		kind string
		err  error
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
				pkg:  "pkg",
				kind: "kind",
				err:  testutil.NewError(),
			},
			want: want{
				err: &baseError{
					pkg:  "pkg",
					kind: "kind",
					err:  testutil.NewError(),
				},
			},
		},
		{
			name: "unwrapped error",
			args: args{
				pkg:  "pkg",
				kind: "kind",
				err:  nil,
			},
			want: want{
				err: &baseError{
					pkg:  "pkg",
					kind: "kind",
					err:  nil,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := NewError(test.args.pkg, test.args.kind, test.args.err)
			testutil.CheckValue(t, err, test.want.err)
		})
	}
}
