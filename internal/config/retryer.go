package config

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
)

var _ retry.IsErrorRetryable = (*isErrorRetryer)(nil)

// codes is the set of HTTP status codes that should be retried.
// 429 Too Many Requests not included in DefaultRetryableHTTPStatusCodes.
var codes = retry.RetryableHTTPStatusCode{
	Codes: map[int]struct{}{
		429: {},
	},
}

// NewRetryer creates a new retryer with the custom retry logic and backoff strategy.
// config.WithRetryer requires a function returning aws.Retryer, but retry.NewStandard
// returns a struct that implements aws.RetryerV2. The SDK internally type-asserts the
// aws.Retryer to aws.RetryerV2 and wraps it if necessary. Therefore, returning aws.Retryer
// here ensures compatibility with config.WithRetryer while still providing full RetryerV2
// functionality at runtime.
// see: https://github.com/aws/aws-sdk-go-v2/blob/main/aws/retry/retry.go#L79-L86
func NewRetryer() aws.Retryer {
	return retry.NewStandard(
		func(o *retry.StandardOptions) {
			o.MaxAttempts = maxRetryAttempts
			o.Retryables = append(o.Retryables, &isErrorRetryer{})
			o.Retryables = append(o.Retryables, codes)
			o.RateLimiter = ratelimit.None
		},
	)
}

// isErrorRetryer is a custom retryable error checker.
type isErrorRetryer struct{}

// IsErrorRetryable checks if the error is retryable based on its message.
func (o *isErrorRetryer) IsErrorRetryable(err error) aws.Ternary {
	if err == nil {
		return aws.UnknownTernary
	}
	if isSlowDownError(err) {
		return aws.TrueTernary
	}
	return aws.UnknownTernary
}

// isSlowDownError checks if the error is an AWS API SlowDown error.
func isSlowDownError(err error) bool {
	if err == nil {
		return false
	}
	// Although default retry settings include "SlowDown",
	// we will check the error messages directly to perform a more thorough evaluation.
	return strings.Contains(err.Error(), "api error SlowDown")
}
