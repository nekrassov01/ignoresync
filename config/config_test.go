package config

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/nekrassov01/ignoresync/testutil"
)

func TestLoadAWSConfig(t *testing.T) {
	type args struct {
		ctx     context.Context
		region  string
		profile string
	}
	type want struct {
		value   aws.Config
		isError bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "empty",
			args: args{
				ctx:     context.Background(),
				region:  "",
				profile: "",
			},
			want: want{
				value: aws.Config{
					Region: "us-east-1",
				},
				isError: false,
			},
		},
		{
			name: "region",
			args: args{
				ctx:     context.Background(),
				region:  "ap-northeast-1",
				profile: "",
			},
			want: want{
				value: aws.Config{
					Region: "ap-northeast-1",
				},
				isError: false,
			},
		},
		{
			name: "error",
			args: args{
				ctx:     context.Background(),
				region:  "",
				profile: "invalid-profile",
			},
			want: want{
				value:   aws.Config{},
				isError: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testutil.UnsetAWSProfile(t)
			got, err := LoadAWSConfig(test.args.ctx, test.args.region, test.args.profile)
			testutil.CheckError(t, err != nil, test.want.isError)
			testutil.CheckValue(t, got.Region, test.want.value.Region)
		})
	}
}
