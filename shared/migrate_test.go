package shared

import (
	"encoding/json"
	"testing"
)

const baseSettings = `"settings":{"charts":["bar"],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"}`

func TestMigrateDataset(t *testing.T) {
	t.Run("migrates legacy flat cpu/os/arch/pkg to Meta", func(t *testing.T) {
		raw := []byte(`{"name":"test","cpu":{"name":"Intel i7","cores":8},"os":"linux","arch":"amd64","pkg":"github.com/foo/bar",` + baseSettings + `,"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(raw, &ds); err != nil {
			t.Fatal(err)
		}
		if ds.Meta != nil {
			t.Fatal("expected nil Meta before migration")
		}
		MigrateDataset(&ds, raw)
		if ds.Meta == nil {
			t.Fatal("expected Meta to be populated after migration")
		}
		if ds.Meta.CPU == nil || ds.Meta.CPU.Name != "Intel i7" || ds.Meta.CPU.Cores != 8 {
			t.Errorf("CPU = %+v", ds.Meta.CPU)
		}
		if ds.Meta.OS != "linux" {
			t.Errorf("OS = %q", ds.Meta.OS)
		}
		if ds.Meta.Arch != "amd64" {
			t.Errorf("Arch = %q", ds.Meta.Arch)
		}
		if ds.Meta.Pkg != "github.com/foo/bar" {
			t.Errorf("Pkg = %q", ds.Meta.Pkg)
		}
	})

	t.Run("migrates legacy cpu/os on history entries", func(t *testing.T) {
		raw := []byte(`{"name":"test","history":[{"tag":"v1","timestamp":"2025-01-01T00:00:00Z","cpu":{"name":"AMD Ryzen","cores":4},"os":"darwin"}],` + baseSettings + `,"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(raw, &ds); err != nil {
			t.Fatal(err)
		}
		MigrateDataset(&ds, raw)
		if len(ds.History) != 1 {
			t.Fatalf("expected 1 history entry, got %d", len(ds.History))
		}
		h := ds.History[0]
		if h.Meta == nil {
			t.Fatal("expected history entry Meta to be populated")
		}
		if h.Meta.CPU == nil || h.Meta.CPU.Name != "AMD Ryzen" || h.Meta.CPU.Cores != 4 {
			t.Errorf("history CPU = %+v", h.Meta.CPU)
		}
		if h.Meta.OS != "darwin" {
			t.Errorf("history OS = %q", h.Meta.OS)
		}
	})

	t.Run("skips migration when Meta already present", func(t *testing.T) {
		raw := []byte(`{"name":"test","cpu":{"name":"Old CPU","cores":2},"meta":{"cpu":{"name":"New CPU","cores":16},"os":"windows"},` + baseSettings + `,"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(raw, &ds); err != nil {
			t.Fatal(err)
		}
		MigrateDataset(&ds, raw)
		if ds.Meta == nil || ds.Meta.CPU == nil || ds.Meta.CPU.Name != "New CPU" {
			t.Errorf("should not overwrite existing Meta: %+v", ds.Meta)
		}
	})

	t.Run("skips history entry migration when Meta already present", func(t *testing.T) {
		raw := []byte(`{"name":"test","history":[{"tag":"v1","timestamp":"2025-01-01T00:00:00Z","cpu":{"name":"Old"},"meta":{"cpu":{"name":"Keep"},"os":"linux"}}],` + baseSettings + `,"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(raw, &ds); err != nil {
			t.Fatal(err)
		}
		MigrateDataset(&ds, raw)
		if ds.History[0].Meta == nil || ds.History[0].Meta.CPU == nil || ds.History[0].Meta.CPU.Name != "Keep" {
			t.Errorf("should not overwrite existing history Meta: %+v", ds.History[0].Meta)
		}
	})

	t.Run("no migration when no legacy fields and no meta", func(t *testing.T) {
		raw := []byte(`{"name":"test",` + baseSettings + `,"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(raw, &ds); err != nil {
			t.Fatal(err)
		}
		MigrateDataset(&ds, raw)
		if ds.Meta != nil {
			t.Errorf("expected nil Meta, got %+v", ds.Meta)
		}
	})

	t.Run("skips migration when rawJSON is nil", func(t *testing.T) {
		ds := &Dataset{}
		MigrateDataset(ds, nil)
		if ds.Meta != nil {
			t.Errorf("expected nil Meta, got %+v", ds.Meta)
		}
	})

	t.Run("mixed history: only migrates entries with nil Meta", func(t *testing.T) {
		raw := []byte(`{"name":"test","history":[` +
			`{"tag":"v1","timestamp":"2025-01-01T00:00:00Z","cpu":{"name":"Old","cores":2}},` +
			`{"tag":"v2","timestamp":"2025-06-01T00:00:00Z","meta":{"cpu":{"name":"New","cores":8},"os":"linux"}}` +
			`],` + baseSettings + `,"data":[]}`)
		var ds Dataset
		if err := json.Unmarshal(raw, &ds); err != nil {
			t.Fatal(err)
		}
		MigrateDataset(&ds, raw)
		if ds.History[0].Meta == nil || ds.History[0].Meta.CPU == nil || ds.History[0].Meta.CPU.Name != "Old" {
			t.Errorf("entry[0] not migrated: %+v", ds.History[0].Meta)
		}
		if ds.History[1].Meta == nil || ds.History[1].Meta.CPU == nil || ds.History[1].Meta.CPU.Name != "New" {
			t.Errorf("entry[1] overwritten: %+v", ds.History[1].Meta)
		}
	})
}
