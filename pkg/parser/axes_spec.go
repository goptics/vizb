package parser

import (
	"github.com/goptics/vizb/shared"
)

// ValueAxes returns the dataset axis descriptors for value mode: each
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
