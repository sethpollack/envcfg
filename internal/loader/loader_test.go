package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSource struct {
	envs map[string]string
}

func (s *testSource) Load() (map[string]string, error) {
	return s.envs, nil
}

func TestLoad(t *testing.T) {
	tt := []struct {
		name        string
		loader      Loader
		expected    map[string]string
		expectedErr bool
	}{
		{
			name: "simple",
			loader: Loader{
				Sources: []Source{
					&testSource{envs: map[string]string{"TEST_KEY": "value"}},
				},
			},
			expected: map[string]string{"TEST_KEY": "value"},
		},
		{
			name: "defaults with no overrides",
			loader: Loader{
				Sources:  []Source{&testSource{envs: map[string]string{"TEST_KEY": "value"}}},
				Defaults: map[string]string{"DEFAULT_KEY": "default_value"},
			},
			expected: map[string]string{
				"DEFAULT_KEY": "default_value",
				"TEST_KEY":    "value",
			},
		},
		{
			name: "defaults with overrides",
			loader: Loader{
				Sources:  []Source{&testSource{envs: map[string]string{"TEST_KEY": "value", "DEFAULT_KEY": "default_override_value"}}},
				Defaults: map[string]string{"DEFAULT_KEY": "default_value"},
			},
			expected: map[string]string{
				"DEFAULT_KEY": "default_override_value",
				"TEST_KEY":    "value",
			},
		},
		{
			name: "with filter",
			loader: Loader{
				Sources: []Source{&testSource{envs: map[string]string{
					"TEST_KEY":  "value",
					"OTHER_KEY": "other_value",
				}}},
				Filters: []func(string) bool{func(key string) bool { return key == "TEST_KEY" }},
			},
			expected: map[string]string{"TEST_KEY": "value"},
		},
		{
			name: "with transform",
			loader: Loader{
				Sources:    []Source{&testSource{envs: map[string]string{"TEST_KEY": "value"}}},
				Transforms: []func(string) string{func(key string) string { return "TRANSFORMED_" + key }},
			},
			expected: map[string]string{"TRANSFORMED_TEST_KEY": "value"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			envs, err := tc.loader.Load()
			if tc.expectedErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.expected, envs)
		})
	}
}
