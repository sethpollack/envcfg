package mapenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tt := []struct {
		name     string
		input    map[string]string
		expected map[string]string
	}{
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: map[string]string{},
		},
		{
			name: "single key-value",
			input: map[string]string{
				"KEY": "value",
			},
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name: "multiple key-values",
			input: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "empty value",
			input: map[string]string{
				"KEY": "",
			},
			expected: map[string]string{
				"KEY": "",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			src := New(tc.input)
			result, err := src.Load()
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
