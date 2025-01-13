package walker

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"

	errs "github.com/sethpollack/envcfg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalk(t *testing.T) {
	tt := map[string]struct {
		env         map[string]string
		cfg         any
		expected    any
		expectedErr error
		skipErrIs   bool
		skip        bool
		skipReason  string
	}{
		"error on non pointer": {
			env: map[string]string{
				"VALUE": "value",
			},
			cfg: struct {
				Value string
			}{},
			expectedErr: errs.ErrNotAPointer,
		},
		"error on non-struct pointer": {
			env: map[string]string{
				"VALUE": "value",
			},
			cfg:         new(string),
			expectedErr: errs.ErrNotAPointer,
		},
		"skip unexported fields": {
			env: map[string]string{
				"VALUE": "value",
			},
			expected: struct {
				value string
			}{},
		},
		"skip ignored fields": {
			env: map[string]string{
				"VALUE": "value",
			},
			expected: struct {
				Option1 string `env:"-"`
				Option2 string `ignore:"true"`
				Option3 string `env:",ignore"`
			}{},
		},
		"ignore nil pointers with no values": {
			env: map[string]string{},
			expected: struct {
				Value *string
			}{},
		},
		"allow default values only on non-struct nil pointers": {
			env: map[string]string{},
			expected: struct {
				Struct *struct {
					Value string `default:"default"`
				}
			}{},
		},
		"allow default values on non-struct nil pointers": {
			env: map[string]string{},
			expected: struct {
				Value *string `default:"default"`
			}{Value: ptr("default")},
		},
		"init any ignores nil pointers with no values": {
			env: map[string]string{},
			expected: struct {
				Value *string `init:"any"`
			}{},
		},
		"init any always allows default values": {
			env: map[string]string{},
			expected: struct {
				Struct *struct {
					Value string `default:"default"`
				} `init:"any"`
				Value *string `default:"default"`
			}{
				Struct: &struct {
					Value string `default:"default"`
				}{
					Value: "default",
				},
				Value: ptr("default"),
			},
		},
		"init always always initializes": {
			env: map[string]string{},
			expected: struct {
				Struct *struct {
					Value string
				} `init:"always"`
				Value *string `init:"always"`
			}{
				Struct: &struct {
					Value string
				}{
					Value: "",
				},
				Value: ptr(""),
			},
		},
		"init never always ignores": {
			env: map[string]string{
				"VALUE":        "value",
				"STRUCT_VALUE": "value",
			},
			expected: struct {
				Struct *struct {
					Value string
				} `init:"never"`
				Value *string `init:"never"`
			}{},
		},
		"nil pointer error": {
			env: map[string]string{
				"VALUE": "value",
			},
			cfg:         &struct{ Value *int }{},
			expectedErr: strconv.ErrSyntax,
		},
		"map nil pointer": {
			env: map[string]string{
				"FIELD_KEY1": "value1",
			},
			expected: struct{ Field *map[string]string }{
				Field: &map[string]string{"key1": "value1"},
			},
		},
		"map of structs nil pointer": {
			skip:       true,
			skipReason: "TODO: fix this",
			env: map[string]string{
				"FIELD_KEY1_VALUE": "value1",
			},
			expected: struct {
				Field *map[string]struct{ Value string }
			}{
				Field: &map[string]struct{ Value string }{
					"key1": {Value: "value1"},
				},
			},
		},
		"delimited slice nil pointer": {
			env: map[string]string{
				"FIELD": "a,b,c",
			},
			expected: struct{ Field *[]string }{Field: &[]string{"a", "b", "c"}},
		},
		"slice nil pointer": {
			env: map[string]string{
				"FIELD_0": "a",
				"FIELD_1": "b",
				"FIELD_2": "c",
			},
			expected: struct{ Field *[]string }{Field: &[]string{"a", "b", "c"}},
		},
		"slice of structs nil pointer": {
			env: map[string]string{
				"FIELD_0_VALUE": "value1",
				"FIELD_1_VALUE": "value2",
			},
			expected: struct{ Field *[]struct{ Value string } }{Field: &[]struct{ Value string }{{Value: "value1"}, {Value: "value2"}}},
		},
		"non nil pointer": {
			env: map[string]string{
				"VALUE": "override",
			},
			cfg:      &struct{ Value *string }{Value: ptr("value")},
			expected: struct{ Value *string }{Value: ptr("override")},
		},
		"required error": {
			env: map[string]string{},
			cfg: &struct {
				Value string `required:"true"`
			}{},
			expectedErr: errs.ErrRequired,
		},
		"not empty error": {
			env: map[string]string{
				"VALUE": "",
			},
			cfg: &struct {
				Value string `notempty:"true"`
			}{},
			expectedErr: errs.ErrNotEmpty,
		},
		"type parser": {
			env: map[string]string{
				"VALUE": "1s",
			},
			expected: struct {
				Value time.Duration
			}{Value: time.Second},
		},
		"type parser with default": {
			env: map[string]string{},
			expected: struct {
				Value time.Duration `default:"1s"`
			}{Value: time.Second},
		},
		"type parser with error": {
			env: map[string]string{
				"VALUE": "invalid",
			},
			cfg: &struct {
				Value time.Duration
			}{},
			expectedErr: assert.AnError,
			skipErrIs:   true,
		},
		"deeply nested structs": {
			env: map[string]string{
				"FIELD_FIELD_VALUE": "value",
			},
			expected: struct {
				Field struct {
					Field struct {
						Value string
					}
				}
			}{
				Field: struct {
					Field struct {
						Value string
					}
				}{
					Field: struct {
						Value string
					}{
						Value: "value",
					},
				},
			},
		},
		"delimited slice": {
			env: map[string]string{
				"SLICE": "a,b,c",
			},
			expected: struct{ Slice []string }{Slice: []string{"a", "b", "c"}},
		},
		"empty delimited slice": {
			env: map[string]string{
				"SLICE": "",
			},
			expected: struct{ Slice []string }{},
		},
		"delimited slice with invalid value": {
			env: map[string]string{
				"SLICE": "a,b,c,",
			},
			cfg:         &struct{ Slice []int }{},
			expectedErr: strconv.ErrSyntax,
		},
		"delimited map": {
			env: map[string]string{
				"MAP": "a:b,c:d",
			},
			expected: struct{ Map map[string]string }{Map: map[string]string{"a": "b", "c": "d"}},
		},
		"delimited map with overrides": {
			env: map[string]string{
				"MAP_OPTION1": "a|b;c|d",
				"MAP_OPTION2": "a|b;c|d",
			},
			expected: struct {
				MapOption1 map[string]string `delim:";" sep:"|"`
				MapOption2 map[string]string `env:",delim=;,sep=|"`
			}{
				MapOption1: map[string]string{"a": "b", "c": "d"},
				MapOption2: map[string]string{"a": "b", "c": "d"},
			},
		},
		"empty delimited map": {
			env: map[string]string{
				"MAP": "",
			},
			expected: struct{ Map map[string]string }{},
		},
		"default delimited map": {
			env: map[string]string{},
			expected: struct {
				Map map[string]string `default:"a:b"`
			}{Map: map[string]string{"a": "b"}},
		},
		"invalid delimited map": {
			env: map[string]string{
				"MAP": "a:b,c",
			},
			cfg:         &struct{ Map map[string]string }{},
			expectedErr: errs.ErrInvalidMapValue,
		},
		"invalid delimited map value": {
			env: map[string]string{
				"MAP": "a:b,c:d",
			},
			cfg:         &struct{ Map map[string]int }{},
			expectedErr: strconv.ErrSyntax,
		},
		"invalid delimited map key": {
			env: map[string]string{
				"MAP": "a:b,c:d",
			},
			cfg:         &struct{ Map map[int]string }{},
			expectedErr: strconv.ErrSyntax,
		},
		"index slice": {
			env: map[string]string{
				"SLICE_0": "a",
				"SLICE_1": "b",
				"SLICE_2": "c",
			},
			expected: struct {
				Slice []string
			}{Slice: []string{"a", "b", "c"}},
		},
		"slice of structs": {
			env: map[string]string{
				"SLICE_0_VALUE": "value1",
				"SLICE_1_VALUE": "value2",
			},
			expected: struct{ Slice []struct{ Value string } }{Slice: []struct{ Value string }{{Value: "value1"}, {Value: "value2"}}},
		},
		"slice of structs only default values": {
			skip:       true,
			skipReason: "TODO: fix this",
			env: map[string]string{
				"SLICE_0_FOO": "", // force traversal, but no matching keys
			},
			expected: struct {
				Slice []struct {
					Value string `default:"default"`
				}
			}{},
		},
		"nil struct with slice of structs with only default values": {
			env: map[string]string{
				"FIELD_SLICE_0_FOO": "", // force traversal, but no matching keys
			},
			expected: struct {
				Field *struct {
					Slice []struct {
						Value string `default:"default"`
					}
				}
			}{},
		},
		"slice with invalid value": {
			env: map[string]string{
				"SLICE_0": "a",
				"SLICE_1": "b",
				"SLICE_2": "c",
			},
			cfg:         &struct{ Slice []int }{},
			expectedErr: strconv.ErrSyntax,
		},
		"flat map": {
			env: map[string]string{
				"MAP_KEY1": "value1",
				"MAP_KEY2": "value2",
			},
			expected: struct{ Map map[string]string }{Map: map[string]string{"key1": "value1", "key2": "value2"}},
		},
		"flat map with invalid value": {
			env: map[string]string{
				"MAP_KEY1": "value1",
				"MAP_KEY2": "value2",
			},
			cfg:         &struct{ Map map[string]int }{},
			expectedErr: strconv.ErrSyntax,
		},
		"flat map with invalid key": {
			env: map[string]string{
				"MAP_KEY1": "value1",
				"MAP_KEY2": "value2",
			},
			cfg:         &struct{ Map map[int]string }{},
			expectedErr: strconv.ErrSyntax,
		},
		"map of structs": {
			env: map[string]string{
				"MAP_KEY1_VALUE": "value1",
				"MAP_KEY2_VALUE": "value2",
			},
			expected: struct {
				Map map[string]struct{ Value string }
			}{Map: map[string]struct{ Value string }{"key1": {Value: "value1"}, "key2": {Value: "value2"}}},
		},
		"map of structs with only default values": {
			env: map[string]string{
				"MAP_KEY1_FOO": "", // force traversal, but no matching keys
			},
			expected: struct {
				Map map[string]struct {
					Value string `default:"default"`
				}
			}{},
		},
		"decoder": {
			env: map[string]string{
				"UNMARSHALER": "hello world!",
			},
			expected: struct {
				Unmarshaler unmarshaler
			}{
				Unmarshaler: unmarshaler{
					Value: "hello world!",
				},
			},
		},
		"decoder with default value": {
			env: map[string]string{},
			expected: struct {
				Unmarshaler unmarshaler `default:"default"`
			}{
				Unmarshaler: unmarshaler{
					Value: "default",
				},
			},
		},
		"decoder error": {
			env: map[string]string{
				"UNMARSHALER": "hello world!",
			},
			cfg:         &struct{ Unmarshaler unmarshalError }{},
			expectedErr: unmarshalErr,
		},
		"decoder unset": {
			env: map[string]string{},
			expected: struct {
				Unmarshaler unset
			}{},
		},
		"decoder unset with default value": {
			env: map[string]string{},
			expected: struct {
				Unmarshaler unset `default:"default"`
			}{
				Unmarshaler: unset{
					Value: "Hello World!",
				},
			},
		},
		"decodeunset": {
			env: map[string]string{},
			expected: struct {
				Unmarshaler unset `decodeunset:"true"`
			}{
				Unmarshaler: unset{
					Value: "Hello World!",
				},
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			if tc.skip {
				t.Skip(tc.skipReason)
			}

			w := New()

			w.Matcher.EnvVars = tc.env

			cfg := tc.cfg
			if cfg == nil {
				cfg = reflect.New(reflect.TypeOf(tc.expected)).Interface()
			}

			err := w.Walk(cfg)

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

type unmarshaler struct {
	Value string
}

func (u *unmarshaler) UnmarshalText(text []byte) error {
	u.Value = string(text)
	return nil
}

var unmarshalErr = errors.New("error")

type unmarshalError struct {
	Value string
}

func (d *unmarshalError) UnmarshalText(text []byte) error {
	return unmarshalErr
}
