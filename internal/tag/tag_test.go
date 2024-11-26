package tag

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTag(t *testing.T) {
	tt := []struct {
		name     string
		input    string
		expected TagMap
	}{
		{
			name:  "no options",
			input: `env:"TEST_FIELD"`,
			expected: TagMap{
				FieldName: "TestField",
				Tags: map[string]Tag{
					"env": {
						Name:    "env",
						Value:   "TEST_FIELD",
						Options: map[string]string{},
					},
					"struct": {
						Name:    "struct",
						Value:   "TestField",
						Options: map[string]string{},
					},
					"struct_snake": {
						Name:    "struct_snake",
						Value:   "test_field",
						Options: map[string]string{},
					},
				},
			},
		},
		{
			name:  "with options",
			input: `env:"TEST,required,min=1,max=10"`,
			expected: TagMap{
				FieldName: "TestField",
				Tags: map[string]Tag{
					"env": {
						Name:  "env",
						Value: "TEST",
						Options: map[string]string{
							"required": "",
							"min":      "1",
							"max":      "10",
						},
					},
					"struct": {
						Name:    "struct",
						Value:   "TestField",
						Options: map[string]string{},
					},
					"struct_snake": {
						Name:    "struct_snake",
						Value:   "test_field",
						Options: map[string]string{},
					},
				},
			},
		},
		{
			name:  "no env tag",
			input: `json:"test_field" toml:"test_field"`,
			expected: TagMap{
				FieldName: "TestField",
				Tags: map[string]Tag{
					"json": {
						Name:    "json",
						Value:   "test_field",
						Options: map[string]string{},
					},
					"toml": {
						Name:    "toml",
						Value:   "test_field",
						Options: map[string]string{},
					},
					"struct": {
						Name:    "struct",
						Value:   "TestField",
						Options: map[string]string{},
					},
					"struct_snake": {
						Name:    "struct_snake",
						Value:   "test_field",
						Options: map[string]string{},
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := ParseTags(reflect.StructField{
				Name: "TestField",
				Tag:  reflect.StructTag(tc.input),
			})
			assert.EqualValues(t, tc.expected, actual)
		})
	}
}
