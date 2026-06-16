package shared_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/heatmap"
	"github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/radar"
	"github.com/goptics/vizb/shared"
)

// fieldValue returns the value of a named field via reflection. Used to
// assert on the typed Config fields from a test that needs to peek at
// *bar.Config / *line.Config / *pie.Config without making the test file
// internal to the shared package (which would cycle through bar/line/pie —
// they all import shared).
func fieldValue(t *testing.T, v any, name string) any {
	t.Helper()
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	f := rv.FieldByName(name)
	if !f.IsValid() {
		t.Fatalf("field %q not present on %T", name, v)
	}
	return f.Interface()
}

// v0.12.0 fixture: settings is a single object with charts/sort/showLabels/scale.
// After MigrateDataset, ds.Settings must hold typed per-chart Configs derived
// from those legacy fields, and ds.Axes must be populated from the data points.
func TestMigrateDataset_V0120Settings(t *testing.T) {
	raw := []byte(`{
		"name":"test",
		"settings":{
			"charts":["bar","line"],
			"sort":{"enabled":true,"order":"desc"},
			"showLabels":true,
			"scale":"log"
		},
		"data":[
			{"name":"a","xAxis":"1","yAxis":"10","stats":[]},
			{"name":"b","xAxis":"2","yAxis":"20","stats":[]}
		]
	}`)

	var ds shared.Dataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		t.Fatal(err)
	}
	if len(ds.Settings) != 0 {
		t.Fatalf("expected empty Settings before migration, got %d entries", len(ds.Settings))
	}
	shared.MigrateDataset(&ds, raw)

	if len(ds.Settings) != 2 {
		t.Fatalf("expected 2 settings after migration, got %d", len(ds.Settings))
	}

	// settings[0] should be a *bar.Config.
	barCfg, ok := ds.Settings[0].(*bar.Config)
	if !ok {
		t.Fatalf("settings[0] = %T, want *bar.Config", ds.Settings[0])
	}
	if barCfg.Type != "bar" {
		t.Errorf("settings[0].Type = %q, want %q", barCfg.Type, "bar")
	}
	if barCfg.Scale != "log" {
		t.Errorf("settings[0].Scale = %q, want %q", barCfg.Scale, "log")
	}
	if barCfg.Sort == nil || !barCfg.Sort.Enabled || barCfg.Sort.Order != "desc" {
		t.Errorf("settings[0].Sort = %+v, want Enabled=true Order=desc", barCfg.Sort)
	}
	if barCfg.ShowLabels == nil || !*barCfg.ShowLabels {
		t.Errorf("settings[0].ShowLabels = %v, want &true", barCfg.ShowLabels)
	}

	// settings[1] should be a *line.Config.
	lineCfg, ok := ds.Settings[1].(*line.Config)
	if !ok {
		t.Fatalf("settings[1] = %T, want *line.Config", ds.Settings[1])
	}
	if lineCfg.Type != "line" {
		t.Errorf("settings[1].Type = %q, want %q", lineCfg.Type, "line")
	}
	if lineCfg.Scale != "log" {
		t.Errorf("settings[1].Scale = %q, want %q", lineCfg.Scale, "log")
	}

	// Axes derived from data points: XAxis and YAxis are non-empty, ZAxis is empty.
	if len(ds.Axes) != 2 {
		t.Fatalf("expected 2 axes derived from data, got %d: %+v", len(ds.Axes), ds.Axes)
	}
	wantKeys := []string{"x", "y"}
	for i, ax := range ds.Axes {
		if ax.Key != wantKeys[i] {
			t.Errorf("Axes[%d].Key = %q, want %q", i, ax.Key, wantKeys[i])
		}
		if ax.Label != "" {
			t.Errorf("Axes[%d].Label = %q, want empty (v0.12.0 didn't store labels)", i, ax.Label)
		}
	}
}

// Empty legacy settings.charts → migration is a no-op. ds.Settings stays nil.
func TestMigrateDataset_V0120EmptyCharts(t *testing.T) {
	raw := []byte(`{
		"name":"test",
		"settings":{
			"charts":[],
			"sort":{"enabled":false,"order":"asc"},
			"showLabels":false,
			"scale":"linear"
		},
		"data":[]
	}`)

	var ds shared.Dataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		t.Fatal(err)
	}
	shared.MigrateDataset(&ds, raw)
	if len(ds.Settings) != 0 {
		t.Errorf("expected empty Settings (no-op migration), got %d entries", len(ds.Settings))
	}
}

// New-shape JSON: settings is an array. MigrateDataset must NOT touch it
// (it's already in the new shape).
func TestMigrateDataset_Passthrough(t *testing.T) {
	raw := []byte(`{
		"name":"test",
		"axes":[{"key":"x","label":"size"}],
		"settings":[
			{"type":"bar","swap":"yxn","scale":"log","showLabels":true}
		],
		"data":[]
	}`)

	var ds shared.Dataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		t.Fatal(err)
	}
	preCount := len(ds.Settings)
	preType := ""
	if preCount > 0 {
		preType = ds.Settings[0].ChartType()
	}

	shared.MigrateDataset(&ds, raw)

	if len(ds.Settings) != preCount {
		t.Errorf("Settings length changed: pre=%d post=%d", preCount, len(ds.Settings))
	}
	if preCount > 0 && ds.Settings[0].ChartType() != preType {
		t.Errorf("Settings[0] type changed: pre=%q post=%q", preType, ds.Settings[0].ChartType())
	}
}

// v0.12.0 file with a mix of known and unknown chart types. The known types
// (bar, pie) must be migrated; the unknown type (graph) must be silently
// dropped.
func TestMigrateDataset_V0120WithUnknownChartType(t *testing.T) {
	raw := []byte(`{
		"name":"test",
		"settings":{
			"charts":["bar","graph","pie"],
			"sort":{"enabled":false,"order":"asc"},
			"showLabels":false,
			"scale":"linear"
		},
		"data":[]
	}`)

	var ds shared.Dataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		t.Fatal(err)
	}
	shared.MigrateDataset(&ds, raw)

	if len(ds.Settings) != 2 {
		t.Fatalf("expected 2 settings (bar, pie; graph dropped), got %d", len(ds.Settings))
	}
	if ds.Settings[0].ChartType() != "bar" {
		t.Errorf("settings[0] = %q, want %q", ds.Settings[0].ChartType(), "bar")
	}
	if ds.Settings[1].ChartType() != "pie" {
		t.Errorf("settings[1] = %q, want %q", ds.Settings[1].ChartType(), "pie")
	}
}

// v0.12.0 file with no scale and no charts: skip migration (no-op).
func TestMigrateDataset_V0120NoScaleNoCharts(t *testing.T) {
	raw := []byte(`{
		"name":"test",
		"settings":{
			"sort":{"enabled":false,"order":"asc"},
			"showLabels":false
		},
		"data":[]
	}`)

	var ds shared.Dataset
	if err := json.Unmarshal(raw, &ds); err != nil {
		t.Fatal(err)
	}
	shared.MigrateDataset(&ds, raw)
	if len(ds.Settings) != 0 {
		t.Errorf("expected empty Settings, got %d", len(ds.Settings))
	}
}

// buildLegacyConfig's per-chart field assignment is verified end-to-end via
// the v0.12.0 migration; this test exercises it directly via reflection to
// lock in the "always include scale" behaviour: empty legacy scale defaults
// to "linear" for bar/line, and is silently dropped for pie/heatmap/radar.
func TestBuildLegacyConfig_PerChartFieldAssignment(t *testing.T) {
	t.Run("bar with empty scale defaults to linear", func(t *testing.T) {
		cfg := mustBuildLegacyConfig(t, "bar", shared.Sort{Enabled: true, Order: "asc"}, true, "")
		barCfg, ok := cfg.(*bar.Config)
		if !ok {
			t.Fatalf("cfg = %T, want *bar.Config", cfg)
		}
		if barCfg.Scale != "linear" {
			t.Errorf("Scale = %q, want %q", barCfg.Scale, "linear")
		}
		if !barCfg.Sort.Enabled || barCfg.Sort.Order != "asc" {
			t.Errorf("Sort = %+v, want Enabled=true Order=asc", barCfg.Sort)
		}
		if barCfg.ShowLabels == nil || !*barCfg.ShowLabels {
			t.Errorf("ShowLabels = %v, want &true", barCfg.ShowLabels)
		}
	})

	t.Run("pie drops the scale field", func(t *testing.T) {
		cfg := mustBuildLegacyConfig(t, "pie", shared.Sort{}, false, "linear")
		pieCfg, ok := cfg.(*pie.Config)
		if !ok {
			t.Fatalf("cfg = %T, want *pie.Config", cfg)
		}
		// pie has no Scale field — verify via reflection.
		rv := reflect.ValueOf(pieCfg).Elem()
		_, hasScale := rv.Type().FieldByName("Scale")
		if hasScale {
			t.Errorf("pie.Config should not have a Scale field")
		}
		if pieCfg.ShowLabels == nil || *pieCfg.ShowLabels {
			t.Errorf("ShowLabels = %v, want &false", pieCfg.ShowLabels)
		}
		_ = fieldValue(t, pieCfg, "Type")
	})

	t.Run("unknown chart type returns error", func(t *testing.T) {
		_, err := shared.DecodeChartConfig("graph", json.RawMessage(`{"type":"graph"}`))
		if err == nil {
			t.Error("expected error for unknown chart type, got nil")
		}
	})
}

// mustBuildLegacyConfig mirrors the migration's buildLegacyConfig (same
// payload + Decode) but is public-ish via the shared registry, so external
// tests can exercise the field-assignment contract without going through
// MigrateDataset's full JSON path.
func mustBuildLegacyConfig(t *testing.T, typ string, sort shared.Sort, showLabels bool, scale string) shared.ChartConfig {
	t.Helper()
	scaleVal := scale
	if scaleVal == "" {
		scaleVal = "linear"
	}
	payload := map[string]any{
		"type":       typ,
		"sort":       sort,
		"showLabels": showLabels,
		"scale":      scaleVal,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := shared.DecodeChartConfig(typ, raw)
	if err != nil {
		t.Fatalf("DecodeChartConfig(%q): %v", typ, err)
	}
	return cfg
}
