package shared_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	_ "github.com/goptics/vizb/cmd/charts/bar"
	_ "github.com/goptics/vizb/cmd/charts/chord"
	_ "github.com/goptics/vizb/cmd/charts/line"
	_ "github.com/goptics/vizb/cmd/charts/pie"
	_ "github.com/goptics/vizb/cmd/charts/scatter"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type AxisJSONSuite struct {
	suite.Suite
}

type DatasetSuite struct {
	suite.Suite
}

func (s *DatasetSuite) fieldByName(v any, name string) any {
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	f := val.FieldByName(name)
	s.Require().True(f.IsValid(), "field %q not present on %T", name, v)
	return f.Interface()
}

func (s *DatasetSuite) TestDatasetUnmarshalJSONDispatchesByType() {
	raw := []byte(`{
		"name":"bench",
		"settings":[
			{"type":"bar","swap":"yxn","scale":"log","showLabels":true},
			{"type":"pie","swap":"n","showLabels":false}
		],
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Require().Len(ds.Settings, 2, "expected two settings entries")

	s.Equal("bar", ds.Settings[0].ChartType())
	s.Equal("yxn", s.fieldByName(ds.Settings[0], "Swap"))
	s.Equal(shared.ScaleLog, s.fieldByName(ds.Settings[0], "Scale"))
	showLabels, ok := s.fieldByName(ds.Settings[0], "ShowLabels").(*bool)
	s.Require().True(ok, "ShowLabels should be *bool, got %T", s.fieldByName(ds.Settings[0], "ShowLabels"))
	s.Require().NotNil(showLabels)
	s.True(*showLabels)

	s.Equal("pie", ds.Settings[1].ChartType())
	s.Equal("n", s.fieldByName(ds.Settings[1], "Swap"))
	pieLabels, ok := s.fieldByName(ds.Settings[1], "ShowLabels").(*bool)
	s.Require().True(ok, "ShowLabels should be *bool, got %T", s.fieldByName(ds.Settings[1], "ShowLabels"))
	s.Require().NotNil(pieLabels)
	s.False(*pieLabels)

	pieVal := reflect.ValueOf(ds.Settings[1])
	for pieVal.Kind() == reflect.Pointer {
		pieVal = pieVal.Elem()
	}
	_, hasScale := pieVal.Type().FieldByName("Scale")
	s.False(hasScale, "pie.Config should not have a Scale field")
}

func (s *DatasetSuite) TestDatasetUnmarshalJSONLegacySettingsObject() {
	raw := []byte(`{
		"name":"legacy",
		"settings":{"charts":["bar"],"scale":"linear"},
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Nil(ds.Settings, "legacy object settings should stay nil for MigrateDataset")
}

func (s *DatasetSuite) TestDatasetUnmarshalJSONEmptySettings() {
	raw := []byte(`{"name":"bench","data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Nil(ds.Settings, "missing settings field should leave Settings nil")
}

func (s *DatasetSuite) TestDatasetIDTopLevelRoundTrip() {
	raw := []byte(`{
		"id":"bench-v1",
		"theme":"purple-passion",
		"name":"bench",
		"meta":{"os":"linux"},
		"settings":[{"type":"bar"}],
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Equal("bench-v1", ds.ID)
	// Legacy theme string migrates into Themes; Theme is cleared.
	s.Empty(ds.Theme)
	s.Require().Len(ds.Themes, 1)
	s.Equal("purple-passion", ds.Themes[0].Name)
	s.Require().NotNil(ds.Meta)
	s.Equal("linux", ds.Meta.OS)

	out, err := json.Marshal(ds)
	s.Require().NoError(err)
	s.Contains(string(out), `"id":"bench-v1"`)
	s.Contains(string(out), `"themes"`)
	s.NotContains(string(out), `"theme":`)
}

func (s *DatasetSuite) TestUnmarshalNewThemesArray() {
	raw := []byte(`{
		"name":"bench",
		"themes":[
			{"name":"roma","colors":["#E01F54","#001852"],"visualMapColors":["#a4d8c2","#E01F54"]},
			{"name":"custom","colors":["#f00","#0f0","#00f"],"visualMapColors":["#f00","#00f"]}
		],
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.Theme)
	s.Require().Len(ds.Themes, 2)
	s.Equal("roma", ds.Themes[0].Name)
	s.Equal([]string{"#E01F54", "#001852"}, ds.Themes[0].Colors)
	s.Equal([]string{"#a4d8c2", "#E01F54"}, ds.Themes[0].VisualMapColors)
	s.Equal("custom", ds.Themes[1].Name)
	s.Equal([]string{"#f00", "#0f0", "#00f"}, ds.Themes[1].Colors)
}

func (s *DatasetSuite) TestUnmarshalLegacyThemeRoma() {
	raw := []byte(`{"name":"bench","theme":"roma","data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.Theme)
	s.Require().Len(ds.Themes, 1)
	s.Equal("roma", ds.Themes[0].Name)
	s.Equal([]string{
		"#E01F54", "#001852", "#f5e8c8", "#b8d2c7", "#c6b38e",
		"#a4d8c2", "#f3d999", "#d3758f", "#dcc392", "#2e4783",
	}, ds.Themes[0].Colors)
	s.Equal([]string{"#a4d8c2", "#E01F54"}, ds.Themes[0].VisualMapColors)
}

func (s *DatasetSuite) TestUnmarshalLegacyThemeCustomHex() {
	raw := []byte(`{"name":"bench","theme":"#f00,#0f0","data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.Theme)
	s.Require().Len(ds.Themes, 1)
	s.Equal("custom", ds.Themes[0].Name)
	s.Equal([]string{"#f00", "#0f0"}, ds.Themes[0].Colors)
	s.Equal([]string{"#f00", "#0f0"}, ds.Themes[0].VisualMapColors)
}

func (s *DatasetSuite) TestUnmarshalLegacyThemeDefault() {
	for _, value := range []string{"default", "DEFAULT", " Default "} {
		raw := []byte(fmt.Sprintf(`{"name":"bench","theme":%q,"data":[]}`, value))
		var ds shared.Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds), value)
		s.Empty(ds.Themes, value)
		s.Empty(ds.Theme, value)
	}
}

func (s *DatasetSuite) TestUnmarshalLegacyThemeEmpty() {
	raw := []byte(`{"name":"bench","theme":"","data":[]}`)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.Themes)
	s.Empty(ds.Theme)
}

func (s *DatasetSuite) TestMarshalWithThemesOmitsLegacyTheme() {
	ds := shared.Dataset{
		Name: "bench",
		Themes: []shared.Theme{{
			Name:            "roma",
			Colors:          []string{"#E01F54", "#001852"},
			VisualMapColors: []string{"#a4d8c2", "#E01F54"},
		}},
		Data: []shared.DataPoint{},
	}

	out, err := json.Marshal(ds)
	s.Require().NoError(err)
	s.Contains(string(out), `"themes"`)
	s.Contains(string(out), `"roma"`)
	s.NotContains(string(out), `"theme":`)
}

func (s *DatasetSuite) TestThemesRoundTrip() {
	original := shared.Dataset{
		Name: "bench",
		Themes: []shared.Theme{
			{
				Name: "westeros",
				Colors: []string{
					"#516b91", "#59c4e6", "#edafda", "#93b7e3", "#a5e7f0",
					"#cbb0e3", "#3f5575", "#41a7cb", "#d58fc4", "#789bc7",
				},
				VisualMapColors: []string{"#59c4e6", "#d58fc4"},
			},
			{
				Name:            "custom",
				Colors:          []string{"#111", "#222", "#333"},
				VisualMapColors: []string{"#111", "#333"},
			},
		},
		Axes: []shared.Axis{{Key: "x"}},
		Data: []shared.DataPoint{},
	}

	raw, err := json.Marshal(original)
	s.Require().NoError(err)

	var got shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &got))
	s.Empty(got.Theme)
	s.Equal(original.Themes, got.Themes)
}

func (s *DatasetSuite) TestThemesArrayWinsOverLegacyTheme() {
	raw := []byte(`{
		"name":"bench",
		"theme":"roma",
		"themes":[{"name":"chalk","colors":["#fc97af","#87f7cf"],"visualMapColors":["#87f7cf","#fc97af"]}],
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.Theme)
	s.Require().Len(ds.Themes, 1)
	s.Equal("chalk", ds.Themes[0].Name)
}

func (s *DatasetSuite) TestUnmarshalInvalidLegacyThemeKeepsString() {
	raw := []byte(`{"name":"bench","theme":"not-a-theme","data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.Themes)
	s.Equal("not-a-theme", ds.Theme)
}

func (s *DatasetSuite) TestUnmarshalLegacyStructuredDefaultNameClearsTheme() {
	// Structured name "default" is not the bare "default" short-circuit; ParseThemeSpec
	// succeeds with Name "default", then migrate clears Theme without embedding.
	raw := []byte(`{"name":"bench","theme":"default:colors=#f00,#0f0","data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.Themes)
	s.Empty(ds.Theme)
}

func (s *DatasetSuite) TestUnmarshalEmptySettingsArrayStillMigratesTheme() {
	raw := []byte(`{"name":"bench","theme":"roma","settings":[],"data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Nil(ds.Settings)
	s.Empty(ds.Theme)
	s.Require().Len(ds.Themes, 1)
	s.Equal("roma", ds.Themes[0].Name)
}

func (s *DatasetSuite) TestDatasetUnmarshalJSONLegacySingleObject() {
	raw := []byte(`{
		"name":"bench",
		"settings":{"charts":["bar"],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"},
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Nil(ds.Settings, "legacy single-object settings should leave Settings nil (MigrateDataset handles conversion)")
}

func TestDatasetSuite(t *testing.T) {
	suite.Run(t, new(DatasetSuite))
}

func (s *AxisJSONSuite) TestAxisTypeOmittedWhenCategory() {
	t := s.T()
	b, err := json.Marshal(shared.Axis{Key: "x", Label: "Price"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"key":"x","label":"Price"}` {
		t.Fatalf("category axis should omit type, got %s", got)
	}
}

func (s *AxisJSONSuite) TestAxisTypeEmittedWhenValue() {
	t := s.T()
	b, err := json.Marshal(shared.Axis{Key: "x", Label: "Price", Type: "value"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"key":"x","label":"Price","type":"value"}` {
		t.Fatalf("value axis should emit type, got %s", got)
	}
}

func TestAxisJSONSuite(t *testing.T) {
	suite.Run(t, new(AxisJSONSuite))
}
