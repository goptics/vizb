package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Background is the bar category background: one wire object folding ECharts
// showBackground (Active) and backgroundStyle (the style fields). Active is
// the on-switch; every style field is optional and maps straight onto the
// rendered series. Numeric style fields are pointers so an explicit zero
// (e.g. borderWidth=0) survives the round trip.
type Background struct {
	Active        bool          `json:"active"`
	Color         string        `json:"color,omitempty"`
	BorderColor   string        `json:"borderColor,omitempty"`
	BorderWidth   *float64      `json:"borderWidth,omitempty"`
	BorderType    string        `json:"borderType,omitempty"`
	BorderRadius  *BorderRadius `json:"borderRadius,omitempty"`
	ShadowBlur    *float64      `json:"shadowBlur,omitempty"`
	ShadowColor   string        `json:"shadowColor,omitempty"`
	ShadowOffsetX *float64      `json:"shadowOffsetX,omitempty"`
	ShadowOffsetY *float64      `json:"shadowOffsetY,omitempty"`
	Opacity       *float64      `json:"opacity,omitempty"`
}

// UnmarshalJSON accepts only the known style keys with their declared types
// and ranges, so a strict decode of a chart config validates the object for
// free: active is a boolean; colors are strings; borderWidth/shadowBlur are
// numbers ≥ 0; borderType is solid|dashed|dotted; borderRadius is a
// non-negative integer or 1–4 of them; shadow offsets are numbers; opacity is
// 0..1. Unknown fields are rejected. Quoted JSON numbers are rejected so
// the REST decode stays as strict as the OpenAPI schema.
func (b *Background) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		return nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return fmt.Errorf("background: must be a JSON object")
	}

	out := Background{}
	for _, key := range sortedRawKeys(fields) {
		if err := decodeBackgroundField(key, fields[key], &out); err != nil {
			return err
		}
	}
	*b = out
	return nil
}

func decodeBackgroundField(key string, raw json.RawMessage, out *Background) error {
	path := "background." + key
	switch key {
	case "active":
		b, err := decodeBackgroundBool(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		out.Active = b
	case "color", "borderColor", "shadowColor":
		s, err := decodeBackgroundString(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		switch key {
		case "color":
			out.Color = s
		case "borderColor":
			out.BorderColor = s
		default:
			out.ShadowColor = s
		}
	case "borderType":
		s, err := decodeBackgroundString(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		switch s {
		case "solid", "dashed", "dotted":
		default:
			return fmt.Errorf("%s %q must be solid, dashed, or dotted", path, s)
		}
		out.BorderType = s
	case "borderWidth", "shadowBlur":
		n, err := decodeBackgroundNumber(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if n < 0 {
			return fmt.Errorf("%s must be non-negative (>= 0), got %s", path, trimmedRaw(raw))
		}
		v := n
		if key == "borderWidth" {
			out.BorderWidth = &v
		} else {
			out.ShadowBlur = &v
		}
	case "shadowOffsetX", "shadowOffsetY":
		n, err := decodeBackgroundNumber(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		v := n
		if key == "shadowOffsetX" {
			out.ShadowOffsetX = &v
		} else {
			out.ShadowOffsetY = &v
		}
	case "opacity":
		n, err := decodeBackgroundNumber(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if n < 0 || n > 1 {
			return fmt.Errorf("%s must be between 0 and 1, got %s", path, trimmedRaw(raw))
		}
		out.Opacity = &n
	case "borderRadius":
		r, err := decodeBackgroundBorderRadius(raw)
		if err != nil {
			return err
		}
		out.BorderRadius = &r
	default:
		return fmt.Errorf("background: unknown field %q", key)
	}
	return nil
}

func decodeBackgroundBool(raw json.RawMessage) (bool, error) {
	var b bool
	if err := json.Unmarshal(raw, &b); err != nil {
		return false, fmt.Errorf("must be a boolean, got %s", trimmedRaw(raw))
	}
	return b, nil
}

func decodeBackgroundString(raw json.RawMessage) (string, error) {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", fmt.Errorf("must be a string, got %s", trimmedRaw(raw))
	}
	return s, nil
}

func decodeBackgroundNumber(raw json.RawMessage) (float64, error) {
	trimmed := trimmedRaw(raw)
	// A JSON string unmarshals into json.Number's string kind, so reject
	// quoted values explicitly before the numeric parse.
	if trimmed == "" || trimmed[0] == '"' {
		return 0, fmt.Errorf("must be a number, got %s", trimmed)
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err != nil {
		return 0, fmt.Errorf("must be a number, got %s", trimmed)
	}
	f, err := n.Float64()
	if err != nil {
		return 0, fmt.Errorf("must be a number, got %s", trimmed)
	}
	return f, nil
}

// decodeBackgroundBorderRadius accepts a non-negative integer or an array of
// 1–4 non-negative integers (normalised to the array wire form).
func decodeBackgroundBorderRadius(raw json.RawMessage) (BorderRadius, error) {
	trimmed := trimmedRaw(raw)
	if strings.HasPrefix(trimmed, "[") {
		// A JSON string unmarshals into json.Number's string kind, so reject
		// quoted elements explicitly before delegating to BorderRadius.
		var elems []json.RawMessage
		if err := json.Unmarshal(raw, &elems); err != nil {
			return nil, fmt.Errorf("background.borderRadius must be an array of 1–4 integers, got %s", trimmed)
		}
		for _, elem := range elems {
			e := trimmedRaw(elem)
			if e == "" || e[0] == '"' {
				return nil, fmt.Errorf("background.borderRadius must be an integer, got %s", e)
			}
		}
		var r BorderRadius
		if err := json.Unmarshal(raw, &r); err != nil {
			return nil, fmt.Errorf("background.%s", err.Error())
		}
		return r, nil
	}
	// A JSON string unmarshals into json.Number's string kind, so reject
	// quoted values explicitly before the integer parse.
	if trimmed == "" || trimmed[0] == '"' {
		return nil, fmt.Errorf("background.borderRadius must be an integer or an array of 1–4 integers, got %s", trimmed)
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil, fmt.Errorf("background.borderRadius must be an integer or an array of 1–4 integers, got %s", trimmed)
	}
	if strings.ContainsAny(n.String(), ".eE") {
		return nil, fmt.Errorf("background.borderRadius must be an integer, got %s", n.String())
	}
	v, err := n.Int64()
	if err != nil || v < 0 {
		return nil, fmt.Errorf("background.borderRadius must be a non-negative integer, got %s", n.String())
	}
	return BorderRadius{int(v)}, nil
}

func trimmedRaw(raw json.RawMessage) string {
	return strings.TrimSpace(string(raw))
}

func sortedRawKeys(fields map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
