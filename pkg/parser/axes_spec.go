package parser

import (
	"fmt"

	"github.com/goptics/vizb/shared"
)

// ParseAxesFlag parses the --axes value (e.g. "price{Price},latency,mem") into
// ordered column specs for x, y[, z]. It reuses the --select tokenizer for
// {label}/quoting/duplicate handling, then enforces value-mode arity: exactly
// 2 or 3 columns.
func ParseAxesFlag(raw string) ([]ColumnSpec, error) {
	specs, err := ParseSelectFlag(raw)
	if err != nil {
		return nil, err
	}
	if n := len(specs); n < 2 || n > 3 {
		return nil, fmt.Errorf("--axes requires 2 or 3 columns (x,y[,z]); got %d", n)
	}
	return specs, nil
}

// IsHybridMode reports scatter hybrid parsing: 2 categorical group dims plus
// exactly 1 numeric --axes column (z).
func IsHybridMode(cfg Config) bool {
	return len(cfg.Axes) == 1 && len(EffectiveGroupColumns(cfg)) == 2
}

// HybridAxes returns dataset axis descriptors for scatter hybrid mode: x and y
// category axes from the group pattern plus a value-type z axis from --axes.
func HybridAxes(cfg Config) []shared.Axis {
	groupAxes := GroupAxes(cfg)
	axes := make([]shared.Axis, 0, 3)
	for _, a := range groupAxes {
		if a.Key == "x" || a.Key == "y" {
			axes = append(axes, a)
		}
	}
	zLabel := cfg.Axes[0].Label
	if zLabel == "" {
		zLabel = cfg.Axes[0].Source
	}
	axes = append(axes, shared.Axis{Key: "z", Label: zLabel, Type: "value"})
	return axes
}

// ValueAxes returns the dataset axis descriptors for --axes value mode: each
// selected column becomes a value-type axis on x, y[, z] in order, carrying its
// {label} (falling back to the column name when no label was given).
func ValueAxes(cfg Config) []shared.Axis {
	keys := []string{"x", "y", "z"}
	axes := make([]shared.Axis, 0, len(cfg.Axes))
	for i, spec := range cfg.Axes {
		if i >= len(keys) {
			break
		}
		label := spec.Label
		if label == "" {
			label = spec.Source
		}
		axes = append(axes, shared.Axis{Key: keys[i], Label: label, Type: "value"})
	}
	return axes
}
