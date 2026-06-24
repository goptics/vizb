// Package testutil provides shared helpers for vizb integration and unit tests.
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/require"
)

// DefaultBenchLine is the standard Go benchmark fixture used across CLI e2e tests.
const DefaultBenchLine = "BenchmarkExample-8 1000000 1234 ns/op"

// TrapOsExitPanic replaces shared.OsExit with a function that records the call
// and panics with "exit". restore must be deferred (or called from TearDownTest)
// to put shared.OsExit back.
func TrapOsExitPanic(t testing.TB) (restore func(), exitCalled *bool) {
	return shared.TrapOsExitPanic(t)
}

// WriteBenchFile writes content to dir/name and returns the path. Empty content
// uses DefaultBenchLine.
func WriteBenchFile(t testing.TB, dir, name, content string) string {
	t.Helper()
	if content == "" {
		content = DefaultBenchLine
	}
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return path
}

// WriteJSON marshals v and writes it to path.
func WriteJSON(t testing.TB, path string, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
}

// ReadDataset reads and unmarshals a single shared.Dataset from path.
func ReadDataset(t testing.TB, path string) shared.Dataset {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	var ds shared.Dataset
	require.NoError(t, json.Unmarshal(content, &ds))
	return ds
}

// CaptureStdout runs fn with os.Stdout redirected and returns what it printed.
func CaptureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

// CaptureStderr runs fn with os.Stderr redirected and returns what it printed.
func CaptureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	fn()

	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}
