package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// BorderRadius is CSS/ECharts corner radii: 1–4 non-negative integers
// (TL, TR, BR, BL). JSON is always an array (ECharts accepts [8] for all corners).
type BorderRadius []int

// UnmarshalJSON accepts an array of 1–4 whole numbers ≥ 0.
func (r *BorderRadius) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) {
		*r = nil
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var raw []json.Number
	if err := dec.Decode(&raw); err != nil {
		return fmt.Errorf("borderRadius: must be an array of 1–4 integers")
	}
	if len(raw) < 1 || len(raw) > 4 {
		return fmt.Errorf("borderRadius: must have 1–4 values, got %d", len(raw))
	}

	out := make(BorderRadius, len(raw))
	for i, n := range raw {
		s := n.String()
		if strings.ContainsAny(s, ".eE") {
			return fmt.Errorf("borderRadius: must be an integer, got %s", s)
		}
		v, err := n.Int64()
		if err != nil {
			return fmt.Errorf("borderRadius: must be an integer, got %s", s)
		}
		if v < 0 {
			return fmt.Errorf("borderRadius: must be non-negative, got %d", v)
		}
		out[i] = int(v)
	}
	*r = out
	return nil
}
