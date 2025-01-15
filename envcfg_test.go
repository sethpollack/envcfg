package envcfg_test

import (
	"errors"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/sethpollack/envcfg"
	errs "github.com/sethpollack/envcfg/errors"
	"github.com/sethpollack/envcfg/sources/osenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	tempFile, err := os.CreateTemp("", "env.txt")
	if err != nil {
		t.Fatal(err)
	}

	_, err = tempFile.WriteString("${OTHER_VAR}")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	tempDotEnvFile, err := os.CreateTemp("", ".env")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDotEnvFile.Name())

	_, err = tempDotEnvFile.WriteString("FIELD=value")
	if err != nil {
		t.Fatal(err)
	}

	tt := map[string]struct {
		env      map[string]string
		cfg      any
		expected any

		options []envcfg.Option

		expectedErr error

		skipErrIs  bool
		skip       bool
		skipReason string
	}{
		"WithTagName": {
			env:     map[string]string{"CUSTOM": "value"},
			options: []envcfg.Option{envcfg.WithTagName("foo")},
			expected: struct {
				Field string `foo:"CUSTOM"`
			}{
				Field: "value",
			},
		},
		"WithDelimiterTag": {
			env:     map[string]string{"FIELD": "key|value"},
			options: []envcfg.Option{envcfg.WithDelimiterTag("custom_delim")},
			expected: struct {
				Field []string `custom_delim:"|"`
			}{
				Field: []string{"key", "value"},
			},
		},
		"WithDelimiter": {
			env:     map[string]string{"FIELD": "key|value"},
			options: []envcfg.Option{envcfg.WithDelimiter("|")},
			expected: struct {
				Field []string
			}{
				Field: []string{"key", "value"},
			},
		},
		"WithSeparatorTag": {
			env:     map[string]string{"FIELD": "key1|value1,key2|value2"},
			options: []envcfg.Option{envcfg.WithSeparatorTag("custom_sep")},
			expected: struct {
				Field map[string]string `custom_sep:"|"`
			}{
				Field: map[string]string{"key1": "value1", "key2": "value2"},
			},
		},
		"WithSeparator": {
			env:     map[string]string{"FIELD": "key|value"},
			options: []envcfg.Option{envcfg.WithSeparator("|")},
			expected: struct {
				Field map[string]string
			}{
				Field: map[string]string{"key": "value"},
			},
		},
		"WithDecodeUnsetTag": {
			options: []envcfg.Option{envcfg.WithDecodeUnsetTag("custom_decodeunset")},
			expected: struct {
				Field unset `custom_decodeunset:"true"`
			}{
				Field: unset{
					Value: "Hello World!",
				},
			},
		},
		"WithDecodeUnset": {
			options: []envcfg.Option{envcfg.WithDecodeUnset()},
			expected: struct {
				Field unset
			}{
				Field: unset{
					Value: "Hello World!",
				},
			},
		},
		"WithInitTag": {
			options: []envcfg.Option{envcfg.WithInitTag("custom_init")},
			expected: struct {
				Field *string `custom_init:"always"`
			}{
				Field: ptr(""),
			},
		},
		"WithInitAny": {
			options: []envcfg.Option{envcfg.WithInitAny()},
			expected: struct {
				Field *string `default:"value"`
			}{
				Field: ptr("value"),
			},
		},
		"WithInitNever": {
			env:     map[string]string{"FIELD": "value"},
			options: []envcfg.Option{envcfg.WithInitNever()},
			expected: struct {
				Field *string `default:"value"`
			}{
				Field: nil,
			},
		},
		"WithInitAlways": {
			options: []envcfg.Option{envcfg.WithInitAlways()},
			expected: struct {
				Field *string
			}{
				Field: ptr(""),
			},
		},
		"WithDefaultTag": {
			options: []envcfg.Option{envcfg.WithDefaultTag("custom_default")},
			expected: struct {
				Field string `custom_default:"value"`
			}{
				Field: "value",
			},
		},
		"WithExpandTag": {
			env:     map[string]string{"FIELD": "${FOO}", "FOO": "value"},
			options: []envcfg.Option{envcfg.WithExpandTag("custom_expand")},
			expected: struct {
				Field string `custom_expand:"true"`
			}{
				Field: "value",
			},
		},
		"WithFileTag": {
			env:     map[string]string{"FIELD": tempFile.Name()},
			options: []envcfg.Option{envcfg.WithFileTag("custom_file")},
			expected: struct {
				Field string `custom_file:"true"`
			}{
				Field: "${OTHER_VAR}",
			},
		},
		"WithNotEmptyTag": {
			env:     map[string]string{"FIELD": ""},
			options: []envcfg.Option{envcfg.WithNotEmptyTag("custom_notempty")},
			expected: struct {
				Field string `custom_notempty:"true"`
			}{},
			expectedErr: errs.ErrNotEmpty,
		},
		"WithNotEmpty": {
			env:     map[string]string{"FIELD": ""},
			options: []envcfg.Option{envcfg.WithNotEmpty()},
			expected: struct {
				Field string
			}{},
			expectedErr: errs.ErrNotEmpty,
		},
		"WithExpand": {
			env:     map[string]string{"FIELD": "${FOO}", "FOO": "value"},
			options: []envcfg.Option{envcfg.WithExpand()},
			expected: struct {
				Field string
			}{
				Field: "value",
			},
		},
		"WithRequiredTag": {
			options: []envcfg.Option{envcfg.WithRequiredTag("custom_required")},
			expected: struct {
				Field string `custom_required:"true"`
			}{},
			expectedErr: errs.ErrRequired,
		},
		"WithRequired": {
			options: []envcfg.Option{envcfg.WithRequired()},
			expected: struct {
				Field string
			}{},
			expectedErr: errs.ErrRequired,
		},
		"WithDisableFallback": {
			env:     map[string]string{"FIELD": "value"},
			options: []envcfg.Option{envcfg.WithDisableFallback()},
			expected: struct {
				Field string
			}{},
		},
		"WithDecoder": {
			env: map[string]string{"FIELD": "hello"},
			options: []envcfg.Option{envcfg.WithDecoder((*customIface)(nil), func(v any, value string) error {
				return v.(*custom).CustomDecode(value)
			})},
			expected: struct {
				Field custom
			}{
				Field: custom{field: "hello world!"},
			},
		},
		"WithTypeParser": {
			env: map[string]string{"FIELD": "value"},
			options: []envcfg.Option{envcfg.WithTypeParser(reflect.TypeOf((*Inter)(nil)).Elem(), func(value string) (any, error) {
				return &Impl{Field: value}, nil
			})},
			expected: struct {
				Field Inter
			}{
				Field: &Impl{Field: "value"},
			},
		},
		"WithTypeParsers": {
			env: map[string]string{"FIELD": "value"},
			options: []envcfg.Option{envcfg.WithTypeParsers(map[reflect.Type]func(value string) (any, error){
				reflect.TypeOf((*Inter)(nil)).Elem(): func(value string) (any, error) {
					return &Impl{Field: value}, nil
				},
			})},
			expected: struct {
				Field Inter
			}{
				Field: &Impl{Field: "value"},
			},
		},
		"WithKindParser": {
			env: map[string]string{"FIELD": "hello"},
			options: []envcfg.Option{envcfg.WithKindParser(reflect.String, func(value string) (any, error) {
				return value + " world", nil
			})},
			expected: struct {
				Field string
			}{
				Field: "hello world",
			},
		},
		"WithKindParsers": {
			env: map[string]string{"FIELD": "hello"},
			options: []envcfg.Option{envcfg.WithKindParsers(map[reflect.Kind]func(value string) (any, error){
				reflect.String: func(value string) (any, error) {
					return value + " world", nil
				},
			})},
			expected: struct {
				Field string
			}{
				Field: "hello world",
			},
		},
		"WithLoader": {
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithMapEnvSource(map[string]string{"FIELD": "value"}),
			)},
			expected: struct {
				Field string
			}{
				Field: "value",
			},
		},
		"WithLoader Error": {
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithSource(&customSource{}),
			)},
			expected: struct {
				Field string
			}{},
			expectedErr: errs.ErrLoadEnv,
		},
		"WithSources": {
			env: map[string]string{"FIELD": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithSources(osenv.New()),
			)},
			expected: struct {
				Field string
			}{
				Field: "value",
			},
		},
		"WithFilter": {
			env: map[string]string{"FIELD": "value", "OTHER": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithFilter(func(key string) bool {
					return key == "FIELD"
				}),
			)},
			expected: struct {
				Field string
				Other string
			}{
				Field: "value",
				Other: "",
			},
		},
		"WithTransform": {
			env: map[string]string{"FIELD": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithTransform(func(key string) string {
					return "TRANSFORMED_" + key
				}),
			)},
			expected: struct {
				TransformedField string
			}{
				TransformedField: "value",
			},
		},
		"WithPrefix": {
			env: map[string]string{"PREFIXED_FIELD": "value", "OTHER": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithPrefix("PREFIXED_"),
			)},
			expected: struct {
				Field string
				Other string
			}{
				Field: "value",
				Other: "",
			},
		},
		"WithSuffix": {
			env: map[string]string{"FIELD_SUFFIX": "value", "OTHER": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithSuffix("_SUFFIX"),
			)},
			expected: struct {
				Field string
				Other string
			}{
				Field: "value",
				Other: "",
			},
		},
		"WithHasPrefix": {
			env: map[string]string{"PREFIXED_FIELD": "value", "OTHER": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithHasPrefix("PREFIXED_"),
			)},
			expected: struct {
				PrefixedField string
				Other         string
			}{
				PrefixedField: "value",
				Other:         "",
			},
		},
		"WithHasSuffix": {
			env: map[string]string{"FIELD_SUFFIX": "value", "OTHER": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithHasSuffix("_SUFFIX"),
			)},
			expected: struct {
				FieldSuffix string
				Other       string
			}{
				FieldSuffix: "value",
				Other:       "",
			},
		},
		"WithHasMatch": {
			env: map[string]string{"MATCHED_FIELD": "value", "OTHER": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithHasMatch(regexp.MustCompile("MATCHED_")),
			)},
			expected: struct {
				MatchedField string
				Other        string
			}{
				MatchedField: "value",
				Other:        "",
			},
		},
		"WithKeys": {
			env: map[string]string{"KEY1": "key1", "KEY2": "key2", "KEY3": "key3", "KEY4": "key4"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithKeys("KEY1", "KEY2"),
			)},
			expected: struct {
				Key1 string
				Key2 string
				Key3 string
				Key4 string
			}{
				Key1: "key1",
				Key2: "key2",
				Key3: "",
				Key4: "",
			},
		},
		"WithTrimPrefix": {
			env: map[string]string{"PREFIXED_FIELD": "hello", "OTHER": "123"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithTrimPrefix("PREFIXED_"),
			)},
			expected: struct {
				Field string
				Other string
			}{
				Field: "hello",
				Other: "123",
			},
		},
		"WithTrimSuffix": {
			env: map[string]string{"FIELD_SUFFIX": "hello", "OTHER": "123"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithTrimSuffix("_SUFFIX"),
			)},
			expected: struct {
				Field string
				Other string
			}{
				Field: "hello",
				Other: "123",
			},
		},
		"WithMapEnvSource": {
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithMapEnvSource(map[string]string{"FIELD": "value"}),
			)},
			expected: struct {
				Field string
			}{
				Field: "value",
			},
		},
		"WithOSEnvSource": {
			env: map[string]string{"FIELD": "value"},
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithOSEnvSource(),
			)},
			expected: struct {
				Field string
			}{
				Field: "value",
			},
		},
		"WithDotEnvSource": {
			options: []envcfg.Option{envcfg.WithLoader(
				envcfg.WithDotEnvSource(tempDotEnvFile.Name()),
			)},
			expected: struct {
				Field string
			}{
				Field: "value",
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.skip {
				t.Skip(tc.skipReason)
			}

			for k, v := range tc.env {
				t.Setenv(k, v)
			}

			cfg := tc.cfg
			if cfg == nil {
				cfg = reflect.New(reflect.TypeOf(tc.expected)).Interface()
			}

			err := envcfg.Parse(cfg, tc.options...)

			if tc.expectedErr != nil {
				require.Error(t, err)
				if !tc.skipErrIs {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				require.NoError(t, err)
				actual := reflect.ValueOf(cfg).Elem().Interface()
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestParseAs(t *testing.T) {
	type Config struct {
		Field string `required:"true"`
	}

	t.Run("success", func(t *testing.T) {
		t.Setenv("FIELD", "value")

		cfg, err := envcfg.ParseAs[Config]()

		require.NoError(t, err)
		assert.Equal(t, Config{Field: "value"}, cfg)
	})

	t.Run("error", func(t *testing.T) {
		cfg, err := envcfg.ParseAs[Config]()

		assert.ErrorIs(t, err, errs.ErrRequired)
		assert.Equal(t, Config{}, cfg)
	})
}

func TestMustParse(t *testing.T) {
	type Config struct {
		Field string `required:"true"`
	}
	t.Run("success", func(t *testing.T) {
		t.Setenv("FIELD", "value")

		cfg := Config{}

		assert.NotPanics(t, func() {
			envcfg.MustParse(&cfg)
		})

		assert.Equal(t, Config{Field: "value"}, cfg)
	})

	t.Run("error", func(t *testing.T) {
		cfg := Config{}

		assert.Panics(t, func() {
			envcfg.MustParse(&cfg)
		})
	})
}

func TestMustParseAs(t *testing.T) {
	type Config struct {
		Field string `required:"true"`
	}
	t.Run("success", func(t *testing.T) {
		t.Setenv("FIELD", "value")

		var cfg Config

		assert.NotPanics(t, func() {
			cfg = envcfg.MustParseAs[Config]()
		})

		assert.Equal(t, Config{Field: "value"}, cfg)
	})

	t.Run("error", func(t *testing.T) {
		assert.Panics(t, func() {
			envcfg.MustParseAs[Config]()
		})
	})
}

func ptr[T any](v T) *T {
	return &v
}

type unset struct {
	Value string
}

func (u *unset) UnmarshalText(text []byte) error {
	u.Value = "Hello World!"
	return nil
}

type customIface interface {
	CustomDecode(value string) error
}

type custom struct {
	field string
}

func (c *custom) CustomDecode(value string) error {
	c.field = value + " world!"
	return nil
}

type Inter interface{}

type Impl struct {
	Field string
}

type customSource struct {
}

func (c *customSource) Load() (map[string]string, error) {
	return nil, errors.New("source error")
}
