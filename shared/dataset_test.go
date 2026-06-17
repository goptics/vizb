package shared_test

import (
	"encoding/json"
	"reflect"
	"testing"

	_ "github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/pie"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/suite"
)

type DatasetSuite struct {
	suite.Suite
}

// fieldByName returns the value of a named field via reflection. Used to assert
// on the typed config's fields without importing the per-chart packages
// directly (which would create an import cycle since bar/line/pie/heatmap/radar
// each import shared).
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

	// settings[0] should be a *bar.Config: ChartType=="bar" + Scale field present.
	s.Equal("bar", ds.Settings[0].ChartType())
	s.Equal("yxn", s.fieldByName(ds.Settings[0], "Swap"))
	s.Equal("log", s.fieldByName(ds.Settings[0], "Scale"))
	showLabels, ok := s.fieldByName(ds.Settings[0], "ShowLabels").(*bool)
	s.Require().True(ok, "ShowLabels should be *bool, got %T", s.fieldByName(ds.Settings[0], "ShowLabels"))
	s.Require().NotNil(showLabels)
	s.True(*showLabels)

	// settings[1] should be a *pie.Config: ChartType=="pie" + no Scale field.
	s.Equal("pie", ds.Settings[1].ChartType())
	s.Equal("n", s.fieldByName(ds.Settings[1], "Swap"))
	pieLabels, ok := s.fieldByName(ds.Settings[1], "ShowLabels").(*bool)
	s.Require().True(ok, "ShowLabels should be *bool, got %T", s.fieldByName(ds.Settings[1], "ShowLabels"))
	s.Require().NotNil(pieLabels)
	s.False(*pieLabels)

	// pie.Config has no Scale field — verify the type distinction via reflection.
	pieVal := reflect.ValueOf(ds.Settings[1])
	for pieVal.Kind() == reflect.Pointer {
		pieVal = pieVal.Elem()
	}
	_, hasScale := pieVal.Type().FieldByName("Scale")
	s.False(hasScale, "pie.Config should not have a Scale field")
}

func (s *DatasetSuite) TestDatasetUnmarshalJSONEmptySettings() {
	raw := []byte(`{"name":"bench","data":[]}`)

	var ds shared.Dataset
	s.Require().NoError(json.Unmarshal(raw, &ds))
	s.Nil(ds.Settings, "missing settings field should leave Settings nil")
}

func (s *DatasetSuite) TestDatasetUnmarshalJSONLegacySingleObject() {
	// v0.12.0 wire format uses a single object (not an array) for settings.
	// UnmarshalJSON must not error and must leave Settings nil so that
	// MigrateDataset can populate it from the legacy struct.
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
