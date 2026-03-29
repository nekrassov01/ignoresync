package operator

import (
	"bytes"
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

func Test_newPrefixBuilder(t *testing.T) {
	type args struct {
		repoHash string
	}
	type want struct {
		value *prefixBuilder
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{repoHash: "repo"},
			want: want{
				value: &prefixBuilder{
					buf:      &bytes.Buffer{},
					repoHash: "repo",
				},
			},
		},
		{
			name: "empty repo name",
			args: args{repoHash: ""},
			want: want{
				value: &prefixBuilder{
					buf:      &bytes.Buffer{},
					repoHash: "",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := newPrefixBuilder(test.args.repoHash)
			testutil.CheckValue(t, got, test.want.value)
			testutil.CheckValue(t, got.buf == nil, false)
		})
	}
}

func Test_prefixBuilder_build(t *testing.T) {
	type fields struct {
		buf      *bytes.Buffer
		repoHash string
	}
	type args struct {
		name string
	}
	type want struct {
		value string
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
				buf:      &bytes.Buffer{},
				repoHash: "repo",
			},
			args: args{
				name: "file.txt",
			},
			want: want{
				value: "repo/file.txt",
			},
		},
		{
			name: "empty name",
			fields: fields{
				buf:      &bytes.Buffer{},
				repoHash: "repo",
			},
			args: args{
				name: "",
			},
			want: want{
				value: "repo/",
			},
		},
		{
			name: "empty repo name",
			fields: fields{
				buf:      &bytes.Buffer{},
				repoHash: "",
			},
			args: args{
				name: "file.txt",
			},
			want: want{
				value: "/file.txt",
			},
		},
		{
			name: "reset existing buffer",
			fields: fields{
				buf:      bytes.NewBufferString("stale-content"),
				repoHash: "repo",
			},
			args: args{
				name: "next.bin",
			},
			want: want{
				value: "repo/next.bin",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &prefixBuilder{
				buf:      test.fields.buf,
				repoHash: test.fields.repoHash,
			}
			got := o.build(test.args.name)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
