package shared_test

import (
	"testing"

	barchart "github.com/goptics/vizb/internal/charts/bar"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

type statChartStub struct {
	enabled bool
	math    []string
}

func (statChartStub) ChartType() string    { return "stub" }
func (s statChartStub) StatEnabled() bool  { return s.enabled }
func (s statChartStub) StatMath() []string { return s.math }
func (statChartStub) SwapString() string   { return "" }

func TestMaterialiseStatFlags(t *testing.T) {
	assert.Nil(t, shared.MaterialiseStatFlags(nil))
	got := shared.MaterialiseStatFlags([]string{})
	assert.NotNil(t, got)
	assert.True(t, got.Enabled)
	assert.Empty(t, got.Math)
	got = shared.MaterialiseStatFlags([]string{"counts", "center"})
	assert.Equal(t, []string{"counts", "center"}, got.Math)
}

func TestStatNeedsCorrelation(t *testing.T) {
	assert.True(t, shared.StatNeedsCorrelation(nil))
	assert.True(t, shared.StatNeedsCorrelation([]string{}))
	assert.True(t, shared.StatNeedsCorrelation([]string{"correlations"}))
	assert.False(t, shared.StatNeedsCorrelation([]string{"counts"}))
}

func TestStatConfigNeedsCorrelation(t *testing.T) {
	assert.False(t, (*shared.StatConfig)(nil).NeedsCorrelation())
	assert.False(t, (&shared.StatConfig{Enabled: false, Math: []string{"correlations"}}).NeedsCorrelation())
	assert.True(t, (&shared.StatConfig{Enabled: true, Math: []string{}}).NeedsCorrelation())
	assert.True(t, (&shared.StatConfig{Enabled: true, Math: []string{"correlations"}}).NeedsCorrelation())
	assert.False(t, (&shared.StatConfig{Enabled: true, Math: []string{"counts"}}).NeedsCorrelation())
}

func TestStatConfigStatMath(t *testing.T) {
	assert.Nil(t, (*shared.StatConfig)(nil).StatMath())
	assert.Equal(t, []string{"shape"}, (&shared.StatConfig{Math: []string{"shape"}}).StatMath())
}

func TestChartConfigNeedsCorrelation(t *testing.T) {
	assert.False(t, shared.ChartConfigNeedsCorrelation(statChartStub{}))
	assert.False(t, shared.ChartConfigNeedsCorrelation(statChartStub{enabled: true, math: []string{"counts"}}))
	assert.True(t, shared.ChartConfigNeedsCorrelation(statChartStub{enabled: true, math: []string{"correlations"}}))

	bar := &barchart.Config{
		Type: "bar",
		Stat: &shared.StatConfig{Enabled: true, Math: []string{"counts"}},
	}
	assert.False(t, shared.ChartConfigNeedsCorrelation(bar))
	bar.Stat.Math = []string{"correlations"}
	assert.True(t, shared.ChartConfigNeedsCorrelation(bar))
}
