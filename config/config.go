package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadAWSConfig loads the AWS configuration with the given region and profile.
func LoadAWSConfig(ctx context.Context, region, profile string) (aws.Config, error) {
	opts := make([]func(*config.LoadOptions) error, 0, 2)
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, NewConfigError(fmt.Errorf("failed to load AWS config: %w", err))
	}
	if region != "" {
		cfg.Region = region
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	return cfg, nil
}
