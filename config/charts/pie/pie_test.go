package pie

import (
	"reflect"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func TestMaterialise_PiePrecedence(t *testing.T) {
	tr := true
	fa := false

	// override beats flags.
	override := &Config{Swap: "yxn", ShowLabels: &tr}
	got := Materialise(Flags{Swap: "xyn", ShowLabels: false}, override)
	assert.Equal(t, "yxn", got.Swap)
	assert.NotNil(t, got.ShowLabels)
	assert.True(t, *got.ShowLabels)

	// flags seed when no override.
	got = Materialise(Flags{Swap: "xyn", ShowLabels: true}, nil)
	assert.Equal(t, "xyn", got.Swap)
	assert.NotNil(t, got.ShowLabels)
	assert.True(t, *got.ShowLabels)

	// nothing set: pointer fields stay nil, internal default does NOT fire
	// (pie has no Scale field, no default).
	got = Materialise(Flags{}, nil)
	assert.Equal(t, "", got.Swap)
	assert.Equal(t, "pie", got.Type)
	assert.Nil(t, got.ShowLabels)
	assert.Nil(t, got.Sort)

	// override.ShowLabels=false wins over flags.ShowLabels=true.
	got = Materialise(Flags{ShowLabels: true}, &Config{ShowLabels: &fa})
	assert.NotNil(t, got.ShowLabels)
	assert.False(t, *got.ShowLabels)

	// override.Sort fills even when flags.Sort is empty.
	overrideSort := &shared.Sort{Enabled: true, Order: "desc"}
	got = Materialise(Flags{}, &Config{Sort: overrideSort})
	assert.NotNil(t, got.Sort)
	assert.True(t, got.Sort.Enabled)
	assert.Equal(t, "desc", got.Sort.Order)

	// flags.Sort builds a Sort struct when non-empty.
	got = Materialise(Flags{Sort: "asc"}, nil)
	assert.NotNil(t, got.Sort)
	assert.True(t, got.Sort.Enabled)
	assert.Equal(t, "asc", got.Sort.Order)
}

// pie has no Scale or AutoRotate fields. This test ensures that invariant —
// if a future maintainer accidentally adds them, the test fails by detecting
// the field's presence in the struct.
func TestPieConfig_NoScaleOrAutoRotate(t *testing.T) {
	typ := reflect.TypeOf(Config{})
	_, hasScale := typ.FieldByName("Scale")
	_, hasAutoRotate := typ.FieldByName("AutoRotate")
	assert.False(t, hasScale, "pie Config should not have a Scale field")
	assert.False(t, hasAutoRotate, "pie Config should not have an AutoRotate field")
}
