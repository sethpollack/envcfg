package osenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tt := []struct {
		name     string
		env      map[string]string
		expected map[string]string
	}{
		{
			name:     "empty environment",
			env:      map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "single variable",
			env: map[string]string{
				"TEST_KEY": "test_value",
			},
			expected: map[string]string{
				"TEST_KEY": "test_value",
			},
		},
		{
			name: "multiple variables",
			env: map[string]string{
				"TEST_KEY1": "value1",
				"TEST_KEY2": "value2",
			},
			expected: map[string]string{
				"TEST_KEY1": "value1",
				"TEST_KEY2": "value2",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			src := New()
			actual, err := src.Load()

			require.NoError(t, err)
			assert.Subset(t, actual, tc.expected)
		})
	}
}
