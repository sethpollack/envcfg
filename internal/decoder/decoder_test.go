package decoder

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type CustomDecoderIface interface {
	CustomDecode(value string) error
}

type customDecoder struct {
	value string
}

func (c *customDecoder) CustomDecode(value string) error {
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

func TestFromReflectValue(t *testing.T) {
	tt := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "custom decoder",
			input: &customDecoder{},
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
	}

	for _, tc := range tt {
		r := New()

		err := r.Build(WithDecoder((*CustomDecoderIface)(nil), func(v any, value string) error {
			return v.(*customDecoder).CustomDecode(value)
		}))
		require.NoError(t, err)

		t.Run(tc.name, func(t *testing.T) {
			rv := reflect.ValueOf(tc.input)

			decoder := r.ToDecoder(rv)
			require.NotNil(t, decoder)

			err := decoder.Decode(tc.name)
			require.NoError(t, err)

			switch v := tc.input.(type) {
			case *customDecoder:
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
