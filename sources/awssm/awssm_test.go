package awssm

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	secret *string
	err    error
}

func (m *mockClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return &secretsmanager.GetSecretValueOutput{
		SecretString: m.secret,
	}, m.err
}

func TestNew(t *testing.T) {
	tt := []struct {
		name     string
		opts     []Option
		expected *source
	}{
		{
			name: "with all options",
			opts: []Option{
				WithClient(&mockClient{}),
				WithRegion("us-west-2"),
				WithSecretID("test-secret"),
				WithProfile("test-profile"),
			},
			expected: &source{
				client:   &mockClient{},
				region:   "us-west-2",
				secretID: "test-secret",
				profile:  "test-profile",
			},
		},
		{
			name:     "with no options",
			opts:     []Option{},
			expected: &source{},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := New(tc.opts...)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestLoad(t *testing.T) {
	tt := []struct {
		name string

		source source

		expected    map[string]string
		expectError bool
	}{
		{
			name: "success",
			source: source{
				client: &mockClient{
					secret: strPtr(`{"key1":"value1","key2":"value2"}`),
				},
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "nil secret",
			source: source{
				client: &mockClient{
					secret: nil,
				},
			},
			expectError: true,
		},
		{
			name: "invalid json",
			source: source{
				client: &mockClient{
					secret: strPtr(`{"key1":"value1","key2":"value2`),
				},
			},
			expectError: true,
		},
		{
			name: "error",
			source: source{
				client: &mockClient{
					err: assert.AnError,
				},
			},
			expectError: true,
		},
		{
			name: "nil client",
			source: source{
				client: nil,
			},
			expectError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			envs, err := tc.source.Load()
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, envs)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
