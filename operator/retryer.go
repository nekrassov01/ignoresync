package operator

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
)

var _ retry.IsErrorRetryable = (*isErrorRetryer)(nil)

// retryer is a shared retryer instance used for all AWS API calls in the operator.
var retryer = NewRetryer()

// codes is the set of HTTP status codes that should be retried.
// 429 Too Many Requests not included in DefaultRetryableHTTPStatusCodes.
var codes = retry.RetryableHTTPStatusCode{
	Codes: map[int]struct{}{
		429: {},
	},
}

// NewRetryer creates a new retryer with the custom retry logic and backoff strategy.
func NewRetryer() aws.RetryerV2 {
	return retry.NewStandard(
		func(o *retry.StandardOptions) {
			o.MaxAttempts = MaxRetryAttempts
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
