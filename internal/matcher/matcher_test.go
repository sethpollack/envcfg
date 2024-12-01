package matcher

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/sethpollack/envcfg/internal/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetValue(t *testing.T) {
	tempFile, err := os.CreateTemp("", "env.txt")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tempFile.WriteString("${OTHER_VAR}")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	tt := []struct {
		name    string
		Path    []tag.TagMap
		EnvVars map[string]string

		Expected    string
		ExpectedErr error
	}{
		{
			name: "not found",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR"},
					},
				},
			},
			Expected: "",
		},
		{
			name: "simple",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR"},
					},
				},
			},
			EnvVars:  map[string]string{"FOO_BAR": "foo"},
			Expected: "foo",
		},
		{
			name: "nested",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "APP"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_FOO_BAR": "foo"},
			Expected: "foo",
		},
		{
			name: "deep nested",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "APP"},
					},
				},
				{
					FieldName: "Other",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "OTHER"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_OTHER_FOO_BAR": "foo"},
			Expected: "foo",
		},
		{
			name: "fallback",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "APP"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "FooBar"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_FOOBAR": "foo"},
			Expected: "foo",
		},
		{
			name: "required",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"required": "true"}},
					},
				},
			},
			ExpectedErr: fmt.Errorf("required field FooBar not found"),
		},
		{
			name: "nested required",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "APP"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"required": "true"}},
					},
				},
			},
			ExpectedErr: fmt.Errorf("required field App.FooBar not found"),
		},
		{
			name: "notempty",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "APP"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"notempty": "true"}},
					},
				},
			},
			EnvVars:     map[string]string{"APP_FOO_BAR": ""},
			ExpectedErr: fmt.Errorf("environment variable APP_FOO_BAR is empty"),
		},
		{
			name: "expand",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "APP"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"expand": "true"}},
					},
				},
			},
			EnvVars: map[string]string{
				"OTHER_VAR":   "other",
				"APP_FOO_BAR": "${OTHER_VAR}",
			},
			Expected: "other",
		},
		{
			name: "default",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "APP"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"default": "foo"}},
					},
				},
			},
			Expected: "foo",
		},
		{
			name: "default + expand",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"default": "${OTHER_VAR}", "expand": "true"}},
					},
				},
			},
			EnvVars: map[string]string{
				"OTHER_VAR": "other",
			},
			Expected: "other",
		},
		{
			name: "file",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"file": "true"}},
					},
				},
			},
			EnvVars: map[string]string{
				"FOO_BAR": tempFile.Name(),
			},
			Expected: "${OTHER_VAR}",
		},
		{
			name: "expand + file",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"file": "true", "expand": "true"}},
					},
				},
			},
			EnvVars: map[string]string{
				"OTHER_VAR": "other",
				"FOO_BAR":   tempFile.Name(),
			},
			Expected: "other",
		},
		{
			name: "invalid file path",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR", Options: map[string]string{"file": "true"}},
					},
				},
			},
			EnvVars:     map[string]string{"FOO_BAR": "invalid"},
			ExpectedErr: fmt.Errorf("open invalid: no such file or directory"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := New()
			m.EnvVars = tc.EnvVars

			actual, _, err := m.GetValue(tc.Path)

			if tc.ExpectedErr != nil {
				assert.EqualError(t, err, tc.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tt := []struct {
		Name     string
		Path     []tag.TagMap
		EnvVars  map[string]string
		Expected bool
	}{
		{
			Name: "not found",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "App"},
					},
				},
			},
			Expected: false,
		},
		{
			Name: "found",
			Path: []tag.TagMap{
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR"},
					},
				},
			},
			EnvVars:  map[string]string{"FOO_BAR": "foo"},
			Expected: true,
		},
		{
			Name: "fallback",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "App"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "FooBar"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_FOOBAR": "foo"},
			Expected: true,
		},
		{
			Name: "fallback mixed",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"struct": {Name: "struct", Value: "App"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"struct_snake": {Name: "struct_snake", Value: "foo_bar"},
					},
				},
				{
					FieldName: "Baz",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "custom"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_FOO_BAR_CUSTOM": "foo"},
			Expected: true,
		},
		{
			Name: "nested",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "APP"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_FOO_BAR": "foo"},
			Expected: true,
		},
		{
			Name: "complex",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "APP"},
					},
				},
				{
					FieldName: "Slice",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "SLICE"},
					},
				},
				{
					FieldName: "0",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "0"},
					},
				},
				{
					FieldName: "FooBar",
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "FOO_BAR"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_SLICE_0_FOO_BAR": "foo"},
			Expected: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := New()
			m.EnvVars = tc.EnvVars

			assert.Equal(t, tc.Expected, m.HasPrefix(tc.Path))
		})
	}
}

func TestGetMapKeys(t *testing.T) {
	tt := []struct {
		Name     string
		Path     []tag.TagMap
		EnvVars  map[string]string
		Expected []string
	}{
		{
			Name: "simple",
			Path: []tag.TagMap{
				{
					FieldName: "Map",
					Type:      reflect.TypeOf(map[string]string{}),
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "MAP"},
					},
				},
			},
			EnvVars:  map[string]string{"MAP_FOO_BAR": "foo"},
			Expected: []string{"foo_bar"},
		},
		{
			Name: "nested",
			Path: []tag.TagMap{
				{
					FieldName: "App",
					Type:      reflect.TypeOf(map[string]string{}),
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "APP"},
					},
				},
				{
					FieldName: "Map",
					Type:      reflect.TypeOf(map[string]string{}),
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "MAP"},
					},
				},
			},
			EnvVars:  map[string]string{"APP_MAP_FOO_BAR_BAZ": "foo"},
			Expected: []string{"foo_bar_baz"},
		},
		{
			Name: "map of structs",
			Path: []tag.TagMap{
				{
					FieldName: "Map",
					Type:      reflect.TypeOf(map[string]struct{ Key string }{}),
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "MAP"},
					},
				},
			},
			EnvVars:  map[string]string{"MAP_FOO_KEY": "foo", "MAP_BAZ_KEY": "baz"},
			Expected: []string{"foo", "baz"},
		},
		{
			Name: "map of slices",
			Path: []tag.TagMap{
				{
					FieldName: "Map",
					Type:      reflect.TypeOf(map[string][]string{}),
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "MAP"},
					},
				},
			},
			EnvVars:  map[string]string{"MAP_SLICE_0": "foo", "MAP_SLICE_1": "bar"},
			Expected: []string{"slice"},
		},
		{
			Name: "map of slices of structs",
			Path: []tag.TagMap{
				{
					FieldName: "Map",
					Type:      reflect.TypeOf(map[string][]struct{ Key string }{}),
					Tags: map[string]tag.Tag{
						"env": {Name: "env", Value: "MAP"},
					},
				},
			},
			EnvVars:  map[string]string{"MAP_SLICE_0_KEY": "foo", "MAP_SLICE_1_KEY": "bar"},
			Expected: []string{"slice"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			m := New()
			m.EnvVars = tc.EnvVars

			assert.ElementsMatch(t, tc.Expected, m.GetMapKeys(tc.Path))
		})
	}
}
