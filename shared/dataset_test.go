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
		"appearance":{"theme":"purple-passion"},
		"name":"bench",
		"meta":{"os":"linux"},
		"settings":[{"type":"bar"}],
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Equal("bench-v1", ds.ID)
	s.Require().Len(ds.ThemeCatalog(), 1)
	s.Equal("purple-passion", ds.ThemeCatalog()[0].Name)
	s.Require().NotNil(ds.Meta)
	s.Equal("linux", ds.Meta.OS)

	out, err := json.Marshal(ds)
	s.Require().NoError(err)
	s.Contains(string(out), `"id":"bench-v1"`)
	s.Contains(string(out), `"appearance"`)
	s.Contains(string(out), `"themes"`)
	s.NotContains(string(out), `"theme":`)
}

func (s *DatasetSuite) TestUnmarshalNewThemesArray() {
	raw := []byte(`{
		"name":"bench",
		"appearance":{"themes":[
			{"name":"roma","colors":["#E01F54","#001852"],"visualMapColors":["#a4d8c2","#E01F54"]},
			{"name":"custom","colors":["#f00","#0f0","#00f"],"visualMapColors":["#f00","#00f"]}
		]},
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Require().Len(ds.ThemeCatalog(), 2)
	s.Equal("roma", ds.ThemeCatalog()[0].Name)
	s.Equal([]string{"#E01F54", "#001852"}, ds.ThemeCatalog()[0].Colors)
	s.Equal([]string{"#a4d8c2", "#E01F54"}, ds.ThemeCatalog()[0].VisualMapColors)
	s.Equal("custom", ds.ThemeCatalog()[1].Name)
	s.Equal([]string{"#f00", "#0f0", "#00f"}, ds.ThemeCatalog()[1].Colors)
}

func (s *DatasetSuite) TestUnmarshalAppearanceThemeRoma() {
	raw := []byte(`{"name":"bench","appearance":{"theme":"roma"},"data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Require().Len(ds.ThemeCatalog(), 1)
	s.Equal("roma", ds.ThemeCatalog()[0].Name)
	s.Equal([]string{
		"#E01F54", "#001852", "#f5e8c8", "#b8d2c7", "#c6b38e",
		"#a4d8c2", "#f3d999", "#d3758f", "#dcc392", "#2e4783",
	}, ds.ThemeCatalog()[0].Colors)
	s.Equal([]string{"#a4d8c2", "#E01F54"}, ds.ThemeCatalog()[0].VisualMapColors)
}

func (s *DatasetSuite) TestUnmarshalAppearanceThemeCustomHex() {
	raw := []byte(`{"name":"bench","appearance":{"theme":"#f00,#0f0"},"data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Require().Len(ds.ThemeCatalog(), 1)
	s.Equal("custom", ds.ThemeCatalog()[0].Name)
	s.Equal([]string{"#f00", "#0f0"}, ds.ThemeCatalog()[0].Colors)
	s.Equal([]string{"#f00", "#0f0"}, ds.ThemeCatalog()[0].VisualMapColors)
}

func (s *DatasetSuite) TestUnmarshalAppearanceThemeDefault() {
	for _, value := range []string{"default", "DEFAULT", " Default "} {
		raw := []byte(fmt.Sprintf(`{"name":"bench","appearance":{"theme":%q},"data":[]}`, value))
		var ds shared.Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds), value)
		s.Empty(ds.ThemeCatalog(), value)
		s.Nil(ds.Appearance, value)
	}
}

func (s *DatasetSuite) TestUnmarshalAppearanceThemeEmpty() {
	raw := []byte(`{"name":"bench","appearance":{"theme":""},"data":[]}`)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.ThemeCatalog())
	s.Nil(ds.Appearance)
}

func (s *DatasetSuite) TestMarshalWithThemesOmitsLegacyTheme() {
	ds := shared.Dataset{
		Name: "bench",
		Appearance: &shared.Appearance{Themes: []shared.Theme{{
			Name:            "roma",
			Colors:          []string{"#E01F54", "#001852"},
			VisualMapColors: []string{"#a4d8c2", "#E01F54"},
		}}},
		Data: []shared.DataPoint{},
	}

	out, err := json.Marshal(ds)
	s.Require().NoError(err)
	s.Contains(string(out), `"appearance"`)
	s.Contains(string(out), `"themes"`)
	s.Contains(string(out), `"roma"`)
	s.NotContains(string(out), `"theme":`)
}

func (s *DatasetSuite) TestThemesRoundTrip() {
	original := shared.Dataset{
		Name: "bench",
		Appearance: &shared.Appearance{Themes: []shared.Theme{
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
		}},
		Axes: []shared.Axis{{Key: "x"}},
		Data: []shared.DataPoint{},
	}

	raw, err := json.Marshal(original)
	s.Require().NoError(err)

	var got shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &got))
	s.Equal(original.ThemeCatalog(), got.ThemeCatalog())
}

func (s *DatasetSuite) TestAppearanceThemesWinsOverThemeString() {
	raw := []byte(`{
		"name":"bench",
		"appearance":{
			"theme":"roma",
			"themes":[{"name":"chalk","colors":["#fc97af","#87f7cf"],"visualMapColors":["#87f7cf","#fc97af"]}]
		},
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Require().Len(ds.ThemeCatalog(), 1)
	s.Equal("chalk", ds.ThemeCatalog()[0].Name)
	s.Empty(ds.Appearance.Theme)
}

func (s *DatasetSuite) TestUnmarshalInvalidAppearanceThemeKeepsString() {
	raw := []byte(`{"name":"bench","appearance":{"theme":"not-a-theme"},"data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.ThemeCatalog())
	s.Equal("not-a-theme", ds.Appearance.Theme)
}

func (s *DatasetSuite) TestUnmarshalAppearanceStructuredDefaultNameClearsTheme() {
	raw := []byte(`{"name":"bench","appearance":{"theme":"default:colors=#f00,#0f0"},"data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Empty(ds.ThemeCatalog())
	s.Nil(ds.Appearance)
}

func (s *DatasetSuite) TestUnmarshalEmptySettingsArrayStillMigratesTheme() {
	raw := []byte(`{"name":"bench","appearance":{"theme":"roma"},"settings":[],"data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Nil(ds.Settings)
	s.Require().Len(ds.ThemeCatalog(), 1)
	s.Equal("roma", ds.ThemeCatalog()[0].Name)
}

func (s *DatasetSuite) TestAppearanceFontSizeRoundTrip() {
	ds := shared.Dataset{
		Name: "bench",
		Appearance: &shared.Appearance{
			FontSize: &shared.FontSize{
				Series: shared.F64(16),
				Legend: shared.F64(10),
			},
		},
		Data: []shared.DataPoint{},
	}
	raw, err := json.Marshal(ds)
	s.Require().NoError(err)
	s.Contains(string(raw), `"fontSize"`)
	s.Contains(string(raw), `"series"`)
	s.NotContains(string(raw), `"theme":`)

	var got shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &got))
	s.Require().NotNil(got.Appearance.FontSize)
	s.Equal(16.0, *got.Appearance.FontSize.Series)
	s.Equal(10.0, *got.Appearance.FontSize.Legend)
	s.Nil(got.Appearance.FontSize.Label)

	all := shared.Dataset{
		Name: "bench",
		Appearance: &shared.Appearance{
			FontSize: &shared.FontSize{Series: shared.F64(14), Legend: shared.F64(14), Label: shared.F64(14)},
		},
		Data: []shared.DataPoint{},
	}
	raw, err = json.Marshal(all)
	s.Require().NoError(err)
	s.Contains(string(raw), `"fontSize":14`)
}

func (s *DatasetSuite) TestRootThemeFieldsAreIgnored() {
	raw := []byte(`{"name":"bench","theme":"roma","themes":[{"name":"chalk","colors":["#fc97af"]}],"data":[]}`)
	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Nil(ds.Appearance)
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
