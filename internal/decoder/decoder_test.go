package decoder

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Custom type implementing Decoder interface
type decoder struct {
	value string
}

func (c *decoder) Decode(value string) error {
	c.value = value
	return nil
}

type customIface interface {
	CustomDecode(value string) error
}

// Custom type implementing customIface
type custom struct {
	value string
}

func (c *custom) CustomDecode(value string) error {
	c.value = value
	return nil
}

// Custom type implementing flag.Value interface
type flagValue struct {
	value string
}

func (f *flagValue) String() string {
	return f.value
}

func (f *flagValue) Set(value string) error {
	f.value = value
	return nil
}

// Custom type implementing encoding.TextUnmarshaler
type textUnmarshaler struct {
	value string
}

func (t *textUnmarshaler) UnmarshalText(text []byte) error {
	t.value = string(text)
	return nil
}

// Custom type implementing encoding.BinaryUnmarshaler
type binaryUnmarshaler struct {
	value string
}

func (b *binaryUnmarshaler) UnmarshalBinary(data []byte) error {
	b.value = string(data)
	return nil
}

func TestToDecoder(t *testing.T) {
	tt := []struct {
		name      string
		input     interface{}
		expectNil bool
	}{
		{
			name:      "non decoder",
			input:     []string{},
			expectNil: true,
		},
		{
			name:  "decoder",
			input: &decoder{},
		},
		{
			name:  "custom",
			input: &custom{},
		},
		{
			name:  "flag value",
			input: &flagValue{},
		},
		{
			name:  "text unmarshaler",
			input: &textUnmarshaler{},
		},
		{
			name:  "binary unmarshaler",
			input: &binaryUnmarshaler{},
		},
		{
			name:      "nil",
			input:     nil,
			expectNil: true,
		},
	}

	for _, tc := range tt {
		r := New()

		r.Decoders[(*customIface)(nil)] = func(v any, value string) error {
			return v.(*custom).CustomDecode(value)
		}

		t.Run(tc.name, func(t *testing.T) {
			rv := reflect.ValueOf(tc.input)

			decoder := r.ToDecoder(rv)

			if tc.expectNil {
				assert.Nil(t, decoder)
				return
			}

			require.NotNil(t, decoder)

			err := decoder.Decode(tc.name)
			require.NoError(t, err)

			switch v := tc.input.(type) {
			case *custom:
				assert.Equal(t, tc.name, v.value)
			case *flagValue:
				assert.Equal(t, tc.name, v.value)
			case *textUnmarshaler:
				assert.Equal(t, tc.name, v.value)
			case *binaryUnmarshaler:
				assert.Equal(t, tc.name, v.value)
			}
		})
	}
}
