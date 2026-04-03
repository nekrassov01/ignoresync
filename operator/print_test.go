package operator

import (
	"bytes"
	"io"
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

func TestOperator_printStatus(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	type args struct {
		msg    string
		path   string
		size   int64
		prefix string
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
			name: "with prefix and detail",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				msg:    "restored",
				path:   "file.txt",
				size:   10,
				prefix: "state:",
			},
			want: want{
				value: "state: restored -> file.txt (10 B)\n",
			},
		},
		{
			name: "without prefix and detail",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				msg:  "restored",
				path: "file.txt",
				size: -1,
			},
			want: want{
				value: "restored -> file.txt\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &Operator{
				w: test.fields.w,
			}
			o.printStatus(test.args.msg, test.args.path, test.args.size, test.args.prefix)
			testutil.CheckValue(t, test.fields.w.(*bytes.Buffer).String(), test.want.value)
		})
	}
}

func Test_sizeString(t *testing.T) {
	type args struct {
		size int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "negative size",
			args: args{
				size: -1,
			},
			want: "0 B",
		},
		{
			name: "positive size",
			args: args{
				size: 1024,
			},
			want: "1.0 kB",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sizeString(test.args.size)
			testutil.CheckValue(t, got, test.want)
		})
	}
}
