package luhn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		expected bool
	}{
		{
			name:     "valid number",
			number:   1345821735824,
			expected: true,
		},
		{
			name:     "invalid number",
			number:   1234,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Check(tt.number)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckString(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "valid number",
			number:   "1345821735824",
			expected: true,
		},
		{
			name:     "invalid number",
			number:   "1234",
			expected: false,
		},
		{
			name:     "invalid string",
			number:   "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckString(tt.number)
			assert.Equal(t, tt.expected, result)
		})
	}
}
