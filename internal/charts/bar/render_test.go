// internal/charts/bar/render_test.go
package bar

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestRenderAppliesBorderRadius verifies that the renderer applies
// borderRadius to the ECharts series configuration
func TestRenderAppliesBorderRadius(t *testing.T) {
    // This test depends on your actual rendering implementation
    // Here's a template based on what your renderer likely looks like:
    
    cfg := &Config{
        Type:         "bar",
        BorderRadius: intPtr(8),
        Stack:        boolPtr(false),
        Horizontal:   boolPtr(false),
    }
    
    // Assuming you have a Render method or function
    // result, err := RenderBarChart(cfg, mockData)
    // require.NoError(t, err)
    
    // Verify the result contains borderRadius in the right places
    // This depends on your actual chart output format
    
    // For example, if you generate ECharts options:
    // options := result.(*echarts.Options)
    // series := options.Series
    
    // For non-stacked: all series should have borderRadius
    // For stacked: only last series should have borderRadius
    
    t.Skip("Implement based on actual renderer")
}

func intPtr(i int) *int { return &i }
func boolPtr(b bool) *bool { return &b }