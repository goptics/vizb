package shared

import (
	"testing"
)

func TestWithSafe(t *testing.T) {
	t.Run("No Panic", func(t *testing.T) {
		executed := false
		err := WithSafe("test function", func() {
			executed = true
		})

		if !executed {
			t.Error("Expected function to be executed")
		}

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Panic Recovery", func(t *testing.T) {
		err := WithSafe("test panic", func() {
			panic("test panic message")
		})

		if err == nil {
			t.Error("Expected error from panic recovery")
		}

		expectedMsg := "panic recovered inside test panic: test panic message"
		if err.Error() != expectedMsg {
			t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
		}
	})

	t.Run("Panic with Different Types", func(t *testing.T) {
		// Test with string panic
		err := WithSafe("string panic", func() {
			panic("string error")
		})
		if err == nil {
			t.Error("Expected error from string panic")
		}

		// Test with int panic
		err = WithSafe("int panic", func() {
			panic(42)
		})
		if err == nil {
			t.Error("Expected error from int panic")
		}
	})
}
