package walker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type unset struct {
	Value string
}

func (u *unset) UnmarshalText(text []byte) error {
	u.Value = "Hello!"
	return nil
}

func TestWalk(t *testing.T) {
	tt := []struct {
		name        string
		walker      *Walker
		cfg         any
		envs        map[string]string
		expected    any
		expectedErr bool
	}{
		{
			name:   "simple",
			walker: New(),
			cfg: &struct {
				Simple string
			}{},
			envs: map[string]string{
				"SIMPLE": "hello",
			},
			expected: &struct {
				Simple string
			}{Simple: "hello"},
		},
		{
			name:   "non-pointer input",
			walker: New(),
			cfg: struct {
				Simple string
			}{},
			expectedErr: true,
		},
		{
			name:        "pointer to non-struct",
			walker:      New(),
			cfg:         new(string),
			expectedErr: true,
		},
		{
			name:   "nested struct",
			walker: New(),
			cfg: &struct {
				Nested struct {
					Value string
				}
			}{},
			envs: map[string]string{
				"NESTED_VALUE": "nested-value",
			},
			expected: &struct {
				Nested struct {
					Value string
				}
			}{
				Nested: struct {
					Value string
				}{
					Value: "nested-value",
				},
			},
		},
		{
			name:   "delimited slice",
			walker: New(),
			cfg: &struct {
				Slice []string
			}{},
			envs: map[string]string{
				"SLICE": "a,b,c",
			},
			expected: &struct {
				Slice []string
			}{
				Slice: []string{"a", "b", "c"},
			},
		},
		{
			name:   "indexed slice",
			walker: New(),
			cfg: &struct {
				Slice []string `index:"true"`
			}{},
			envs: map[string]string{
				"SLICE_0": "a",
				"SLICE_1": "b",
				"SLICE_2": "c",
			},
			expected: &struct {
				Slice []string `index:"true"`
			}{
				Slice: []string{"a", "b", "c"},
			},
		},
		{
			name:   "slice of structs",
			walker: New(),
			cfg: &struct {
				Slice []struct{ Value string }
			}{},
			envs: map[string]string{
				"SLICE_0_VALUE": "value",
			},
			expected: &struct {
				Slice []struct{ Value string }
			}{
				Slice: []struct{ Value string }{
					{Value: "value"},
				},
			},
		},
		{
			name:   "delimited map",
			walker: New(),
			cfg: &struct {
				Map map[string]string
			}{},
			envs: map[string]string{
				"MAP": "a:b,c:d",
			},
			expected: &struct {
				Map map[string]string
			}{
				Map: map[string]string{"a": "b", "c": "d"},
			},
		},
		{
			name:   "flat map",
			walker: New(),
			cfg: &struct {
				Map map[string]string
			}{},
			envs: map[string]string{
				"MAP_KEY1": "value1",
				"MAP_KEY2": "value2",
			},
			expected: &struct {
				Map map[string]string
			}{
				Map: map[string]string{"key1": "value1", "key2": "value2"},
			},
		},
		{
			name:   "map of structs",
			walker: New(),
			cfg: &struct {
				Map map[string]struct{ Value string }
			}{},
			envs: map[string]string{
				"MAP_KEY1_VALUE": "value1",
				"MAP_KEY2_VALUE": "value2",
			},
			expected: &struct {
				Map map[string]struct{ Value string }
			}{
				Map: map[string]struct{ Value string }{
					"key1": {Value: "value1"},
					"key2": {Value: "value2"},
				},
			},
		},
		{
			name:   "ignore tag",
			walker: New(),
			cfg: &struct {
				Ignored string `env:"-"`
				Normal  string
			}{},
			envs: map[string]string{
				"IGNORED": "should-not-set",
				"NORMAL":  "normal-value",
			},
			expected: &struct {
				Ignored string `env:"-"`
				Normal  string
			}{
				Normal: "normal-value",
			},
		},
		{
			name:   "custom delimiter",
			walker: New(),
			cfg: &struct {
				List []string `delim:"|"`
			}{},
			envs: map[string]string{
				"LIST": "a|b|c",
			},
			expected: &struct {
				List []string `delim:"|"`
			}{
				List: []string{"a", "b", "c"},
			},
		},
		{
			name:   "map with custom separator and delimiter",
			walker: New(),
			cfg: &struct {
				Dict map[string]string `sep:"|" delim:";"`
			}{},
			envs: map[string]string{
				"DICT": "key1|val1;key2|val2",
			},
			expected: &struct {
				Dict map[string]string `sep:"|" delim:";"`
			}{
				Dict: map[string]string{
					"key1": "val1",
					"key2": "val2",
				},
			},
		},
		{
			name:   "init values slice with match",
			walker: New(),
			cfg: &struct {
				Slice []string
			}{},
			envs: map[string]string{
				"SLICE_0": "value1",
				"SLICE_1": "value2",
			},
			expected: &struct {
				Slice []string
			}{
				Slice: []string{"value1", "value2"},
			},
		},
		{
			name:   "init values slice with no match",
			walker: New(),
			cfg: &struct {
				Slice []string
			}{},
			envs: map[string]string{},
			expected: &struct {
				Slice []string
			}{},
		},
		{
			name:   "init values map with match",
			walker: New(),
			cfg: &struct {
				Map map[string]string
			}{},
			envs: map[string]string{
				"MAP_KEY1": "value1",
				"MAP_KEY2": "value2",
			},
			expected: &struct {
				Map map[string]string
			}{
				Map: map[string]string{"key1": "value1", "key2": "value2"},
			},
		},
		{
			name:   "init values map with no match",
			walker: New(),
			cfg: &struct {
				Map map[string]string
			}{},
			envs: map[string]string{},
			expected: &struct {
				Map map[string]string
			}{},
		},
		{
			name:   "never init slice",
			walker: New(),
			cfg: &struct {
				Slice []string `init:"never"`
			}{},
			envs: map[string]string{
				"SLICE_0": "value",
			},
			expected: &struct {
				Slice []string `init:"never"`
			}{},
		},
		{
			name:   "always init slice",
			walker: New(),
			cfg: &struct {
				Slice []string `init:"always"`
			}{},
			envs: map[string]string{},
			expected: &struct {
				Slice []string `init:"always"`
			}{
				Slice: []string{},
			},
		},
		{
			name:   "never init map",
			walker: New(),
			cfg: &struct {
				Map map[string]string `init:"never"`
			}{},
			envs: map[string]string{
				"MAP_KEY": "value",
			},
			expected: &struct {
				Map map[string]string `init:"never"`
			}{},
		},
		{
			name:   "always init map",
			walker: New(),
			cfg: &struct {
				Map map[string]string `init:"always"`
			}{},
			envs: map[string]string{},
			expected: &struct {
				Map map[string]string `init:"always"`
			}{
				Map: map[string]string{},
			},
		},
		{
			name:   "never init pointer",
			walker: New(),
			cfg: &struct {
				Pointer *struct{ Value string } `init:"never"`
			}{},
			envs: map[string]string{
				"POINTER_VALUE": "value",
			},
			expected: &struct {
				Pointer *struct{ Value string } `init:"never"`
			}{
				Pointer: nil,
			},
		},
		{
			name:   "always init pointer",
			walker: New(),
			cfg: &struct {
				Pointer *struct{ Value string } `init:"always"`
			}{},
			envs: map[string]string{},
			expected: &struct {
				Pointer *struct{ Value string } `init:"always"`
			}{
				Pointer: &struct{ Value string }{
					Value: "",
				},
			},
		},
		{
			name:   "multiple types",
			walker: New(),
			cfg: &struct {
				String  string
				Int     int
				Bool    bool
				Float   float64
				Strings []string
			}{},
			envs: map[string]string{
				"STRING":  "str-value",
				"INT":     "42",
				"BOOL":    "true",
				"FLOAT":   "3.14",
				"STRINGS": "a,b,c",
			},
			expected: &struct {
				String  string
				Int     int
				Bool    bool
				Float   float64
				Strings []string
			}{
				String:  "str-value",
				Int:     42,
				Bool:    true,
				Float:   3.14,
				Strings: []string{"a", "b", "c"},
			},
		},
		{
			name:   "no decode unset",
			walker: New(),
			cfg: &struct {
				Unset unset
			}{},
			envs: map[string]string{},
			expected: &struct {
				Unset unset
			}{
				Unset: unset{},
			},
		},
		{
			name:   "decode unset",
			walker: New(),
			cfg: &struct {
				Unset unset `decodeunset:"true"`
			}{},
			envs: map[string]string{},
			expected: &struct {
				Unset unset `decodeunset:"true"`
			}{
				Unset: unset{
					Value: "Hello!",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			w := tc.walker
			w.Matcher.EnvVars = tc.envs

			err := w.Walk(tc.cfg)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, tc.cfg)
			}
		})
	}
}
