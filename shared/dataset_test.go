package shared_test

import (
	"encoding/json"
	"reflect"
	"testing"

	_ "github.com/goptics/vizb/cmd/charts/bar"
	_ "github.com/goptics/vizb/cmd/charts/line"
	_ "github.com/goptics/vizb/cmd/charts/pie"
	_ "github.com/goptics/vizb/cmd/charts/scatter"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

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
	s.Equal("log", s.fieldByName(ds.Settings[0], "Scale"))
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
		"theme":"sunset",
		"name":"bench",
		"meta":{"os":"linux"},
		"settings":[{"type":"bar"}],
		"data":[]
	}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Equal("bench-v1", ds.ID)
	s.Equal("sunset", ds.Theme)
	s.Require().NotNil(ds.Meta)
	s.Equal("linux", ds.Meta.OS)

	out, err := json.Marshal(ds)
	s.Require().NoError(err)
	s.Contains(string(out), `"id":"bench-v1"`)
	s.Contains(string(out), `"theme":"sunset"`)
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

func TestAxisTypeOmittedWhenCategory(t *testing.T) {
	b, err := json.Marshal(shared.Axis{Key: "x", Label: "Price"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"key":"x","label":"Price"}` {
		t.Fatalf("category axis should omit type, got %s", got)
	}
}

func TestAxisTypeEmittedWhenValue(t *testing.T) {
	b, err := json.Marshal(shared.Axis{Key: "x", Label: "Price", Type: "value"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"key":"x","label":"Price","type":"value"}` {
		t.Fatalf("value axis should emit type, got %s", got)
	}
}
