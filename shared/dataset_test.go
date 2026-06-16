package shared_test

import (
	"encoding/json"
	"reflect"
	"testing"

	_ "github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/pie"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fieldByName returns the value of a named field via reflection. Used to assert
// on the typed config's fields without importing the per-chart packages
// directly (which would create an import cycle since bar/line/pie/heatmap/radar
// each import shared).
func fieldByName(t *testing.T, v any, name string) any {
	t.Helper()
	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	f := val.FieldByName(name)
	require.True(t, f.IsValid(), "field %q not present on %T", name, v)
	return f.Interface()
}

func TestDataset_UnmarshalJSON_DispatchesByType(t *testing.T) {
	raw := []byte(`{
		"name":"bench",
		"settings":[
			{"type":"bar","swap":"yxn","scale":"log","showLabels":true},
			{"type":"pie","swap":"n","showLabels":false}
		],
		"data":[]
	}`)

	var ds shared.Dataset
	require.NoError(t, json.Unmarshal(raw, &ds))

	require.Len(t, ds.Settings, 2, "expected two settings entries")

	// settings[0] should be a *bar.Config: ChartType=="bar" + Scale field present.
	assert.Equal(t, "bar", ds.Settings[0].ChartType())
	assert.Equal(t, "yxn", fieldByName(t, ds.Settings[0], "Swap"))
	assert.Equal(t, "log", fieldByName(t, ds.Settings[0], "Scale"))
	showLabels, ok := fieldByName(t, ds.Settings[0], "ShowLabels").(*bool)
	require.True(t, ok, "ShowLabels should be *bool, got %T", fieldByName(t, ds.Settings[0], "ShowLabels"))
	require.NotNil(t, showLabels)
	assert.True(t, *showLabels)

	// settings[1] should be a *pie.Config: ChartType=="pie" + no Scale field.
	assert.Equal(t, "pie", ds.Settings[1].ChartType())
	assert.Equal(t, "n", fieldByName(t, ds.Settings[1], "Swap"))
	pieLabels, ok := fieldByName(t, ds.Settings[1], "ShowLabels").(*bool)
	require.True(t, ok, "ShowLabels should be *bool, got %T", fieldByName(t, ds.Settings[1], "ShowLabels"))
	require.NotNil(t, pieLabels)
	assert.False(t, *pieLabels)

	// pie.Config has no Scale field — verify the type distinction via reflection.
	pieVal := reflect.ValueOf(ds.Settings[1])
	for pieVal.Kind() == reflect.Pointer {
		pieVal = pieVal.Elem()
	}
	_, hasScale := pieVal.Type().FieldByName("Scale")
	assert.False(t, hasScale, "pie.Config should not have a Scale field")
}

func TestDataset_UnmarshalJSON_EmptySettings(t *testing.T) {
	raw := []byte(`{"name":"bench","data":[]}`)

	var ds shared.Dataset
	require.NoError(t, json.Unmarshal(raw, &ds))
	assert.Nil(t, ds.Settings, "missing settings field should leave Settings nil")
}

func TestDataset_UnmarshalJSON_LegacySingleObject(t *testing.T) {
	// v0.12.0 wire format uses a single object (not an array) for settings.
	// UnmarshalJSON must not error and must leave Settings nil so that
	// MigrateDataset can populate it from the legacy struct.
	raw := []byte(`{
		"name":"bench",
		"settings":{"charts":["bar"],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"},
		"data":[]
	}`)

	var ds shared.Dataset
	require.NoError(t, json.Unmarshal(raw, &ds))
	assert.Nil(t, ds.Settings, "legacy single-object settings should leave Settings nil (MigrateDataset handles conversion)")
}
