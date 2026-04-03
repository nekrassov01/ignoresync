package prompt

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/nekrassov01/ignoresync/color"
	"github.com/nekrassov01/ignoresync/testutil"
)

func TestConfirm(t *testing.T) {
	type args struct {
		label string
		msg   string
	}
	type want struct {
		value   string
		output  string
		isError bool
	}
	tests := []struct {
		name  string
		stdin string
		args  args
		want  want
	}{
		{
			name:  "yes",
			stdin: "y\n",
			args: args{
				label: "overwrite",
				msg:   "skipped",
			},
			want: want{
				value:   "y",
				output:  "overwrite " + color.Mute("[y/N]") + " ",
				isError: false,
			},
		},
		{
			name:  "upper",
			stdin: "Y\n",
			args: args{
				label: "overwrite",
				msg:   "skipped",
			},
			want: want{
				value:   "Y",
				output:  "overwrite " + color.Mute("[y/N]") + " ",
				isError: false,
			},
		},
		{
			name:  "with spaces",
			stdin: "  y  \n",
			args: args{
				label: "overwrite",
				msg:   "skipped",
			},
			want: want{
				value:   "y",
				output:  "overwrite " + color.Mute("[y/N]") + " ",
				isError: false,
			},
		},
		{
			name:  "upper with spaces",
			stdin: "  Y  \n",
			args: args{
				label: "overwrite",
				msg:   "skipped",
			},
			want: want{
				value:   "Y",
				output:  "overwrite " + color.Mute("[y/N]") + " ",
				isError: false,
			},
		},
		{
			name:  "empty means reject",
			stdin: "\n",
			args: args{
				label: "overwrite",
				msg:   "skipped",
			},
			want: want{
				value:   "",
				output:  "overwrite " + color.Mute("[y/N]") + " ",
				isError: true,
			},
		},
		{
			name:  "n means reject",
			stdin: "n\n",
			args: args{
				label: "overwrite",
				msg:   "skipped",
			},
			want: want{
				value:   "n",
				output:  "overwrite " + color.Mute("[y/N]") + " ",
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testutil.SetStdin(t, testutil.CreateTemp(t, test.stdin))
			buf := &bytes.Buffer{}
			got, err := Confirm(buf, test.args.label, test.args.msg)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
			testutil.CheckValue(t, buf.String(), test.want.output)
		})
	}
}

func TestSecret(t *testing.T) {
	type args struct {
		label        string
		validateFunc func(string) error
	}
	type want struct {
		value   string
		output  string
		isError bool
	}
	tests := []struct {
		name  string
		stdin string
		args  args
		want  want
	}{
		{
			name:  "success without validation",
			stdin: "secret-value\n",
			args: args{
				label: "enter secret",
			},
			want: want{
				value:   "secret-value",
				output:  "enter secret \n",
				isError: false,
			},
		},
		{
			name:  "success with validation",
			stdin: "valid\n",
			args: args{
				label: "enter secret",
				validateFunc: func(input string) error {
					if input != "valid" {
						return errors.New("invalid")
					}
					return nil
				},
			},
			want: want{
				value:   "valid",
				output:  "enter secret \n",
				isError: false,
			},
		},
		{
			name:  "validation failure then success",
			stdin: "short\nvalid\n",
			args: args{
				label: "enter secret",
				validateFunc: func(input string) error {
					if input != "valid" {
						return errors.New("invalid")
					}
					return nil
				},
			},
			want: want{
				value:   "valid",
				output:  "enter secret " + color.Warn("invalid") + "\nenter secret \n",
				isError: false,
			},
		},
		{
			name:  "read error",
			stdin: "",
			args: args{
				label: "enter secret",
			},
			want: want{
				value:   "",
				output:  "enter secret ",
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testutil.SetStdin(t, testutil.CreateTemp(t, test.stdin))
			buf := &bytes.Buffer{}
			got, err := Secret(buf, test.args.label, test.args.validateFunc)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
			testutil.CheckValue(t, buf.String(), test.want.output)
		})
	}
}

func Test_readSecret(t *testing.T) {
	type args struct {
		f *os.File
	}
	type want struct {
		value   string
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "read line",
			args: args{
				f: testutil.CreateTemp(t, "secret\n"),
			},
			want: want{
				value:   "secret",
				isError: false,
			},
		},
		{
			name: "read line trims carriage return",
			args: args{
				f: testutil.CreateTemp(t, "secret\r\n"),
			},
			want: want{
				value:   "secret",
				isError: false,
			},
		},
		{
			name: "read line trims carriage return without newline",
			args: args{
				f: testutil.CreateTemp(t, ""),
			},
			want: want{
				value:   "",
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := readSecret(test.args.f, nil)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func Test_isTerminal(t *testing.T) {
	type args struct {
		f *os.File
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
			name: "regular file is not terminal",
			args: args{
				f: testutil.CreateTemp(t, "data\n"),
			},
			want: want{
				value: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isTerminal(test.args.f)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}
