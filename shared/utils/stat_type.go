package utils

import "fmt"

// CreateStatType generates a formatted stat type based on the stat type, unit, and per value.
func CreateStatType(name, unit, per string) string {
	if unit != "" && per != "" {
		return fmt.Sprintf("%s (%s/%s)", name, unit, per)
	}

	if unit != "" {
		return fmt.Sprintf("%s (%s)", name, unit)
	}

	if per != "" {
		return fmt.Sprintf("%s/%s", name, per)
	}

	return name
}
