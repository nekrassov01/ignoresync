package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAWSConfig loads the AWS configuration with the given region and profile.
func LoadAWSConfig(ctx context.Context, region, profile string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRetryer(NewRetryer),
	)
	if err != nil {
		return cfg, NewConfigError(fmt.Errorf("failed to load aws config: %w", err))
	}
	if region != "" {
		cfg.Region = region
	}
	if cfg.Region == "" {
		cfg.Region = defaultRegion
	}
	return cfg, nil
}
