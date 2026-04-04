package ignoresync

import (
	"testing"

	"github.com/nekrassov01/ignoresync/testutil"
)

func TestFormatError(t *testing.T) {
	type args struct {
		pkg  string
		kind string
		err  error
	}
	type want struct {
		value string
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
				value: "pkg: kind: error",
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
				value: "pkg: kind",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := FormatError(test.args.pkg, test.args.kind, test.args.err)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
