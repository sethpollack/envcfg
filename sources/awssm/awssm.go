package awssm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/sethpollack/envcfg/internal/loader"
)

var _ loader.Source = (*source)(nil)

type Client interface {
	GetSecretValue(ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type Option func(*source)

// WithClient sets the AWS Secrets Manager client
func WithClient(client Client) Option {
	return func(s *source) {
		s.client = client
	}
}

// WithRegion sets the AWS region where secrets are stored
func WithRegion(region string) Option {
	return func(s *source) {
		s.region = region
	}
}

// WithSecretID sets the ID or ARN of the secret to load
func WithSecretID(id string) Option {
	return func(s *source) {
		s.secretID = id
	}
}

// WithProfile sets the AWS profile to use
func WithProfile(profile string) Option {
	return func(s *source) {
		s.profile = profile
	}
}

type source struct {
	client   Client
	region   string
	secretID string
	profile  string
}

func New(opts ...Option) *source {
	s := &source{}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *source) Load() (map[string]string, error) {
	if s.client == nil {
		var cfgOpts []func(*config.LoadOptions) error

		cfgOpts = append(cfgOpts, config.WithRegion(s.region))

		if s.profile != "" {
			cfgOpts = append(cfgOpts, config.WithSharedConfigProfile(s.profile))
		}

		cfg, err := config.LoadDefaultConfig(context.Background(), cfgOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		s.client = secretsmanager.NewFromConfig(cfg)
	}

	input := &secretsmanager.GetSecretValueInput{
		SecretId: &s.secretID,
	}

	result, err := s.client.GetSecretValue(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return nil, fmt.Errorf("secret string is nil")
	}

	secretData := make(map[string]string)
	if err := json.Unmarshal([]byte(*result.SecretString), &secretData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret data: %w", err)
	}

	return secretData, nil
}
