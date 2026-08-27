package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/internal/specparse"
	"github.com/goptics/vizb/pkg/cliout"
)

const defaultLogBase = 10.0

// ScaleLinear and ScaleLog are the string-form scales (no axes, default base).
var (
	ScaleLinear = Scale{Type: "linear"}
	ScaleLog    = Scale{Type: "log"}
)

// Scale is a chart value scale. JSON is a string ("linear"|"log") when there
// are no axes and no non-default base; otherwise an object
// {type, axes, base, baseX?, baseY?, baseZ?}.
type Scale struct {
	Type  string
	Axes  []string
	Base  *float64
	BaseX *float64
	BaseY *float64
	BaseZ *float64
}

// IsScaleBag reports whether raw is an unwrapped key=value scale bag
// (`type=log;axes=x`). Nested-colon `log:axes=x` is not a bag.
func IsScaleBag(raw string) bool {
	eq := strings.Index(raw, "=")
	colon := strings.Index(raw, ":")
	return eq >= 0 && (colon < 0 || colon > eq)
}

// EncodeScaleValue converts a --scale flag value (bare linear|log or an
// unwrapped bag) into the Dataset payload: a string or an object map.
func EncodeScaleValue(raw string, fields []flags.ObjectField) any {
	raw = strings.TrimSpace(raw)
	bag, err := ParseObjectBagString(raw, fields)
	if err == nil {
		return ScaleFromBag(bag).Payload()
	}
	return strings.ToLower(raw)
}

// Payload is the JSON-shaped seed value: "linear"/"log" or an object map.
func (s Scale) Payload() any {
	if len(s.Axes) == 0 && !isNonDefaultBase(s.Base) && !isNonDefaultBase(s.BaseX) &&
		!isNonDefaultBase(s.BaseY) && !isNonDefaultBase(s.BaseZ) {
		if s.Type == "" {
			return "linear"
		}
		return s.Type
	}
	m := map[string]any{"type": s.Type}
	if len(s.Axes) > 0 {
		m["axes"] = append([]string(nil), s.Axes...)
	}
	if strings.EqualFold(s.Type, "log") {
		base := defaultLogBase
		if s.Base != nil {
			base = *s.Base
		}
		m["base"] = base
		if isNonDefaultBase(s.BaseX) {
			m["baseX"] = *s.BaseX
		}
		if isNonDefaultBase(s.BaseY) {
			m["baseY"] = *s.BaseY
		}
		if isNonDefaultBase(s.BaseZ) {
			m["baseZ"] = *s.BaseZ
		}
	}
	return m
}

func isNonDefaultBase(p *float64) bool {
	return p != nil && *p != defaultLogBase
}

func (s Scale) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Payload())
}

func (s *Scale) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		*s = Scale{}
		return nil
	}
	if len(trimmed) > 0 && trimmed[0] == '"' {
		var typ string
		if err := json.Unmarshal(trimmed, &typ); err != nil {
			return fmt.Errorf("scale: must be \"linear\" or \"log\"")
		}
		switch strings.ToLower(typ) {
		case "linear", "log":
			*s = Scale{Type: strings.ToLower(typ)}
			return nil
		default:
			return fmt.Errorf("scale value %q is invalid (must be \"linear\" or \"log\")", typ)
		}
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		return fmt.Errorf("scale: must be a string or object")
	}

	out := Scale{}
	for _, key := range sortedRawKeys(fields) {
		if err := decodeScaleField(key, fields[key], &out); err != nil {
			return err
		}
	}
	if out.Type == "" {
		out.Type = "log"
	}
	*s = out
	return nil
}

func decodeScaleField(key string, raw json.RawMessage, out *Scale) error {
	path := "scale." + key
	switch key {
	case "type":
		typ, err := decodeBackgroundString(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		switch strings.ToLower(typ) {
		case "linear", "log":
			out.Type = strings.ToLower(typ)
		default:
			return fmt.Errorf("%s %q is invalid (must be \"linear\" or \"log\")", path, typ)
		}
	case "axes":
		var axes []string
		if err := json.Unmarshal(raw, &axes); err != nil {
			return fmt.Errorf("%s: must be an array of axis names", path)
		}
		out.Axes = axes
	case "base", "baseX", "baseY", "baseZ":
		n, err := decodeBackgroundNumber(raw)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		v := n
		switch key {
		case "base":
			out.Base = &v
		case "baseX":
			out.BaseX = &v
		case "baseY":
			out.BaseY = &v
		default:
			out.BaseZ = &v
		}
	default:
		return fmt.Errorf("scale: unknown field %q", key)
	}
	return nil
}

// ScaleFromBag maps a parsed object bag onto Scale (warn-and-default invalid
// type/base; skip unknown axes).
func ScaleFromBag(bag map[string]any) Scale {
	sc := Scale{}
	if t, ok := bag["type"].(string); ok && t != "" {
		switch strings.ToLower(t) {
		case "linear", "log":
			sc.Type = strings.ToLower(t)
		default:
			cliout.Warn(fmt.Sprintf(
				"Warning: Invalid scale type '%s'. Reason: must be \"linear\" or \"log\". Using default 'linear'",
				t,
			))
			sc.Type = "linear"
		}
	} else {
		sc.Type = "log"
	}
	if raw, ok := bag["axes"].(string); ok {
		sc.Axes = parseScaleAxes(raw)
	}
	sc.Base = parseScaleBase(bag, "base")
	sc.BaseX = parseScaleBase(bag, "baseX")
	sc.BaseY = parseScaleBase(bag, "baseY")
	sc.BaseZ = parseScaleBase(bag, "baseZ")
	return sc
}

func parseScaleAxes(raw string) []string {
	parts := specparse.SplitList(raw)
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		a := strings.ToLower(strings.TrimSpace(p))
		if a != "x" && a != "y" && a != "z" {
			cliout.Warn(fmt.Sprintf(
				"Warning: Invalid scale axis '%s'. Reason: must be x, y, or z. Skipping",
				p,
			))
			continue
		}
		if seen[a] {
			continue
		}
		seen[a] = true
		out = append(out, a)
	}
	return out
}

func parseScaleBase(bag map[string]any, key string) *float64 {
	raw, ok := bag[key]
	if !ok {
		return nil
	}
	n, ok := asScaleFloat(raw)
	if !ok || n <= 0 || n == 1 {
		cliout.Warn(fmt.Sprintf(
			"Warning: Invalid scale %s '%v'. Reason: must be greater than 0 and not 1. Using default '%g'",
			key, raw, defaultLogBase,
		))
		ten := defaultLogBase
		return &ten
	}
	return &n
}

func asScaleFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, !math.IsNaN(n) && !math.IsInf(n, 0)
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}
