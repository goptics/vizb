package cli

import (
	"testing"

	"github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBorderRadiusValidation tests the validation function
func TestBorderRadiusValidation(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expectErr bool
	}{
		{"Valid positive integer", "8", false},
		{"Valid zero", "0", false},
		{"Valid large integer", "100", false},
		{"Invalid negative", "-5", true},
		{"Invalid non-integer", "abc", true},
		{"Invalid float", "8.5", true},
		{"Empty string", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := charts.ValidateBorderRadiusValue(tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBarConfigCreation tests the New() function
func TestBarConfigCreation(t *testing.T) {
	cfg := barchart.New()
	require.NotNil(t, cfg)

	barCfg, ok := cfg.(*barchart.Config)
	require.True(t, ok)

	assert.Equal(t, "bar", barCfg.Type)
	assert.Nil(t, barCfg.BorderRadius)
	// ShouldApplyBorderRadius was removed; no assertion needed.
}

// TestConfigChartType tests the ChartType method
func TestConfigChartType(t *testing.T) {
	cfg := &barchart.Config{Type: "bar"}
	assert.Equal(t, "bar", cfg.ChartType())
}