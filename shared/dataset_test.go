package shared

import (
	"encoding/json"
	"testing"
)

func TestAxisTypeOmittedWhenCategory(t *testing.T) {
	b, err := json.Marshal(Axis{Key: "x", Label: "Price"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"key":"x","label":"Price"}` {
		t.Fatalf("category axis should omit type, got %s", got)
	}
}

func TestAxisTypeEmittedWhenValue(t *testing.T) {
	b, err := json.Marshal(Axis{Key: "x", Label: "Price", Type: "value"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"key":"x","label":"Price","type":"value"}` {
		t.Fatalf("value axis should emit type, got %s", got)
	}
}
