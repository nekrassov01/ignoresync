package config

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go"
	"github.com/nekrassov01/ignoresync/internal/testutil"
)

func TestNewRetryer(t *testing.T) {
	type args struct {
		err error
	}
	type want struct {
		maxAttempts int
		isRetryable bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "keeps standard retryables",
			args: args{
				err: &smithy.GenericAPIError{
					Code:    "RequestTimeout",
					Message: "request timed out",
				},
			},
			want: want{
				maxAttempts: maxRetryAttempts,
				isRetryable: true,
			},
		},
		{
			name: "adds slowdown retryable",
			args: args{
				err: &smithy.GenericAPIError{
					Code:    "SlowDown",
					Message: "slow down",
				},
			},
			want: want{
				maxAttempts: maxRetryAttempts,
				isRetryable: true,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := NewRetryer()
			testutil.CheckValue(t, got.MaxAttempts(), test.want.maxAttempts)
			testutil.CheckValue(t, got.IsErrorRetryable(test.args.err), test.want.isRetryable)
		})
	}
}

func Test_isErrorRetryer_IsErrorRetryable(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want aws.Ternary
	}{
		{
			name: "nil",
			args: args{},
			want: aws.UnknownTernary,
		},
		{
			name: "slowdown",
			args: args{
				err: &smithy.GenericAPIError{
					Code:    "SlowDown",
					Message: "slow down",
				},
			},
			want: aws.TrueTernary,
		},
		{
			name: "other api error",
			args: args{
				err: &smithy.GenericAPIError{
					Code:    "AccessDenied",
					Message: "denied",
				},
			},
			want: aws.UnknownTernary,
		},
		{
			name: "plain error",
			args: args{
				err: errors.New("api error SlowDown"),
			},
			want: aws.TrueTernary,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := &isErrorRetryer{}
			got := o.IsErrorRetryable(test.args.err)
			testutil.CheckValue(t, got, test.want)
		})
	}
}

func Test_isSlowDownError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{},
			want: false,
		},
		{
			name: "slowdown code",
			args: args{
				err: &smithy.GenericAPIError{
					Code:    "SlowDown",
					Message: "slow down",
				},
			},
			want: true,
		},
		{
			name: "other code",
			args: args{
				err: &smithy.GenericAPIError{
					Code:    "Throttling",
					Message: "throttled",
				},
			},
			want: false,
		},
		{
			name: "message only",
			args: args{
				err: errors.New("api error SlowDown: retry later"),
			},
			want: true,
		},
		{
			name: "other message",
			args: args{
				err: errors.New("api error AccessDenied: denied"),
			},
			want: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := isSlowDownError(test.args.err)
			testutil.CheckValue(t, got, test.want)
		})
	}
}
