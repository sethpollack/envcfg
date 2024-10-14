package matcher

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/sethpollack/envcfg/internal/loader"
	"github.com/stretchr/testify/assert"
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
		name      string
		FieldName string
		Tag       string
		Prefixes  []string
		EnvVars   map[string]string

		Expected    string
		ExpectedErr error
	}{
		{
			name:      "not found",
			FieldName: "FooBar",
			Prefixes:  []string{"APP"},
			Expected:  "",
		},
		{
			name:      "no env tag",
			FieldName: "FooBar",
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"APP_FOO_BAR": "foo"},
			Expected:  "foo",
		},
		{
			name:      "no env tag with no prefix",
			FieldName: "FooBar",
			EnvVars:   map[string]string{"FOO_BAR": "foo"},
			Expected:  "foo",
		},
		{
			name:      "env tag with no options",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR"`,
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"APP_FOO_BAR": "foo"},
			Expected:  "foo",
		},
		{
			name:      "env tag with no prefix",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR"`,
			EnvVars:   map[string]string{"FOO_BAR": "foo"},
			Expected:  "foo",
		},
		{
			name:      "env tag with multiple prefixes",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR"`,
			Prefixes:  []string{"APP", "OTHER"},
			EnvVars:   map[string]string{"APP_OTHER_FOO_BAR": "foo"},
			Expected:  "foo",
		},
		{
			name:      "env tag with no exactmatch",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR"`,
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"FOO_BAR": "foo"},
			Expected:  "",
		},
		{
			name:        "env tag with required",
			FieldName:   "FooBar",
			Tag:         `env:"FOO_BAR,required"`,
			Prefixes:    []string{"APP"},
			EnvVars:     map[string]string{"FOO_BAR": "foo"},
			Expected:    "",
			ExpectedErr: fmt.Errorf("required field FooBar not found"),
		},
		{
			name:        "env tag with notempty",
			FieldName:   "FooBar",
			Tag:         `env:"FOO_BAR,notempty"`,
			Prefixes:    []string{"APP"},
			EnvVars:     map[string]string{"APP_FOO_BAR": ""},
			Expected:    "",
			ExpectedErr: fmt.Errorf("environment variable APP_FOO_BAR is empty"),
		},
		{
			name:      "env tag with best match",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR,match=best"`,
			Prefixes:  []string{"APP", "OTHER"},
			EnvVars:   map[string]string{"OTHER_FOO_BAR": "best"},
			Expected:  "best",
		},
		{
			name:      "env tag with exact match",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR,match=exact"`,
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"FOO_BAR": "exact"},
			Expected:  "exact",
		},
		{
			name:      "env tag with expand",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR,expand"`,
			Prefixes:  []string{"APP"},
			EnvVars: map[string]string{
				"OTHER_VAR":   "other",
				"APP_FOO_BAR": "${OTHER_VAR}",
			},
			Expected: "other",
		},
		{
			name:      "env tag with default",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR,default=foo"`,
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{},
			Expected:  "foo",
		},
		{
			name:      "env tag with default + expand",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR,default=${OTHER_VAR},expand"`,
			Prefixes:  []string{"APP"},
			EnvVars: map[string]string{
				"OTHER_VAR": "other",
			},
			Expected: "other",
		},
		{
			name:      "env tag with file",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR,file"`,
			Prefixes:  []string{"APP"},
			EnvVars: map[string]string{
				"APP_FOO_BAR": tempFile.Name(),
			},
			Expected: "${OTHER_VAR}",
		},
		{
			name:      "env tag with expand + file",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR,expand,file"`,
			Prefixes:  []string{"APP"},
			EnvVars: map[string]string{
				"OTHER_VAR":   "other",
				"APP_FOO_BAR": tempFile.Name(),
			},
			Expected: "other",
		},
		{
			name:        "invalid file path",
			FieldName:   "FooBar",
			Tag:         `env:"FOO_BAR,file"`,
			Prefixes:    []string{"APP"},
			EnvVars:     map[string]string{"APP_FOO_BAR": "invalid"},
			ExpectedErr: fmt.Errorf("open invalid: no such file or directory"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := New()
			if err := m.Build(loader.WithEnvVarsSource(tc.EnvVars)); err != nil {
				assert.NoError(t, err)
			}

			actual, _, err := m.GetValue(
				reflect.StructField{
					Name: tc.FieldName,
					Tag:  reflect.StructTag(tc.Tag),
				},
				tc.Prefixes,
			)
			if tc.ExpectedErr != nil {
				assert.EqualError(t, err, tc.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestGetPrefix(t *testing.T) {
	tt := []struct {
		name      string
		FieldName string
		Tag       string
		Prefixes  []string
		EnvVars   map[string]string

		Expected    string
		ExpectedErr error
	}{
		{
			name:      "has env tag",
			FieldName: "FooBar",
			Tag:       `env:"FOO_BAR"`,
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"APP_FOO_BAR": "foo"},
			Expected:  "FOO_BAR",
		},
		{
			name:      "no env tag",
			FieldName: "FooBar",
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"APP_FOOBAR": "foo"},
			Expected:  "FOOBAR",
		},
		{
			name:      "no match, fallback to fieldName",
			FieldName: "FooBar",
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"FOOBAR": "foo"},
			Expected:  "FOOBAR",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := New()
			if err := m.Build(loader.WithEnvVarsSource(tc.EnvVars)); err != nil {
				assert.NoError(t, err)
			}

			actual := m.GetPrefix(
				reflect.StructField{
					Name: tc.FieldName,
					Tag:  reflect.StructTag(tc.Tag),
				},
				tc.Prefixes,
			)

			assert.Equal(t, tc.Expected, actual)
		})
	}
}

func TestGetMapKeys(t *testing.T) {
	tt := []struct {
		name      string
		FieldName string
		Type      reflect.Type
		Tag       string
		Prefixes  []string
		EnvVars   map[string]string

		Expected []string
	}{
		{
			name:      "map of strings",
			FieldName: "FooBar",
			Type:      reflect.TypeOf(map[string]string{}),
			Tag:       `env:"FOO_BAR"`,
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"APP_FOO_BAR": "foo", "APP_FOO_BAR_BAZ": "baz"},
			Expected:  []string{"foo_bar", "foo_bar_baz"},
		},
		{
			name:      "map of ints",
			FieldName: "FooBar",
			Type:      reflect.TypeOf(map[string]int{}),
			Tag:       `env:"FOO_BAR"`,
			Prefixes:  []string{"APP"},
			EnvVars:   map[string]string{"APP_FOO_BAR": "1", "APP_FOO_BAR_BAZ": "2"},
			Expected:  []string{"foo_bar", "foo_bar_baz"},
		},
		{
			name:      "map of structs, prefix match",
			FieldName: "FooBar",
			Type: reflect.TypeOf(map[string]struct {
				Name string `env:"NAME"`
			}{}),
			Tag:      `env:"FOO_BAR"`,
			Prefixes: []string{"APP"},
			EnvVars:  map[string]string{"APP_FOO_BAR_PRIMARY_NAME": "foo", "APP_FOO_BAR_SECONDARY_NAME": "baz"},
			Expected: []string{"primary", "secondary"},
		},
		{
			name:      "map of structs, exact match",
			FieldName: "FooBar",
			Type: reflect.TypeOf(map[string]struct {
				Name string `env:"NAME,match=exact"`
			}{}),
			Tag:      `env:"FOO_BAR"`,
			Prefixes: []string{"APP"},
			EnvVars:  map[string]string{"PRIMARY_NAME": "foo", "SECONDARY_NAME": "baz"},
			Expected: []string{"primary", "secondary"},
		},
		{
			name:      "map of structs, best match",
			FieldName: "FooBar",
			Type: reflect.TypeOf(map[string]struct {
				Name string `env:"NAME,match=best"`
			}{}),
			Tag:      `env:"FOO_BAR"`,
			Prefixes: []string{"APP", "OTHER"},
			EnvVars:  map[string]string{"OTHER_PRIMARY_NAME": "foo", "OTHER_SECONDARY_NAME": "baz"},
			Expected: []string{"primary", "secondary"},
		},
		{
			name:      "overlaping keys",
			FieldName: "FooBar",
			Type: reflect.TypeOf(map[string]struct {
				Key           string `env:"KEY"`
				OtherKey      string `env:"OTHER_KEY"`
				OtherOtherKey string `env:"OTHER_OTHER_KEY"`
			}{}),
			Tag:      `env:"FOO_BAR"`,
			Prefixes: []string{},
			EnvVars:  map[string]string{"FOO_BAR_PRIMARY_KEY": "", "FOO_BAR_SECONDARY_OTHER_KEY": "", "FOO_BAR_THIRD_OTHER_OTHER_KEY": ""},
			Expected: []string{"primary", "secondary", "third"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			m := New()
			if err := m.Build(loader.WithEnvVarsSource(tc.EnvVars)); err != nil {
				assert.NoError(t, err)
			}

			actual := m.GetMapKeys(
				reflect.StructField{
					Type: tc.Type,
					Name: tc.FieldName,
					Tag:  reflect.StructTag(tc.Tag),
				},
				tc.Prefixes,
			)

			assert.ElementsMatch(t, tc.Expected, actual)
		})
	}
}
