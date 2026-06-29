package parser

import (
	"fmt"

	"github.com/goptics/vizb/shared"
)

// AxisColumnKind classifies an axis column for mixed/value routing. axisKey is
// the resolved placement (x, y, or z). Implementations treat x as category when
// not purely numeric; y/z must have at least one numeric cell.
type AxisColumnKind func(source, axisKey string) (kind string, err error)

// ResolveAxesTypes classifies each axis column as category or value.
// Exactly one categorical column is required for mixed mode (on x). All-value
// axes select pure value mode. Two or more categoricals is fatal.
func ResolveAxesTypes(cfg *Config, kindFn AxisColumnKind) error {
	if len(cfg.Axes) == 0 {
		return nil
	}

	catCount := 0
	for i := range cfg.Axes {
		spec := &cfg.Axes[i]
		kind, err := kindFn(spec.Source, spec.AxisKey)
		if err != nil {
			return err
		}
		spec.AxisType = kind
		if kind == "category" {
			catCount++
			if spec.AxisKey != "x" {
				return fmt.Errorf(
					"categorical column %q must be on the x axis; use x:%s (e.g. x:%s,y:latency)",
					spec.Source, spec.Source, spec.Source,
				)
			}
		}
	}

	if catCount == 0 {
		return nil
	}
	if catCount > 1 {
		return fmt.Errorf(
			"axis spec has %d categorical columns; use explicit x:col syntax for one category axis",
			catCount,
		)
	}
	return nil
}

// MixedAxes returns dataset axis descriptors for mixed mode (category x + value y[,z]).
func MixedAxes(cfg Config) []shared.Axis {
	axes := make([]shared.Axis, 0, len(cfg.Axes))
	for _, spec := range cfg.Axes {
		label := spec.Label
		if label == "" {
			label = spec.Source
		}
		axisType := ""
		if spec.AxisType == "value" {
			axisType = "value"
		}
		axes = append(axes, shared.Axis{Key: spec.AxisKey, Label: label, Type: axisType})
	}
	return axes
}

// ValueAxes returns the dataset axis descriptors for pure value mode: each
// selected column becomes a value-type axis on x, y[, z], carrying its
// {label} (falling back to the column name when no label was given).
func ValueAxes(cfg Config) []shared.Axis {
	keys := []string{"x", "y", "z"}
	axes := make([]shared.Axis, 0, len(cfg.Axes))
	for i, spec := range cfg.Axes {
		if i >= len(keys) {
			break
		}
		key := keys[i]
		if spec.AxisKey != "" {
			key = spec.AxisKey
		}
		label := spec.Label
		if label == "" {
			label = spec.Source
		}
		axes = append(axes, shared.Axis{Key: key, Label: label, Type: "value"})
	}
	return axes
}

// SelectViewAxesCfg returns a cfg copy with Axes from the first solo --select view.
func SelectViewAxesCfg(cfg Config) Config {
	c := cfg
	if len(cfg.SelectViews) > 0 {
		c.Axes = append([]ColumnSpec(nil), cfg.SelectViews[0].Columns...)
	}
	return c
}

// DatasetAxesForSelectView builds mixed or value dataset axes from a solo --select
// view. The parser sets AxisType on each column via ResolveAxesTypes inside
// DispatchSelectMode; this helper only projects the carried types into axis
// descriptors. The results parameter is unused (kept for API stability).
func DatasetAxesForSelectView(view []ColumnSpec, _ []shared.DataPoint) []shared.Axis {
	cfg := Config{Axes: append([]ColumnSpec(nil), view...)}
	if isMixedAxes(cfg) {
		return MixedAxes(cfg)
	}
	return ValueAxes(cfg)
}
