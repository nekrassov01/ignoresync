package manager

import (
	"encoding/hex"
	"testing"

	"github.com/nekrassov01/ignoresync/params"
	"github.com/nekrassov01/ignoresync/testutil"
)

func TestGenerateCredential(t *testing.T) {
	type want struct {
		isError bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "success",
			want: want{
				isError: false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, key, err := GenerateCredential()
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckHex(t, id, hex.EncodedLen(params.KeyIDSize))
			testutil.CheckHex(t, hex.EncodeToString(key), hex.EncodedLen(params.MasterKeySize))
		})
	}
}

func TestEncodeCredential(t *testing.T) {
	type args struct {
		id  string
		key []byte
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
			name: "success",
			args: args{
				id:  "11111111aaaaaaaa",
				key: []byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53},
			},
			want: want{
				value: "11111111aaaaaaaa30313233343536373839303132333435",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := EncodeCredential(test.args.id, test.args.key)
			testutil.CheckValue(t, got, test.want.value)
		})
	}
}

func Test_DecodeCredential(t *testing.T) {
	type args struct {
		cred string
	}
	type want struct {
		id      string
		key     []byte
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				cred: "11111111aaaaaaaa30313233343536373839303132333435",
			},
			want: want{
				id:      "11111111aaaaaaaa",
				key:     []byte{48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53},
				isError: false,
			},
		},
		{
			name: "size too short",
			args: args{
				cred: "11111111aaaaaaaa3031323334353637383930313233343",
			},
			want: want{
				id:      "",
				key:     nil,
				isError: true,
			},
		},
		{
			name: "size too long",
			args: args{
				cred: "11111111aaaaaaaa303132333435363738393031323334350",
			},
			want: want{
				id:      "",
				key:     nil,
				isError: true,
			},
		},
		{
			name: "invalid characters",
			args: args{
				cred: "11111111aaaaaaaa3031323334353637383930313233343g",
			},
			want: want{
				id:      "",
				key:     nil,
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, key, err := DecodeCredential(test.args.cred)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, id, test.want.id)
			testutil.CheckValue(t, key, test.want.key)
		})
	}
}

func Test_ValidateCredential(t *testing.T) {
	type args struct {
		cred string
	}
	type want struct {
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				cred: "11111111aaaaaaaa30313233343536373839303132333435",
			},
			want: want{
				isError: false,
			},
		},
		{
			name: "size too short",
			args: args{
				cred: "11111111aaaaaaaa3031323334353637383930313233343",
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "size too long",
			args: args{
				cred: "11111111aaaaaaaa303132333435363738393031323334350",
			},
			want: want{
				isError: true,
			},
		},
		{
			name: "invalid characters",
			args: args{
				cred: "11111111aaaaaaaa3031323334353637383930313233343g",
			},
			want: want{
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateCredential(test.args.cred)
			testutil.CheckError(t, err != nil, test.want.isError)
		})
	}
}
