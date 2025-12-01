package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateStatType(t *testing.T) {
	tests := []struct {
		name     string
		statName string
		unit     string
		per      string
		expected string
	}{
		{
			name:     "All fields present",
			statName: "Throughput",
			unit:     "MB",
			per:      "s",
			expected: "Throughput (MB/s)",
		},
		{
			name:     "Name and Unit present",
			statName: "Memory",
			unit:     "KB",
			per:      "",
			expected: "Memory (KB)",
		},
		{
			name:     "Name and Per present",
			statName: "Operations",
			unit:     "",
			per:      "op",
			expected: "Operations/op",
		},
		{
			name:     "Only Name present",
			statName: "Count",
			unit:     "",
			per:      "",
			expected: "Count",
		},
		{
			name:     "Empty Name",
			statName: "",
			unit:     "MB",
			per:      "s",
			expected: " (MB/s)",
		},
		{
			name:     "All empty",
			statName: "",
			unit:     "",
			per:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateStatType(tt.statName, tt.unit, tt.per)
			assert.Equal(t, tt.expected, result, "CreateStatType(%s, %s, %s) should equal %s", tt.statName, tt.unit, tt.per, tt.expected)
		})
	}
}
