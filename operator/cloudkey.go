package operator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/nekrassov01/ignoresync/manager"
)

// generateCloudKey generates a new key using AWS KMS.
func (o *Operator) generateCloudKey(ctx context.Context, state *manager.State) (plaintext []byte, ciphertext []byte, err error) {
	in := &kms.GenerateDataKeyInput{
		KeyId:             aws.String(state.CSEKey.ARN.String()),
		KeySpec:           types.DataKeySpecAes256,
		EncryptionContext: state.CSEKey.EncryptionContext,
	}
	out, err := o.kms.GenerateDataKey(ctx, in)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate cloud key: %w", err)
	}
	return out.Plaintext, out.CiphertextBlob, nil
}

// decryptCloudKey decrypts the given encrypted data using AWS KMS.
func (o *Operator) decryptCloudKey(ctx context.Context, state *manager.State, ciphertext []byte) ([]byte, error) {
	in := &kms.DecryptInput{
		CiphertextBlob:    ciphertext,
		EncryptionContext: state.CSEKey.EncryptionContext,
	}
	out, err := o.kms.Decrypt(ctx, in)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with cloud key: %w", err)
	}
	return out.Plaintext, nil
}
