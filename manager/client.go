package manager

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/nekrassov01/ignoresync/params"
)

var _ ISTS = (*STS)(nil)

// ISTS defines the interface for STS client.
type ISTS interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// STS defines a struct for STS client.
type STS struct {
	*sts.Client
}

// Manager is a client for app state.
type Manager struct {
	Account string
	Region  string
	OSUser  string

	sts  ISTS
	salt []byte
	w    io.Writer
}

// New instantiates the Manager by performing STS GetCallerIdentity API call
// to retrieve the account ID and region, and returns the resolved Manager.
func New(ctx context.Context, w io.Writer, cfg aws.Config) (*Manager, error) {
	m := &Manager{
		sts:    &STS{Client: sts.NewFromConfig(cfg)},
		Region: cfg.Region,
		OSUser: getUser(),
		w:      w,
	}
	in := &sts.GetCallerIdentityInput{}
	out, err := m.sts.GetCallerIdentity(ctx, in)
	if err != nil {
		return nil, NewAuthError(fmt.Errorf("not authenticated: %w", err))
	}
	m.Account = aws.ToString(out.Account)
	m.salt = make([]byte, 0, len(m.Account)+len(m.Region))
	m.salt = append(m.salt, m.Account...)
	m.salt = append(m.salt, m.Region...)
	return m, nil
}

// getUser retrieves the current OS user name.
func getUser() string {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}
	if user == "" {
		user = params.DefaultUserName
	}
	return user
}
