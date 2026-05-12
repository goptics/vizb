package shared

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadJSONFile[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("unmarshal file %s: %w", path, err)
	}
	return &v, nil
}

func WriteJSONFile(path string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
