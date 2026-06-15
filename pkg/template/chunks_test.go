package template

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeChunks unmarshals the JS object literal SelectChunks emits into a set of keys.
func decodeChunks(t *testing.T, js string) map[string]string {
	t.Helper()
	var m map[string]string
	require.NoError(t, json.Unmarshal([]byte(js), &m), "SelectChunks must emit valid JSON")
	return m
}

func TestSelectChunks(t *testing.T) {
	entry := VizbEntryKey
	barRoot := VizbChartRoots["bar"]
	lineRoot := VizbChartRoots["line"]
	pieRoot := VizbChartRoots["pie"]
	heatRoot := VizbChartRoots["heatmap"]
	radarRoot := VizbChartRoots["radar"]
	root3D := VizbChartRoots["3d"]

	require.NotEmpty(t, entry, "generated VizbEntryKey must be present")
	require.NotEmpty(t, barRoot, "generated VizbChartRoots must be populated")
	require.NotEmpty(t, radarRoot, "generated VizbChartRoots must contain radar")

	t.Run("bar only, no 3D drops other renderers and the 3D stack", func(t *testing.T) {
		got := decodeChunks(t, string(SelectChunks([]string{"bar"}, false)))

		assert.Contains(t, got, entry, "entry chunk is always shipped")
		assert.Contains(t, got, barRoot, "selected chart's renderer is kept")
		assert.NotContains(t, got, lineRoot, "unselected line renderer is pruned")
		assert.NotContains(t, got, pieRoot, "unselected pie renderer is pruned")
		assert.NotContains(t, got, heatRoot, "unselected heatmap renderer is pruned")
		assert.NotContains(t, got, root3D, "3D renderer is pruned when needs3D is false")
	})

	t.Run("bar with needs3D keeps the 3D stack", func(t *testing.T) {
		got := decodeChunks(t, string(SelectChunks([]string{"bar"}, true)))

		assert.Contains(t, got, barRoot)
		assert.Contains(t, got, root3D, "3D renderer is kept when needs3D is true")
	})

	t.Run("pie keeps pie, never the 3D stack", func(t *testing.T) {
		// needs3D should never be true for pie in practice, but even if forced the
		// 3D root is only added when 3d is among the enabled roots, which it is via
		// the needs3D flag — so assert the realistic pie-without-3D case.
		got := decodeChunks(t, string(SelectChunks([]string{"pie"}, false)))

		assert.Contains(t, got, pieRoot)
		assert.NotContains(t, got, root3D)
		assert.NotContains(t, got, barRoot)
	})

	t.Run("radar keeps radar, prunes unrelated renderers", func(t *testing.T) {
		got := decodeChunks(t, string(SelectChunks([]string{"radar"}, false)))

		assert.Contains(t, got, radarRoot, "radar renderer is kept")
		assert.NotContains(t, got, barRoot, "unselected bar renderer is pruned")
		assert.NotContains(t, got, lineRoot, "unselected line renderer is pruned")
		assert.NotContains(t, got, pieRoot, "unselected pie renderer is pruned")
		assert.NotContains(t, got, heatRoot, "unselected heatmap renderer is pruned")
		assert.NotContains(t, got, root3D, "3D renderer is pruned when needs3D is false")
	})

	t.Run("empty selection ships all renderers", func(t *testing.T) {
		got := decodeChunks(t, string(SelectChunks(nil, false)))

		for name, root := range VizbChartRoots {
			if name == "3d" {
				continue // 3d only via needs3D
			}
			assert.Contains(t, got, root, "default selection keeps %s", name)
		}
	})

	t.Run("kept chunks never reference a pruned chunk", func(t *testing.T) {
		got := decodeChunks(t, string(SelectChunks([]string{"bar"}, false)))
		for key := range got {
			for _, dep := range VizbChunkImports[key] {
				// A pruned dep is only allowed if it's a gated renderer root.
				if _, kept := got[dep]; kept {
					continue
				}
				_, gated := chartRootSet()[dep]
				assert.True(t, gated, "kept chunk %s references pruned non-root %s", key, dep)
			}
		}
	})
}

// chartRootSet is a test helper mirroring the pruner's gated set.
func chartRootSet() map[string]bool {
	s := make(map[string]bool, len(VizbChartRoots))
	for _, k := range VizbChartRoots {
		s[k] = true
	}
	return s
}
