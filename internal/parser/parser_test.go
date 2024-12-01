package parser

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKind(t *testing.T) {
	tt := []struct {
		name        string
		kind        reflect.Kind
		value       string
		expected    any
		expectedErr bool
	}{
		{
			name:     "string",
			kind:     reflect.String,
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			kind:     reflect.String,
			value:    "",
			expected: nil,
		},
		{
			name:     "int",
			kind:     reflect.Int,
			value:    "42",
			expected: int(42),
		},
		{
			name:     "empty int string",
			kind:     reflect.Int,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid int",
			kind:        reflect.Int,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "int8",
			kind:     reflect.Int8,
			value:    "42",
			expected: int8(42),
		},
		{
			name:     "empty int8 string",
			kind:     reflect.Int8,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid int8",
			kind:        reflect.Int8,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "int16",
			kind:     reflect.Int16,
			value:    "42",
			expected: int16(42),
		},
		{
			name:     "empty int16 string",
			kind:     reflect.Int16,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid int16",
			kind:        reflect.Int16,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "int32",
			kind:     reflect.Int32,
			value:    "42",
			expected: int32(42),
		},
		{
			name:     "empty int32 string",
			kind:     reflect.Int32,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid int32",
			kind:        reflect.Int32,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "int64",
			kind:     reflect.Int64,
			value:    "42",
			expected: int64(42),
		},
		{
			name:     "empty int64 string",
			kind:     reflect.Int64,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid int64",
			kind:        reflect.Int64,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "uint",
			kind:     reflect.Uint,
			value:    "42",
			expected: uint(42),
		},
		{
			name:     "empty uint string",
			kind:     reflect.Uint,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid uint",
			kind:        reflect.Uint,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "uint8",
			kind:     reflect.Uint8,
			value:    "42",
			expected: uint8(42),
		},
		{
			name:     "empty uint8 string",
			kind:     reflect.Uint8,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid uint8",
			kind:        reflect.Uint8,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "uint16",
			kind:     reflect.Uint16,
			value:    "42",
			expected: uint16(42),
		},
		{
			name:     "empty uint16 string",
			kind:     reflect.Uint16,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid uint16",
			kind:        reflect.Uint16,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "uint32",
			kind:     reflect.Uint32,
			value:    "42",
			expected: uint32(42),
		},
		{
			name:     "empty uint32 string",
			kind:     reflect.Uint32,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid uint32",
			kind:        reflect.Uint32,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "uint64",
			kind:     reflect.Uint64,
			value:    "42",
			expected: uint64(42),
		},
		{
			name:     "empty uint64 string",
			kind:     reflect.Uint64,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid uint64",
			kind:        reflect.Uint64,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "float32",
			kind:     reflect.Float32,
			value:    "3.14",
			expected: float32(3.14),
		},
		{
			name:     "empty float32 string",
			kind:     reflect.Float32,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid float32",
			kind:        reflect.Float32,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "float64",
			kind:     reflect.Float64,
			value:    "3.14",
			expected: float64(3.14),
		},
		{
			name:     "empty float64 string",
			kind:     reflect.Float64,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid float64",
			kind:        reflect.Float64,
			value:       "invalid",
			expectedErr: true,
		},
		{
			name:     "bool",
			kind:     reflect.Bool,
			value:    "true",
			expected: true,
		},
		{
			name:     "empty bool string",
			kind:     reflect.Bool,
			value:    "",
			expected: nil,
		},
		{
			name:        "invalid bool",
			kind:        reflect.Bool,
			value:       "invalid",
			expectedErr: true,
		},
	}

	p := New()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			newValue, _, err := p.ParseKind(tc.kind, tc.value)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, newValue)
			}
		})
	}
}

func TestParseType(t *testing.T) {
	tt := []struct {
		name        string
		typ         reflect.Type
		value       string
		expected    any
		expectedErr bool
	}{
		{
			name:     "duration",
			typ:      reflect.TypeOf(time.Nanosecond),
			value:    "1s",
			expected: time.Second,
		},
		{
			name:     "empty duration",
			typ:      reflect.TypeOf(time.Nanosecond),
			value:    "",
			expected: nil,
		},
	}

	p := New()

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			newValue, _, err := p.ParseType(tc.typ, tc.value)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, newValue)
			}
		})
	}
}

func TestParseTypeWithParser(t *testing.T) {
	type Inter interface{}

	type Impl struct {
		Value string
	}

	var inter Inter

	p := New()
	p.TypeParsers[reflect.TypeOf(&inter).Elem()] = func(value string) (any, error) {
		return &Impl{Value: value}, nil
	}

	newValue, _, err := p.ParseType(reflect.TypeOf(&inter).Elem(), "hello")

	require.NoError(t, err)
	assert.Equal(t, &Impl{Value: "hello"}, newValue)
}
