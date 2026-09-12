package shared

import (
	"bytes"
	"encoding/json"
	"errors"
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

var errInvalidFontSize = errors.New("must be a finite number greater than 0")

func validFontSize(n float64) bool {
	return !math.IsNaN(n) && !math.IsInf(n, 0) && n > 0
}

func (f *FontSize) set(key string, n float64) {
	switch key {
	case "series":
		f.Series = F64(n)
	case "legend":
		f.Legend = F64(n)
	case "label":
		f.Label = F64(n)
	}
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
	fs, _ := parseFontSize(data, false)
	if fs == nil {
		*f = FontSize{}
		return nil
	}
	*f = *fs
	return nil
}

// ParseFontSizeJSONStrict decodes a JSON number or object and errors on unknown
// keys, non-numeric values, non-finite numbers, and values ≤ 0.
func ParseFontSizeJSONStrict(data []byte) (*FontSize, error) {
	return parseFontSize(data, true)
}

// parseFontSize decodes a JSON number (all three keys) or object. In strict mode
// unknown keys and invalid values error; otherwise they are dropped.
func parseFontSize(data []byte, strict bool) (*FontSize, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	if trimmed[0] != '{' {
		var n float64
		if err := json.Unmarshal(trimmed, &n); err != nil {
			if strict {
				return nil, fmt.Errorf("must be a number or object")
			}
			return nil, nil
		}
		if !validFontSize(n) {
			if strict {
				return nil, errInvalidFontSize
			}
			return nil, nil
		}
		return &FontSize{Series: F64(n), Legend: F64(n), Label: F64(n)}, nil
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		if strict {
			return nil, fmt.Errorf("must be a number or object")
		}
		return nil, nil
	}
	out := FontSize{}
	for key, raw := range fields {
		k := strings.ToLower(strings.TrimSpace(key))
		switch k {
		case "series", "legend", "label":
			n, ok := decodeFontSizeNumber(raw)
			if !ok {
				if strict {
					return nil, fmt.Errorf("%s %s", key, errInvalidFontSize)
				}
				continue
			}
			out.set(k, n)
		default:
			if strict {
				return nil, fmt.Errorf("unknown key %q", key)
			}
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
			out.set(key, n)
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
	if err != nil || !validFontSize(n) {
		return 0, errInvalidFontSize
	}
	return n, nil
}

func fontSizeWarn(label, value, reason string) string {
	return fmt.Sprintf("Warning: Invalid %s '%s'. Reason: %s. Skipping", label, value, reason)
}
