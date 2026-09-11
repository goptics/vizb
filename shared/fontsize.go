package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FontSize is the dataset-level chart text size (CSS pixels).
// On the wire it is a hybrid: a JSON number when series, legend, and label are
// all set and equal; otherwise an object of the keys that were set.
type FontSize struct {
	Series *float64 `json:"series,omitempty"`
	Legend *float64 `json:"legend,omitempty"`
	Label  *float64 `json:"label,omitempty"`
}

// Empty reports whether no size is set.
func (f *FontSize) Empty() bool {
	return f == nil || (f.Series == nil && f.Legend == nil && f.Label == nil)
}

func validFontSize(n float64) bool {
	return !math.IsNaN(n) && !math.IsInf(n, 0) && n > 0
}

func (f FontSize) allEqual() bool {
	return f.Series != nil && f.Legend != nil && f.Label != nil &&
		*f.Series == *f.Legend && *f.Legend == *f.Label
}

func (f FontSize) MarshalJSON() ([]byte, error) {
	if f.allEqual() {
		return json.Marshal(*f.Series)
	}
	type wire FontSize
	return json.Marshal(wire(f))
}

// UnmarshalJSON accepts a JSON number or object. Invalid keys and values are
// dropped so a Dataset file still loads; the convert API validates strictly
// before assignment.
func (f *FontSize) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) || len(trimmed) == 0 {
		*f = FontSize{}
		return nil
	}
	if trimmed[0] != '{' {
		n, ok := decodeFontSizeNumber(trimmed)
		if !ok {
			*f = FontSize{}
			return nil
		}
		f.Series, f.Legend, f.Label = F64(n), F64(n), F64(n)
		return nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		*f = FontSize{}
		return nil
	}
	out := FontSize{}
	for key, raw := range fields {
		n, ok := decodeFontSizeNumber(raw)
		if !ok {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "series":
			out.Series = F64(n)
		case "legend":
			out.Legend = F64(n)
		case "label":
			out.Label = F64(n)
		}
	}
	*f = out
	return nil
}

// ParseFontSizeJSONStrict decodes a JSON number or object and errors on unknown
// keys, non-numeric values, non-finite numbers, and values ≤ 0.
func ParseFontSizeJSONStrict(data []byte) (*FontSize, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	if trimmed[0] != '{' {
		var n float64
		if err := json.Unmarshal(trimmed, &n); err != nil {
			return nil, fmt.Errorf("must be a number or object")
		}
		if !validFontSize(n) {
			return nil, fmt.Errorf("must be a finite number greater than 0")
		}
		return &FontSize{Series: F64(n), Legend: F64(n), Label: F64(n)}, nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return nil, fmt.Errorf("must be a number or object")
	}
	out := FontSize{}
	for key, raw := range fields {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "series", "legend", "label":
			var n float64
			if err := json.Unmarshal(raw, &n); err != nil {
				return nil, fmt.Errorf("%s must be a finite number greater than 0", key)
			}
			if !validFontSize(n) {
				return nil, fmt.Errorf("%s must be a finite number greater than 0", key)
			}
			switch strings.ToLower(key) {
			case "series":
				out.Series = F64(n)
			case "legend":
				out.Legend = F64(n)
			case "label":
				out.Label = F64(n)
			}
		default:
			return nil, fmt.Errorf("unknown key %q", key)
		}
	}
	if out.Empty() {
		return nil, nil
	}
	return &out, nil
}

func decodeFontSizeNumber(raw json.RawMessage) (float64, bool) {
	var n float64
	if err := json.Unmarshal(raw, &n); err != nil {
		return 0, false
	}
	if !validFontSize(n) {
		return 0, false
	}
	return n, true
}

// ParseFontSizeFlag parses --font-size: a bare finite float64 > 0 (all three
// keys) or a series=;legend=;label= bag. Invalid keys/values are skipped and
// returned as warning strings. Nil means omit fontSize.
func ParseFontSizeFlag(raw string) (*FontSize, []string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	if !strings.Contains(raw, "=") {
		n, err := parseFontSizeToken(raw)
		if err != nil {
			return nil, []string{fontSizeWarn("font-size", raw, err.Error())}
		}
		return &FontSize{Series: F64(n), Legend: F64(n), Label: F64(n)}, nil
	}

	out := FontSize{}
	var warnings []string
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			warnings = append(warnings, fontSizeWarn("font-size", part, "expected key=value"))
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.TrimSpace(val)
		switch key {
		case "series", "legend", "label":
			n, err := parseFontSizeToken(val)
			if err != nil {
				warnings = append(warnings, fontSizeWarn("font-size "+key, val, err.Error()))
				continue
			}
			switch key {
			case "series":
				out.Series = F64(n)
			case "legend":
				out.Legend = F64(n)
			case "label":
				out.Label = F64(n)
			}
		default:
			warnings = append(warnings, fontSizeWarn("font-size key", key, "unknown key"))
		}
	}
	if out.Empty() {
		if len(warnings) == 0 {
			warnings = append(warnings, fontSizeWarn("font-size", raw, "empty bag"))
		}
		return nil, warnings
	}
	return &out, warnings
}

func parseFontSizeToken(raw string) (float64, error) {
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("must be a finite number greater than 0")
	}
	if !validFontSize(n) {
		return 0, fmt.Errorf("must be a finite number greater than 0")
	}
	return n, nil
}

func fontSizeWarn(label, value, reason string) string {
	return fmt.Sprintf("Warning: Invalid %s '%s'. Reason: %s. Skipping", label, value, reason)
}
