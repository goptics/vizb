package line

import (
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func TestMaterialise_LinePrecedence(t *testing.T) {
	tr := true
	fa := false

	// override beats flags.
	override := &Config{Swap: "yxn", Scale: "log", ShowLabels: &tr, AutoRotate: &tr}
	got := Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: false, AutoRotate: false}, override)
	assert.Equal(t, "yxn", got.Swap)
	assert.Equal(t, "log", got.Scale)
	assert.NotNil(t, got.ShowLabels)
	assert.True(t, *got.ShowLabels)
	assert.NotNil(t, got.AutoRotate)
	assert.True(t, *got.AutoRotate)

	// flags seed when no override.
	got = Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: true, AutoRotate: true}, nil)
	assert.Equal(t, "xyn", got.Swap)
	assert.Equal(t, "linear", got.Scale)
	assert.NotNil(t, got.ShowLabels)
	assert.True(t, *got.ShowLabels)
	assert.NotNil(t, got.AutoRotate)
	assert.True(t, *got.AutoRotate)

	// internal default Scale=linear when unset.
	got = Materialise(Flags{}, nil)
	assert.Equal(t, "linear", got.Scale)
	assert.Nil(t, got.ShowLabels)
	assert.Nil(t, got.AutoRotate)

	// override.ShowLabels=false must win over flags.ShowLabels=true.
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

// Asserts line's Config has Scale and AutoRotate fields (compile-time check
// via the precedence test's use of them, and runtime sanity).
func TestLineConfig_HasScaleAndAutoRotate(t *testing.T) {
	got := Materialise(Flags{Scale: "log", AutoRotate: true}, nil)
	assert.Equal(t, "log", got.Scale)
	assert.NotNil(t, got.AutoRotate)
	assert.True(t, *got.AutoRotate)
}
