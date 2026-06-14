package shared

import "encoding/json"

// MigrateDataset populates ds.Settings.Axes from the legacy top-level
// "axisLabels" JSON field when Settings.Axes is empty. rawJSON must be
// the original bytes from which ds was unmarshalled; pass nil to skip
// legacy migration (axes remain empty).
func MigrateDataset(ds *Dataset, rawJSON []byte) {
	// Nothing to do if axes already present or no raw bytes for second pass.
	if len(ds.Settings.Axes) > 0 || len(rawJSON) == 0 {
		return
	}

	// Second-pass: unmarshal only the legacy axisLabels field.
	var legacy struct {
		AxisLabels struct {
			Name string `json:"name,omitempty"`
			X    string `json:"x,omitempty"`
			Y    string `json:"y,omitempty"`
			Z    string `json:"z,omitempty"`
		} `json:"axisLabels"`
	}
	if err := json.Unmarshal(rawJSON, &legacy); err != nil {
		return // not parseable, leave Axes empty
	}

	// Build Axes in canonical order, including only non-empty labels.
	for _, entry := range []struct{ key, label string }{
		{"name", legacy.AxisLabels.Name},
		{"x", legacy.AxisLabels.X},
		{"y", legacy.AxisLabels.Y},
		{"z", legacy.AxisLabels.Z},
	} {
		if entry.label != "" {
			ds.Settings.Axes = append(ds.Settings.Axes, Axis{Key: entry.key, Label: entry.label})
		}
	}
}
