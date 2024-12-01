package dotenv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tt := []struct {
		name        string
		content     string
		expected    map[string]string
		expectedErr bool
	}{
		{
			name:        "empty file",
			content:     "",
			expected:    map[string]string{},
			expectedErr: false,
		},
		{
			name:    "single variable",
			content: "KEY=value",
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:    "multiple variables",
			content: "KEY1=value1\nKEY2=value2",
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name:    "variables with empty values",
			content: "KEY1=\nKEY2=value2",
			expected: map[string]string{
				"KEY1": "",
				"KEY2": "value2",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile := filepath.Join(t.TempDir(), "test.env")
			err := os.WriteFile(tmpFile, []byte(tc.content), 0644)
			require.NoError(t, err)

			src := New(tmpFile)
			result, err := src.Load()

			if tc.expectedErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}

	t.Run("non-existent file", func(t *testing.T) {
		src := New("non-existent-file")
		_, err := src.Load()
		require.Error(t, err)
	})
}
