package manager

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var _ ISTS = (*mockManager)(nil)

// mockManager is a mock implementation of the ISTS interface for testing.
type mockManager struct {
	getCallerIdentityFunc func(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
	w                     io.Writer
}

// GetCallerIdentity calls the mocked GetCallerIdentity function.
func (o *mockManager) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	return o.getCallerIdentityFunc(ctx, params, optFns...)
}

// NewMockState creates a new State with default test values.
func NewMockState(mods ...func(*State)) *State {
	state := &State{
		KeyID: "11111111aaaaaaaa",
		MasterKeys: map[string][]byte{
			"11111111aaaaaaaa": {48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 48, 49, 50, 51, 52, 53},
		},
		Account: "012345678901",
		Region:  "us-east-1",
		Bucket: &BucketInfo{
			ARN: arn.ARN{
				Partition: "aws",
				Service:   "s3",
				Region:    "",
				AccountID: "",
				Resource:  "fe116c89d2a35d3089a8012f6624fe2d",
			},
		},
		SSEKey: &KMSKeyInfo{
			ARN: arn.ARN{
				Partition: "aws",
				Service:   "kms",
				Region:    "us-east-1",
				AccountID: "012345678901",
				Resource:  "alias/5fb10905a8a3288d7adb42fa0faa306d",
			},
			EncryptionContext: map[string]string{
				encryptionContextKey: "fe116c89d2a35d3089a8012f6624fe2d/alias/5fb10905a8a3288d7adb42fa0faa306d",
			},
		},
		CSEKey: &KMSKeyInfo{
			ARN: arn.ARN{
				Partition: "aws",
				Service:   "kms",
				Region:    "us-east-1",
				AccountID: "012345678901",
				Resource:  "alias/200bc2c483b95d35dcff1850cca56aef",
			},
			EncryptionContext: map[string]string{
				encryptionContextKey: "fe116c89d2a35d3089a8012f6624fe2d/alias/200bc2c483b95d35dcff1850cca56aef",
			},
		},
	}
	for _, mod := range mods {
		mod(state)
	}
	return state
}
