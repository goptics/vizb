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

// TestBorderRadiusConfig tests the config helper methods
func TestBorderRadiusConfig(t *testing.T) {
	tests := []struct {
		name            string
		radius          *int
		stack           *bool
		horizontal      *bool
		shouldApply     bool
		expectedRadius  int
		isStacked       bool
		isHorizontal    bool
		expectedCorners [4]int
	}{
		{
			name:            "Nil radius",
			radius:          nil,
			stack:           boolPtr(false),
			horizontal:      boolPtr(false),
			shouldApply:     false,
			expectedRadius:  0,
			isStacked:       false,
			isHorizontal:    false,
			expectedCorners: [4]int{0, 0, 0, 0},
		},
		{
			name:            "Zero radius",
			radius:          intPtr(0),
			stack:           boolPtr(false),
			horizontal:      boolPtr(false),
			shouldApply:     false,
			expectedRadius:  0,
			isStacked:       false,
			isHorizontal:    false,
			expectedCorners: [4]int{0, 0, 0, 0},
		},
		{
			name:            "Positive radius vertical",
			radius:          intPtr(8),
			stack:           boolPtr(false),
			horizontal:      boolPtr(false),
			shouldApply:     true,
			expectedRadius:  8,
			isStacked:       false,
			isHorizontal:    false,
			expectedCorners: [4]int{8, 8, 0, 0}, // top corners
		},
		{
			name:            "Positive radius horizontal",
			radius:          intPtr(8),
			stack:           boolPtr(false),
			horizontal:      boolPtr(true),
			shouldApply:     true,
			expectedRadius:  8,
			isStacked:       false,
			isHorizontal:    true,
			expectedCorners: [4]int{0, 8, 8, 0}, // right corners
		},
		{
			name:            "Stacked with radius",
			radius:          intPtr(8),
			stack:           boolPtr(true),
			horizontal:      boolPtr(false),
			shouldApply:     true,
			expectedRadius:  8,
			isStacked:       true,
			isHorizontal:    false,
			expectedCorners: [4]int{8, 8, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &barchart.Config{
				Type:         "bar",
				BorderRadius: tt.radius,
				Stack:        tt.stack,
				Horizontal:   tt.horizontal,
			}

			assert.Equal(t, tt.shouldApply, cfg.ShouldApplyBorderRadius())
			assert.Equal(t, tt.expectedRadius, cfg.GetBorderRadius())
			assert.Equal(t, tt.isStacked, cfg.IsStacked())
			assert.Equal(t, tt.isHorizontal, cfg.IsHorizontal())

			// Test GetCornerValues for top series
			if tt.shouldApply {
				assert.Equal(t, tt.expectedCorners, cfg.GetCornerValues())
			}
		})
	}
}

// TestBorderRadiusSeriesLogic tests the series-specific corner logic
func TestBorderRadiusSeriesLogic(t *testing.T) {
	tests := []struct {
		name            string
		stack           bool
		isTopSeries     bool
		expectedCorners [4]int
	}{
		{
			name:            "Non-stacked - all get radius",
			stack:           false,
			isTopSeries:     false,
			expectedCorners: [4]int{8, 8, 0, 0},
		},
		{
			name:            "Stacked - top series gets radius",
			stack:           true,
			isTopSeries:     true,
			expectedCorners: [4]int{8, 8, 0, 0},
		},
		{
			name:            "Stacked - lower series no radius",
			stack:           true,
			isTopSeries:     false,
			expectedCorners: [4]int{0, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &barchart.Config{
				Type:         "bar",
				BorderRadius: intPtr(8),
				Stack:        &tt.stack,
				Horizontal:   boolPtr(false),
			}

			corners := cfg.GetCornerValuesForSeries(tt.isTopSeries)
			assert.Equal(t, tt.expectedCorners, corners)
		})
	}
}

// TestBorderRadiusSeriesLogicHorizontal tests horizontal stacked behavior
func TestBorderRadiusSeriesLogicHorizontal(t *testing.T) {
	tests := []struct {
		name            string
		stack           bool
		isTopSeries     bool
		expectedCorners [4]int
	}{
		{
			name:            "Horizontal non-stacked - all get radius",
			stack:           false,
			isTopSeries:     false,
			expectedCorners: [4]int{0, 8, 8, 0}, // right corners
		},
		{
			name:            "Horizontal stacked - top series gets radius",
			stack:           true,
			isTopSeries:     true,
			expectedCorners: [4]int{0, 8, 8, 0}, // right corners
		},
		{
			name:            "Horizontal stacked - lower series no radius",
			stack:           true,
			isTopSeries:     false,
			expectedCorners: [4]int{0, 0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &barchart.Config{
				Type:         "bar",
				BorderRadius: intPtr(8),
				Stack:        &tt.stack,
				Horizontal:   boolPtr(true),
			}

			corners := cfg.GetCornerValuesForSeries(tt.isTopSeries)
			assert.Equal(t, tt.expectedCorners, corners)
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
	assert.False(t, barCfg.ShouldApplyBorderRadius())
}

// TestConfigChartType tests the ChartType method
func TestConfigChartType(t *testing.T) {
	cfg := &barchart.Config{Type: "bar"}
	assert.Equal(t, "bar", cfg.ChartType())
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
