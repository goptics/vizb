package bar

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/goptics/vizb/config/charts"
	_ "github.com/goptics/vizb/config/charts/heatmap"
	_ "github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/radar"
	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func TestRegistered(t *testing.T) {
	got := charts.Registered()
	sort.Strings(got)
	want := []string{"bar", "heatmap", "line", "pie", "radar"}
	assert.Equal(t, want, got, "Registered() should report all five chart types")
}

func TestRegister_Duplicate(t *testing.T) {
	factory := func() shared.ChartConfig { return &Config{} }
	charts.Register("test_dup", factory)
	assert.Panics(t, func() {
		charts.Register("test_dup", factory)
	}, "Registering the same type twice should panic")
}

func TestNew_UnknownType(t *testing.T) {
	_, err := charts.New("graph")
	assert.Error(t, err, "New should fail for unregistered types")
}

func TestNew_KnownType(t *testing.T) {
	cfg, err := charts.New("bar")
	assert.NoError(t, err)
	bar, ok := cfg.(*Config)
	assert.True(t, ok, "bar factory should return a *Config value")
	assert.Equal(t, "bar", bar.ChartType())
}

func TestDecode_BarRoundTrip(t *testing.T) {
	original := Config{Type: "bar", Swap: "yxn", Scale: "log"}
	raw, err := json.Marshal(original)
	assert.NoError(t, err)

	cfg, err := charts.Decode("bar", raw)
	assert.NoError(t, err)
	got, ok := cfg.(*Config)
	assert.True(t, ok)
	assert.Equal(t, original, *got)
}

func TestDecode_UnknownType(t *testing.T) {
	_, err := charts.Decode("graph", json.RawMessage(`{"type":"graph"}`))
	assert.Error(t, err)
}

func TestMaterialise_BarPrecedence(t *testing.T) {
	// Shared assertion helpers — pointer values are compared via deref.
	tr := true
	fa := false

	// Step 1: override beats flags (Swap overridden, Scale overridden).
	override := &Config{Swap: "yxn", Scale: "log", ShowLabels: &tr, AutoRotate: &tr}
	got := Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: false, AutoRotate: false}, override)
	assert.Equal(t, "yxn", got.Swap, "override.Swap beats flags.Swap")
	assert.Equal(t, "log", got.Scale, "override.Scale beats flags.Scale")
	assert.NotNil(t, got.ShowLabels)
	assert.True(t, *got.ShowLabels, "override.ShowLabels beats flags.ShowLabels")
	assert.NotNil(t, got.AutoRotate)
	assert.True(t, *got.AutoRotate, "override.AutoRotate beats flags.AutoRotate")

	// Step 2: flags seed when no override.
	got = Materialise(Flags{Swap: "xyn", Scale: "linear", ShowLabels: true, AutoRotate: true}, nil)
	assert.Equal(t, "xyn", got.Swap)
	assert.Equal(t, "linear", got.Scale)
	assert.NotNil(t, got.ShowLabels)
	assert.True(t, *got.ShowLabels)
	assert.NotNil(t, got.AutoRotate)
	assert.True(t, *got.AutoRotate)

	// Step 3: when override only fills a subset, the unfilled fields come from flags.
	partial := &Config{Swap: "n"} // only Swap set
	got = Materialise(Flags{Swap: "xyn", Scale: "log", ShowLabels: true, AutoRotate: true}, partial)
	assert.Equal(t, "n", got.Swap, "override.Swap wins")
	assert.Equal(t, "log", got.Scale, "flags.Scale fills un-overridden Scale")
	assert.NotNil(t, got.ShowLabels)
	assert.True(t, *got.ShowLabels)
	assert.NotNil(t, got.AutoRotate)
	assert.True(t, *got.AutoRotate)

	// Step 4: internal default kicks in when nothing is set.
	got = Materialise(Flags{}, nil)
	assert.Equal(t, "", got.Swap)
	assert.Equal(t, "linear", got.Scale, "internal default Scale=linear when unset")
	assert.Nil(t, got.ShowLabels, "ShowLabels nil = JSON absent (user default)")
	assert.Nil(t, got.AutoRotate, "AutoRotate nil = JSON absent (user default)")
	assert.Nil(t, got.Sort, "Sort nil = JSON absent (user default)")

	// *bool false from override (not unset) must round-trip as a non-nil pointer.
	got = Materialise(Flags{ShowLabels: true}, &Config{ShowLabels: &fa})
	assert.NotNil(t, got.ShowLabels)
	assert.False(t, *got.ShowLabels, "override.ShowLabels=false must win over flags.ShowLabels=true")

	// *Sort round-trip: override.Sort fills even when flags.Sort is empty.
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
