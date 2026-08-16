package shared

import (
	"fmt"
	"sort"
	"strings"

	"github.com/goptics/vizb/internal/flags"
	"github.com/goptics/vizb/internal/specparse"
)

// ParseObjectBag converts object-flag bag fields (key=value props from a
// brace-delimited --chart value) into a typed payload map keyed by field name,
// using the flag's ObjectFields descriptors. Unknown keys, bare fields, and
// values that fail their field validation are errors.
func ParseObjectBag(props []specparse.Prop, fields []flags.ObjectField) (map[string]any, error) {
	byName := make(map[string]flags.ObjectField, len(fields))
	for _, field := range fields {
		byName[field.Name] = field
	}

	out := map[string]any{}
	for _, prop := range props {
		field, known := byName[prop.Key]
		if !known {
			return nil, fmt.Errorf("unknown object field %q (valid fields: %s)", prop.Key, objectFieldNames(fields))
		}
		if !prop.HasValue {
			return nil, fmt.Errorf("object field %q requires a value (e.g. %s=…)", prop.Key, prop.Key)
		}
		if field.Validate != nil {
			if err := field.Validate(prop.Value); err != nil {
				return nil, err
			}
		}
		out[prop.Key] = encodeObjectField(field, prop.Value)
	}
	return out, nil
}

// ParseObjectBagString parses a raw semicolon bag ("key=value;key=value") —
// the unwrapped flag form of an object value (e.g. `--bg color=…;opacity=0.5`)
// — into a typed payload map via ParseObjectBag.
func ParseObjectBagString(raw string, fields []flags.ObjectField) (map[string]any, error) {
	props, err := specparse.ParseBag(raw)
	if err != nil {
		return nil, err
	}
	return ParseObjectBag(props, fields)
}

func encodeObjectField(field flags.ObjectField, raw string) any {
	if field.Encode == nil {
		return raw
	}
	return field.Encode(raw)
}

func objectFieldNames(fields []flags.ObjectField) string {
	names := make([]string, 0, len(fields))
	for _, field := range fields {
		names = append(names, field.Name)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
