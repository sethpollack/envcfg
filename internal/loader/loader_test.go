package loader

import (
	"errors"
	"testing"

	errs "github.com/sethpollack/envcfg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testSource struct {
	envs map[string]string
	err  error
}

func (s *testSource) Load() (map[string]string, error) {
	return s.envs, s.err
}

func TestLoad(t *testing.T) {
	tt := []struct {
		name        string
		loader      Loader
		expected    map[string]string
		expectedErr error
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
		{
			name: "with error",
			loader: Loader{
				Sources: []Source{&testSource{err: errors.New("test error")}},
			},
			expectedErr: errs.ErrLoadEnv,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			envs, err := tc.loader.Load()

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, envs)
			}
		})
	}
}
