package matcher

import (
	"os"
	"reflect"
	"testing"

	errs "github.com/sethpollack/envcfg/errors"
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

	tt := map[string]struct {
		Path    []tag.TagMap
		EnvVars map[string]string

		// Options
		Required        bool
		NotEmpty        bool
		Expand          bool
		DisableFallback bool

		Expected          string
		ExpectedIsFound   bool
		ExpectedIsDefault bool
		ExpectedErr       error
	}{
		"not found": {
			Path: parsePath(
				element{FieldName: "FooBar"},
			),
			Expected: "",
		},
		"simple": {
			Path: parsePath(
				element{FieldName: "FooBar"},
			),
			EnvVars:         map[string]string{"FOO_BAR": "foo"},
			Expected:        "foo",
			ExpectedIsFound: true,
		},
		"nested": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar"},
			),
			EnvVars:         map[string]string{"APP_FOO_BAR": "foo"},
			Expected:        "foo",
			ExpectedIsFound: true,
		},
		"deep nested": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "Other"},
				element{FieldName: "FooBar"},
			),
			EnvVars:         map[string]string{"APP_OTHER_FOO_BAR": "foo"},
			Expected:        "foo",
			ExpectedIsFound: true,
		},
		"fallback": {
			Path: parsePath(
				element{FieldName: "App", TagStr: `env:"APP"`},
				element{FieldName: "FooBar", TagStr: `json:"foo_bar"`},
				element{FieldName: "Baz", TagStr: `yaml:"baz"`},
			),
			EnvVars:         map[string]string{"APP_FOOBAR_BAZ": "foo"},
			Expected:        "foo",
			ExpectedIsFound: true,
		},
		"disable fallback": {
			Path: parsePath(
				element{FieldName: "FooBar", TagStr: `json:"foo_bar"`},
			),
			EnvVars:         map[string]string{"FOO_BAR": "foo"},
			DisableFallback: true,
			Expected:        "",
			ExpectedIsFound: false,
		},
		"required": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `env:",required=true"`},
			),
			ExpectedErr: errs.ErrRequired,
		},
		"required alt": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `required:"true"`},
			),
			ExpectedErr: errs.ErrRequired,
		},
		"required override": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar"},
			),
			Required:    true,
			ExpectedErr: errs.ErrRequired,
		},
		"notempty": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `env:",notempty=true"`},
			),
			EnvVars:     map[string]string{"APP_FOO_BAR": ""},
			ExpectedErr: errs.ErrNotEmpty,
		},
		"notempty alt": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `notempty:"true"`},
			),
			EnvVars:     map[string]string{"APP_FOO_BAR": ""},
			ExpectedErr: errs.ErrNotEmpty,
		},
		"notempty override": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar"},
			),
			EnvVars:     map[string]string{"APP_FOO_BAR": ""},
			NotEmpty:    true,
			ExpectedErr: errs.ErrNotEmpty,
		},
		"expand": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `env:",expand=true"`},
			),
			EnvVars: map[string]string{
				"OTHER_VAR":   "other",
				"APP_FOO_BAR": "${OTHER_VAR}",
			},
			Expected:        "other",
			ExpectedIsFound: true,
		},
		"expand alt": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `expand:"true"`},
			),
			EnvVars: map[string]string{
				"OTHER_VAR":   "other",
				"APP_FOO_BAR": "${OTHER_VAR}",
			},
			Expected:        "other",
			ExpectedIsFound: true,
		},
		"expand override": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar"},
			),
			EnvVars: map[string]string{
				"OTHER_VAR":   "other",
				"APP_FOO_BAR": "${OTHER_VAR}",
			},
			Expand:          true,
			Expected:        "other",
			ExpectedIsFound: true,
		},
		"default": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `env:",default=foo"`},
			),
			Expected:          "foo",
			ExpectedIsFound:   false,
			ExpectedIsDefault: true,
		},
		"default alt": {
			Path: parsePath(
				element{FieldName: "App"},
				element{FieldName: "FooBar", TagStr: `default:"foo"`},
			),
			Expected:          "foo",
			ExpectedIsFound:   false,
			ExpectedIsDefault: true,
		},
		"default + expand": {
			Path: parsePath(
				element{FieldName: "FooBar", TagStr: `default:"${OTHER_VAR}" expand:"true"`},
			),
			EnvVars: map[string]string{
				"OTHER_VAR": "other",
			},
			Expected:          "other",
			ExpectedIsFound:   false,
			ExpectedIsDefault: true,
		},
		"file": {
			Path: parsePath(
				element{FieldName: "FooBar", TagStr: `env:",file=true"`},
			),
			EnvVars: map[string]string{
				"FOO_BAR": tempFile.Name(),
			},
			Expected:        "${OTHER_VAR}",
			ExpectedIsFound: true,
		},
		"file alt": {
			Path: parsePath(
				element{FieldName: "FooBar", TagStr: `file:"true"`},
			),
			EnvVars: map[string]string{
				"FOO_BAR": tempFile.Name(),
			},
			Expected:        "${OTHER_VAR}",
			ExpectedIsFound: true,
		},
		"expand + file": {
			Path: parsePath(
				element{FieldName: "FooBar", TagStr: `file:"true" expand:"true"`},
			),
			EnvVars: map[string]string{
				"OTHER_VAR": "other",
				"FOO_BAR":   tempFile.Name(),
			},
			Expected:        "other",
			ExpectedIsFound: true,
		},
		"invalid file path": {
			Path: parsePath(
				element{FieldName: "FooBar", TagStr: `file:"true"`},
			),
			EnvVars:     map[string]string{"FOO_BAR": "invalid"},
			ExpectedErr: errs.ErrReadFile,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			m := New()

			m.EnvVars = tc.EnvVars
			m.Required = tc.Required
			m.NotEmpty = tc.NotEmpty
			m.Expand = tc.Expand
			m.DisableFallback = tc.DisableFallback

			actual, isFound, isDefault, err := m.GetValue(tc.Path)

			if tc.ExpectedErr != nil {
				assert.ErrorIs(t, err, tc.ExpectedErr)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.Expected, actual)
			assert.Equal(t, tc.ExpectedIsDefault, isDefault)
			assert.Equal(t, tc.ExpectedIsFound, isFound)
		})
	}
}

func TestHasPrefix(t *testing.T) {
	tt := map[string]struct {
		Path     []tag.TagMap
		EnvVars  map[string]string
		Expected bool
	}{
		"not found": {
			Path: parsePath(
				element{FieldName: "App"},
			),
			Expected: false,
		},
		"found": {
			Path: parsePath(
				element{FieldName: "FooBar", TagStr: `env:"FOO_BAR"`},
			),
			EnvVars:  map[string]string{"FOO_BAR": "foo"},
			Expected: true,
		},
		"fallback": {
			Path: parsePath(
				element{FieldName: "App", TagStr: `struct:"App"`},
				element{FieldName: "FooBar", TagStr: `struct:"FooBar"`},
			),
			EnvVars:  map[string]string{"APP_FOOBAR": "foo"},
			Expected: true,
		},
		"fallback mixed": {
			Path: parsePath(
				element{FieldName: "App", TagStr: `struct:"App"`},
				element{FieldName: "FooBar", TagStr: `struct_snake:"foo_bar"`},
				element{FieldName: "Baz", TagStr: `env:"custom"`},
			),
			EnvVars:  map[string]string{"APP_FOO_BAR_CUSTOM": "foo"},
			Expected: true,
		},
		"nested": {
			Path: parsePath(
				element{FieldName: "App", TagStr: `env:"APP"`},
				element{FieldName: "FooBar", TagStr: `env:"FOO_BAR"`},
			),
			EnvVars:  map[string]string{"APP_FOO_BAR": "foo"},
			Expected: true,
		},
		"complex": {
			Path: parsePath(
				element{FieldName: "App", TagStr: `env:"APP"`},
				element{FieldName: "Slice", TagStr: `env:"SLICE"`},
				element{FieldName: "0", TagStr: `env:"0"`},
				element{FieldName: "FooBar", TagStr: `env:"FOO_BAR"`},
			),
			EnvVars:  map[string]string{"APP_SLICE_0_FOO_BAR": "foo"},
			Expected: true,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			m := New()
			m.EnvVars = tc.EnvVars

			assert.Equal(t, tc.Expected, m.HasPrefix(tc.Path))
		})
	}
}

func TestGetMapKeys(t *testing.T) {
	tt := map[string]struct {
		Path     []tag.TagMap
		EnvVars  map[string]string
		Expected []string
	}{
		"empty": {
			Path:     []tag.TagMap{},
			EnvVars:  map[string]string{},
			Expected: []string{},
		},
		"simple": {
			Path: parsePath(
				element{
					FieldName: "Map",
					TagStr:    `env:"MAP"`,
					Type:      reflect.TypeOf(map[string]string{}),
				},
			),
			EnvVars:  map[string]string{"MAP_FOO_BAR": "foo"},
			Expected: []string{"foo_bar"},
		},
		"prefixed": {
			Path: parsePath(
				element{FieldName: "App", TagStr: `env:"APP"`},
				element{
					FieldName: "Map",
					TagStr:    `env:"MAP"`,
					Type:      reflect.TypeOf(map[string]string{}),
				},
			),
			EnvVars:  map[string]string{"APP_MAP_FOO_BAR": "foo"},
			Expected: []string{"foo_bar"},
		},
		"prefix alt": {
			Path: parsePath(
				element{FieldName: "App", TagStr: `toml:"app"`},
				element{FieldName: "Map", TagStr: `env:"MAP"`, Type: reflect.TypeOf(map[string]string{})},
			),
			EnvVars:  map[string]string{"APP_MAP_FOO_BAR_BAZ": "foo"},
			Expected: []string{"foo_bar_baz"},
		},
		"map of structs": {
			Path: parsePath(
				element{
					FieldName: "Map",
					TagStr:    `env:"MAP"`,
					Type:      reflect.TypeOf(map[string]struct{ Key string }{}),
				},
			),
			EnvVars:  map[string]string{"MAP_FOO_KEY": "foo", "MAP_BAZ_KEY": "baz"},
			Expected: []string{"foo", "baz"},
		},
		"ambiguous map of structs": {
			Path: parsePath(
				element{
					FieldName: "Map",
					TagStr:    `env:"MAP"`,
					Type: reflect.TypeOf(map[string]struct {
						Key      string
						CamelKey string
						SnakeKey string
						TagKey   string `env:"CUSTOM_KEY"`
					}{}),
				},
			),
			EnvVars: map[string]string{
				"MAP_A_KEY":          "foo",
				"MAP_A_CAMELKEY":     "foo",
				"MAP_A_SNAKE_KEY":    "foo",
				"MAP_A_CUSTOM_KEY":   "foo",
				"MAP_B_KEY":          "foo",
				"MAP_B_CAMELKEY":     "foo",
				"MAP_B_SNAKE_KEY":    "foo",
				"MAP_B_CUSTOM_KEY":   "foo",
				"MAP_C_KEY":          "foo",
				"MAP_C_CAMELKEY":     "foo",
				"MAP_C_SNAKE_KEY":    "foo",
				"MAP_C_CUSTOM_KEY":   "foo",
				"MAP_D_D_KEY":        "foo",
				"MAP_D_D_CAMELKEY":   "foo",
				"MAP_D_D_SNAKE_KEY":  "foo",
				"MAP_D_D_CUSTOM_KEY": "foo",
			},
			Expected: []string{"a", "b", "c", "d_d"},
		},
		"map of slices": {
			Path: parsePath(
				element{
					FieldName: "Map",
					TagStr:    `env:"MAP"`,
					Type:      reflect.TypeOf(map[string][]string{}),
				},
			),
			EnvVars:  map[string]string{"MAP_SLICE_0": "foo", "MAP_SLICE_1": "bar"},
			Expected: []string{"slice"},
		},
		"map of slices of structs": {
			Path: parsePath(
				element{
					FieldName: "Map",
					Type:      reflect.TypeOf(map[string][]struct{ Key string }{}),
					TagStr:    `env:"MAP"`,
				},
			),
			EnvVars:  map[string]string{"MAP_SLICE_0_KEY": "foo", "MAP_SLICE_1_KEY": "bar"},
			Expected: []string{"slice"},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			m := New()
			m.EnvVars = tc.EnvVars

			assert.ElementsMatch(t, tc.Expected, m.GetMapKeys(tc.Path))
		})
	}
}

type element struct {
	FieldName string
	TagStr    string
	Type      reflect.Type
}

func parsePath(e ...element) []tag.TagMap {
	result := make([]tag.TagMap, 0, len(e))

	for _, el := range e {
		field := reflect.StructField{
			Name: el.FieldName,
			Tag:  reflect.StructTag(el.TagStr),
			Type: el.Type,
		}

		result = append(result, tag.ParseTags(field))
	}

	return result
}
