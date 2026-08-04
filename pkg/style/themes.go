package style

import (
	"fmt"
	"strings"
)

// Theme is a fully expanded color theme for embedding on datasets.
// Name identifies the theme; Colors is the categorical series palette;
// VisualMapColors is the continuous gradient pair (exactly two entries).
type Theme struct {
	Name            string
	Colors          []string
	VisualMapColors []string // exactly 2
}

// built-in catalog (colors + visual-map pairs match ui/src/lib/themeCatalog.ts).
var themeCatalog = map[string]Theme{
	"default": {
		Name: "default",
		Colors: []string{
			"#5470C6", "#3BA272", "#FC8452", "#73C0DE", "#EE6666",
			"#FAC858", "#9A60B4", "#EA7CCC", "#91CC75", "#FF9F7F",
		},
		VisualMapColors: []string{"#91CC75", "#EE6666"},
	},
	"vintage": {
		Name: "vintage",
		Colors: []string{
			"#d87c7c", "#919e8b", "#d7ab82", "#6e7074", "#61a0a8",
			"#efa18d", "#787464", "#cc7e63", "#724e58", "#4b565b",
		},
		VisualMapColors: []string{"#919e8b", "#d87c7c"},
	},
	"meadow": {
		Name: "meadow",
		Colors: []string{
			"#dd6b66", "#759aa0", "#e69d87", "#8dc1a9", "#ea7e53",
			"#eedd78", "#73a373", "#73b9bc", "#7289ab", "#91ca8c",
		},
		VisualMapColors: []string{"#8dc1a9", "#dd6b66"},
	},
	"westeros": {
		Name: "westeros",
		Colors: []string{
			"#516b91", "#59c4e6", "#edafda", "#93b7e3", "#a5e7f0",
			"#cbb0e3", "#3f5575", "#41a7cb", "#d58fc4", "#789bc7",
		},
		VisualMapColors: []string{"#59c4e6", "#d58fc4"},
	},
	"essos": {
		Name: "essos",
		Colors: []string{
			"#893448", "#d95850", "#eb8146", "#ffb248", "#f2d643",
			"#ebdba4", "#6f293c", "#b84445", "#d96a39", "#d0b72f",
		},
		VisualMapColors: []string{"#f2d643", "#d95850"},
	},
	"wonderland": {
		Name: "wonderland",
		Colors: []string{
			"#4ea397", "#22c3aa", "#7bd9a5", "#d0648a", "#f58db2",
			"#f2b3c9", "#367d75", "#199b89", "#5dbd88", "#ad4f75",
		},
		VisualMapColors: []string{"#4ea397", "#d0648a"},
	},
	"walden": {
		Name: "walden",
		Colors: []string{
			"#3fb1e3", "#6be6c1", "#626c91", "#a0a7e6", "#c4ebad",
			"#96dee8", "#318db5", "#51b99b", "#4e5674", "#8188bc",
		},
		VisualMapColors: []string{"#6be6c1", "#a0a7e6"},
	},
	"chalk": {
		Name: "chalk",
		Colors: []string{
			"#fc97af", "#87f7cf", "#f7f494", "#72ccff", "#f7c5a0",
			"#d4a4eb", "#d2f5a6", "#76f2f2", "#ff7f9f", "#5de0bd",
		},
		VisualMapColors: []string{"#87f7cf", "#fc97af"},
	},
	"infographic": {
		Name: "infographic",
		Colors: []string{
			"#C1232B", "#27727B", "#FCCE10", "#E87C25", "#B5C334",
			"#FE8463", "#9BCA63", "#FAD860", "#F3A43B", "#60C0DD",
		},
		VisualMapColors: []string{"#60C0DD", "#C1232B"},
	},
	"macarons": {
		Name: "macarons",
		Colors: []string{
			"#2ec7c9", "#b6a2de", "#5ab1ef", "#ffb980", "#d87a80",
			"#8d98b3", "#e5cf0d", "#97b552", "#95706d", "#dc69aa",
		},
		VisualMapColors: []string{"#5ab1ef", "#d87a80"},
	},
	"roma": {
		Name: "roma",
		Colors: []string{
			"#E01F54", "#001852", "#f5e8c8", "#b8d2c7", "#c6b38e",
			"#a4d8c2", "#f3d999", "#d3758f", "#dcc392", "#2e4783",
		},
		VisualMapColors: []string{"#a4d8c2", "#E01F54"},
	},
	"shine": {
		Name: "shine",
		Colors: []string{
			"#c12e34", "#e6b600", "#0098d9", "#2b821d", "#005eaa",
			"#339ca8", "#cda819", "#32a487", "#d94a50", "#f2c313",
		},
		VisualMapColors: []string{"#0098d9", "#c12e34"},
	},
	"purple-passion": {
		Name: "purple-passion",
		Colors: []string{
			"#8a7ca8", "#e098c7", "#8fd3e8", "#71669e", "#cc70af",
			"#7cb4cc", "#6d6189", "#ba79aa", "#72b8cc", "#574e7d",
		},
		VisualMapColors: []string{"#8fd3e8", "#cc70af"},
	},
}

// cloneTheme returns a deep copy so callers cannot mutate the catalog.
func cloneTheme(t Theme) Theme {
	out := Theme{
		Name:            t.Name,
		Colors:          append([]string(nil), t.Colors...),
		VisualMapColors: append([]string(nil), t.VisualMapColors...),
	}
	return out
}

// ParseThemeSpec parses a theme specification into a full Theme.
//
// Parse order:
//  1. known built-in name (case-insensitive) → curated colors + visual map
//  2. structured name:colors=#hex,...;visualMapColors=#hex,#hex
//     (colors required; visualMapColors optional → first+last of colors;
//     props semicolon-separated; order independent)
//  3. bare #hex,#hex,... (≥2) → Name "custom", visualMap first+last
//  4. else error
func ParseThemeSpec(value string) (Theme, error) {
	theme, _, err := parseThemeSpec(value)
	return theme, err
}

// parseThemeSpec is ParseThemeSpec plus whether the input was an anonymous
// bare-hex palette (not a built-in name or structured form).
func parseThemeSpec(value string) (Theme, bool, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Theme{}, false, fmt.Errorf("expected a built-in theme, structured theme, or at least two comma-separated hex colors")
	}

	// 1. Built-in name
	if t, ok := themeCatalog[strings.ToLower(trimmed)]; ok {
		return cloneTheme(t), false, nil
	}

	// 2. Structured form: name:prop=...;prop=...
	if name, props, ok := splitStructured(trimmed); ok {
		theme, err := parseStructured(name, props)
		return theme, false, err
	}

	// 3. Bare hex palette (anonymous)
	if colors, err := parseHexList(trimmed); err == nil {
		return Theme{
			Name:            "custom",
			Colors:          colors,
			VisualMapColors: visualMapFromColors(colors),
		}, true, nil
	}

	return Theme{}, false, fmt.Errorf("expected a built-in theme, structured theme, or at least two comma-separated hex colors")
}

// splitStructured splits "name:props" when the value looks structured
// (contains ':' and a colors= property). Returns false for bare names.
func splitStructured(value string) (name, props string, ok bool) {
	idx := strings.Index(value, ":")
	if idx < 0 {
		return "", "", false
	}
	// Bare hex lists never contain ':'.
	name = strings.TrimSpace(value[:idx])
	props = strings.TrimSpace(value[idx+1:])
	if name == "" || props == "" {
		return "", "", false
	}
	// Must look like key=value props (at least one '=').
	if !strings.Contains(props, "=") {
		return "", "", false
	}
	return name, props, true
}

func parseStructured(name, props string) (Theme, error) {
	var colors []string
	var visualMap []string
	hasColors := false
	hasVisualMap := false

	for _, part := range strings.Split(props, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, found := strings.Cut(part, "=")
		if !found {
			return Theme{}, fmt.Errorf("invalid structured theme property %q", part)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch strings.ToLower(key) {
		case "colors":
			parsed, err := parseHexList(val)
			if err != nil {
				return Theme{}, fmt.Errorf("colors: %w", err)
			}
			colors = parsed
			hasColors = true
		case "visualmapcolors":
			parsed, err := parseHexList(val)
			if err != nil {
				return Theme{}, fmt.Errorf("visualMapColors: %w", err)
			}
			if len(parsed) != 2 {
				return Theme{}, fmt.Errorf("visualMapColors requires exactly two hex colors")
			}
			visualMap = parsed
			hasVisualMap = true
		default:
			return Theme{}, fmt.Errorf("unknown structured theme property %q", key)
		}
	}

	if !hasColors {
		return Theme{}, fmt.Errorf("structured theme requires colors with at least two hex values")
	}
	if !hasVisualMap {
		visualMap = visualMapFromColors(colors)
	}

	return Theme{
		Name:            name,
		Colors:          colors,
		VisualMapColors: visualMap,
	}, nil
}

// parseHexList parses a comma-separated list of ≥2 #rgb/#rrggbb colors.
func parseHexList(value string) ([]string, error) {
	parts := strings.Split(value, ",")
	if len(parts) < 2 {
		return nil, fmt.Errorf("expected at least two comma-separated hex colors")
	}
	colors := make([]string, 0, len(parts))
	for _, part := range parts {
		c := strings.TrimSpace(part)
		if !hexColorPattern.MatchString(c) {
			return nil, fmt.Errorf("invalid hex color %q", c)
		}
		colors = append(colors, c)
	}
	return colors, nil
}

func visualMapFromColors(colors []string) []string {
	if len(colors) == 0 {
		return nil
	}
	return []string{colors[0], colors[len(colors)-1]}
}

// ResolveThemes parses theme specs into an ordered catalog for dataset.Themes.
//
// Rules:
//   - each spec is parsed via ParseThemeSpec
//   - built-in "default" is omitted (UI owns default)
//   - anonymous bare-hex customs get unique names: custom, custom-2, …
//   - duplicate names (case-insensitive): last content wins, first-seen order kept
//   - if activeName is non-empty, not "default", and present, that theme is moved first
//   - first returned entry is the active theme when the list is non-empty and active was matched
func ResolveThemes(specs []string, activeName string) ([]Theme, error) {
	type entry struct {
		theme     Theme
		anonymous bool
	}
	entries := make([]entry, 0, len(specs))
	for _, spec := range specs {
		theme, anonymous, err := parseThemeSpec(spec)
		if err != nil {
			return nil, err
		}
		// Skip built-in default (UI owns it). Also skip any theme named "default".
		if strings.EqualFold(theme.Name, "default") {
			continue
		}
		entries = append(entries, entry{
			theme:     theme,
			anonymous: anonymous,
		})
	}

	// Assign unique names to anonymous customs, avoiding collisions with named themes.
	used := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if !e.anonymous {
			used[strings.ToLower(e.theme.Name)] = struct{}{}
		}
	}
	for i := range entries {
		if !entries[i].anonymous {
			continue
		}
		name := nextAnonymousName(used)
		entries[i].theme.Name = name
		used[strings.ToLower(name)] = struct{}{}
	}

	// Dedupe by name (case-insensitive): last content wins, first-seen order.
	byKey := make(map[string]Theme, len(entries))
	order := make([]string, 0, len(entries))
	for _, e := range entries {
		key := strings.ToLower(e.theme.Name)
		if _, exists := byKey[key]; !exists {
			order = append(order, key)
		}
		byKey[key] = e.theme
	}

	out := make([]Theme, 0, len(order))
	for _, key := range order {
		out = append(out, byKey[key])
	}

	// Move active theme first when requested and present.
	active := strings.TrimSpace(activeName)
	if active != "" && !strings.EqualFold(active, "default") {
		out = moveThemeFirst(out, active)
	}

	return out, nil
}

func nextAnonymousName(used map[string]struct{}) string {
	if _, ok := used["custom"]; !ok {
		return "custom"
	}
	for i := 2; ; i++ {
		name := fmt.Sprintf("custom-%d", i)
		if _, ok := used[strings.ToLower(name)]; !ok {
			return name
		}
	}
}

func moveThemeFirst(themes []Theme, activeName string) []Theme {
	idx := -1
	for i, t := range themes {
		if strings.EqualFold(t.Name, activeName) {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return themes
	}
	active := themes[idx]
	out := make([]Theme, 0, len(themes))
	out = append(out, active)
	out = append(out, themes[:idx]...)
	out = append(out, themes[idx+1:]...)
	return out
}
