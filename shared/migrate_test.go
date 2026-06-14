package shared

import (
	"encoding/json"
	"testing"
)

func TestMigrateDataset(t *testing.T) {
	t.Run("migrates legacy axisLabels to Settings.Axes", func(t *testing.T) {
		raw := []byte(`{"name":"test","axisLabels":{"x":"Size","y":"ns/op"},"settings":{"charts":["bar"],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"},"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(raw, &ds); err != nil {
			t.Fatal(err)
		}
		// Before migration: Axes is empty (old field not in Dataset struct)
		if len(ds.Settings.Axes) != 0 {
			t.Fatal("expected empty Axes before migration")
		}
		MigrateDataset(&ds, raw)
		// After: Axes populated from legacy axisLabels
		if len(ds.Settings.Axes) != 2 {
			t.Fatalf("expected 2 axes, got %d", len(ds.Settings.Axes))
		}
		if ds.Settings.Axes[0] != (Axis{Key: "x", Label: "Size"}) {
			t.Errorf("axes[0] = %+v", ds.Settings.Axes[0])
		}
		if ds.Settings.Axes[1] != (Axis{Key: "y", Label: "ns/op"}) {
			t.Errorf("axes[1] = %+v", ds.Settings.Axes[1])
		}
	})

	t.Run("skips migration when Settings.Axes already populated", func(t *testing.T) {
		raw := []byte(`{"name":"test","axisLabels":{"x":"Old"},"settings":{"charts":[],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear","axes":[{"key":"x","label":"New"}]},"data":[]}`)
		var ds Dataset
		json.Unmarshal(raw, &ds)
		MigrateDataset(&ds, raw)
		if len(ds.Settings.Axes) != 1 || ds.Settings.Axes[0].Label != "New" {
			t.Errorf("should not overwrite existing axes: %+v", ds.Settings.Axes)
		}
	})

	t.Run("skips migration when rawJSON is nil", func(t *testing.T) {
		ds := &Dataset{}
		MigrateDataset(ds, nil)
		if len(ds.Settings.Axes) != 0 {
			t.Errorf("expected empty axes, got %+v", ds.Settings.Axes)
		}
	})

	t.Run("handles legacy with all four axes in canonical order", func(t *testing.T) {
		raw := []byte(`{"name":"test","axisLabels":{"name":"Impl","x":"Size","y":"ns/op","z":"Workers"},"settings":{"charts":[],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"},"data":[]}`)
		var ds Dataset
		json.Unmarshal(raw, &ds)
		MigrateDataset(&ds, raw)
		want := []Axis{
			{Key: "name", Label: "Impl"},
			{Key: "x", Label: "Size"},
			{Key: "y", Label: "ns/op"},
			{Key: "z", Label: "Workers"},
		}
		if len(ds.Settings.Axes) != 4 {
			t.Fatalf("expected 4 axes, got %d: %+v", len(ds.Settings.Axes), ds.Settings.Axes)
		}
		for i, got := range ds.Settings.Axes {
			if got != want[i] {
				t.Errorf("axes[%d] = %+v, want %+v", i, got, want[i])
			}
		}
	})

	t.Run("no migration when no legacy axisLabels and no axes", func(t *testing.T) {
		raw := []byte(`{"name":"test","settings":{"charts":[],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"},"data":[]}`)
		var ds Dataset
		json.Unmarshal(raw, &ds)
		MigrateDataset(&ds, raw)
		if len(ds.Settings.Axes) != 0 {
			t.Errorf("expected empty axes, got %+v", ds.Settings.Axes)
		}
	})

	// Simulates the parseInputFile array branch in cmd/merge.go: decode via
	// []json.RawMessage so each element's raw bytes carry axisLabels.
	t.Run("migrates element from raw array element bytes", func(t *testing.T) {
		rawElem := []byte(`{"name":"elem","axisLabels":{"x":"N","y":"ns/op"},"settings":{"charts":["bar"],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"},"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(rawElem, &ds); err != nil {
			t.Fatal(err)
		}
		MigrateDataset(&ds, rawElem)
		if len(ds.Settings.Axes) != 2 {
			t.Fatalf("expected 2 axes from per-element bytes, got %d", len(ds.Settings.Axes))
		}
		if ds.Settings.Axes[0] != (Axis{Key: "x", Label: "N"}) {
			t.Errorf("axes[0] = %+v", ds.Settings.Axes[0])
		}
		if ds.Settings.Axes[1] != (Axis{Key: "y", Label: "ns/op"}) {
			t.Errorf("axes[1] = %+v", ds.Settings.Axes[1])
		}
	})
}
