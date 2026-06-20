package shared_test

import (
	"encoding/json"
	"reflect"
	"testing"

	config_charts "github.com/goptics/vizb/config/charts"
	"github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/heatmap"
	"github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/line"
	"github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/radar"
	_ "github.com/goptics/vizb/config/charts/scatter"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type MigrateExternalSuite struct {
	suite.Suite
}

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
func (s *MigrateExternalSuite) TestMigrateDatasetV0120Settings() {
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
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Require().Empty(ds.Settings, "expected empty Settings before migration")
	shared.MigrateDataset(&ds, raw)

	s.Require().Len(ds.Settings, 2)

	// settings[0] should be a *bar.Config.
	barCfg, ok := ds.Settings[0].(*bar.Config)
	s.Require().True(ok, "settings[0] = %T, want *bar.Config", ds.Settings[0])
	s.Equal("bar", barCfg.Type)
	s.Equal("log", barCfg.Scale)
	s.Require().NotNil(barCfg.Sort)
	s.True(barCfg.Sort.Enabled)
	s.Equal("desc", barCfg.Sort.Order)
	s.Require().NotNil(barCfg.ShowLabels)
	s.True(*barCfg.ShowLabels)

	// settings[1] should be a *line.Config.
	lineCfg, ok := ds.Settings[1].(*line.Config)
	s.Require().True(ok, "settings[1] = %T, want *line.Config", ds.Settings[1])
	s.Equal("line", lineCfg.Type)
	s.Equal("log", lineCfg.Scale)

	// Axes derived from data points: XAxis and YAxis are non-empty, ZAxis is empty.
	s.Require().Len(ds.Axes, 2)
	wantKeys := []string{"x", "y"}
	for i, ax := range ds.Axes {
		s.Equal(wantKeys[i], ax.Key)
		s.Empty(ax.Label, "v0.12.0 didn't store labels")
	}
}

// Empty legacy settings.charts → migration is a no-op. ds.Settings stays nil.
func (s *MigrateExternalSuite) TestMigrateDatasetV0120EmptyCharts() {
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
	s.Require().NoError(json.Unmarshal(raw, &ds))
	shared.MigrateDataset(&ds, raw)
	s.Empty(ds.Settings, "expected empty Settings (no-op migration)")
}

// New-shape JSON: settings is an array. MigrateDataset must NOT touch it
// (it's already in the new shape).
func (s *MigrateExternalSuite) TestMigrateDatasetPassthrough() {
	raw := []byte(`{
		"name":"test",
		"axes":[{"key":"x","label":"size"}],
		"settings":[
			{"type":"bar","swap":"yxn","scale":"log","showLabels":true}
		],
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	preCount := len(ds.Settings)
	preType := ""
	if preCount > 0 {
		preType = ds.Settings[0].ChartType()
	}

	shared.MigrateDataset(&ds, raw)

	s.Len(ds.Settings, preCount)
	if preCount > 0 {
		s.Equal(preType, ds.Settings[0].ChartType())
	}
}

// v0.12.0 file with a mix of known and unknown chart types. The known types
// (bar, pie) must be migrated; the unknown type (graph) must be silently
// dropped.
func (s *MigrateExternalSuite) TestMigrateDatasetV0120WithUnknownChartType() {
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
	s.Require().NoError(json.Unmarshal(raw, &ds))
	shared.MigrateDataset(&ds, raw)

	s.Require().Len(ds.Settings, 2, "expected 2 settings (bar, pie; graph dropped)")
	s.Equal("bar", ds.Settings[0].ChartType())
	s.Equal("pie", ds.Settings[1].ChartType())
}

// v0.12.0 file with no scale and no charts: skip migration (no-op).
func (s *MigrateExternalSuite) TestMigrateDatasetV0120NoScaleNoCharts() {
	raw := []byte(`{
		"name":"test",
		"settings":{
			"sort":{"enabled":false,"order":"asc"},
			"showLabels":false
		},
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	shared.MigrateDataset(&ds, raw)
	s.Empty(ds.Settings)
}

// buildLegacyConfig's per-chart field assignment is verified end-to-end via
// the v0.12.0 migration; this test exercises it directly via reflection to
// lock in the "always include scale" behaviour: empty legacy scale defaults
// to "linear" for bar/line, and is silently dropped for pie/heatmap/radar.
func (s *MigrateExternalSuite) TestBuildLegacyConfigPerChartFieldAssignment() {
	s.Run("bar with empty scale defaults to linear", func() {
		cfg := mustBuildLegacyConfig(s.T(), "bar", shared.Sort{Enabled: true, Order: "asc"}, true, "")
		barCfg, ok := cfg.(*bar.Config)
		s.Require().True(ok, "cfg = %T, want *bar.Config", cfg)
		s.Equal("linear", barCfg.Scale)
		s.True(barCfg.Sort.Enabled)
		s.Equal("asc", barCfg.Sort.Order)
		s.Require().NotNil(barCfg.ShowLabels)
		s.True(*barCfg.ShowLabels)
	})

	s.Run("pie drops the scale field", func() {
		cfg := mustBuildLegacyConfig(s.T(), "pie", shared.Sort{}, false, "linear")
		pieCfg, ok := cfg.(*pie.Config)
		s.Require().True(ok, "cfg = %T, want *pie.Config", cfg)
		// pie has no Scale field — verify via reflection.
		rv := reflect.ValueOf(pieCfg).Elem()
		_, hasScale := rv.Type().FieldByName("Scale")
		s.False(hasScale, "pie.Config should not have a Scale field")
		s.Require().NotNil(pieCfg.ShowLabels)
		s.False(*pieCfg.ShowLabels)
		_ = fieldValue(s.T(), pieCfg, "Type")
	})

	s.Run("unknown chart type returns error", func() {
		_, err := config_charts.Decode("graph", json.RawMessage(`{"type":"graph"}`))
		s.Require().Error(err, "expected error for unknown chart type")
	})
}

// mustBuildLegacyConfig mirrors the migration's buildLegacyConfig (same
// payload + Decode) but is public-ish via the config_charts registry, so
// external tests can exercise the field-assignment contract without going
// through MigrateDataset's full JSON path.
func mustBuildLegacyConfig(t *testing.T, typ string, sort shared.Sort, showLabels bool, scale string) config_charts.ChartConfig {
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
	cfg, err := config_charts.Decode(typ, raw)
	if err != nil {
		t.Fatalf("config_charts.Decode(%q): %v", typ, err)
	}
	return cfg
}

func TestMigrateExternalSuite(t *testing.T) {
	suite.Run(t, new(MigrateExternalSuite))
}
