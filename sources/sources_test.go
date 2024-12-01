package sources

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToMap(t *testing.T) {
	tt := []struct {
		name     string
		env      []string
		expected map[string]string
	}{
		{
			name:     "empty slice",
			env:      []string{},
			expected: map[string]string{},
		},
		{
			name: "single valid entry",
			env:  []string{"KEY=value"},
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name: "multiple valid entries",
			env:  []string{"KEY1=value1", "KEY2=value2"},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "entry with multiple equals signs",
			env:  []string{"KEY=value=with=equals"},
			expected: map[string]string{
				"KEY": "value=with=equals",
			},
		},
		{
			name:     "entry without equals sign",
			env:      []string{"INVALID"},
			expected: map[string]string{},
		},
		{
			name: "mixed valid and invalid entries",
			env:  []string{"VALID=value", "INVALID", "ANOTHER=valid"},
			expected: map[string]string{
				"VALID":   "value",
				"ANOTHER": "valid",
			},
		},
		{
			name: "empty value",
			env:  []string{"KEY="},
			expected: map[string]string{
				"KEY": "",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := ToMap(tc.env)
			assert.Equal(t, tc.expected, result)
		})
	}
}
