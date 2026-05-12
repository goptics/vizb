package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goptics/vizb/shared"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActionCommandBasic(t *testing.T) {
	orig := shared.ActionState
	defer func() { shared.ActionState = orig }()

	tmpDir := t.TempDir()
	input := `goos: linux
goarch: amd64
pkg: example.com/foo
BenchmarkFoo-16    1000000    1234 ns/op    567 B/op    10 allocs/op
`
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	outPath := filepath.Join(tmpDir, "benchmarks.json")
	shared.ActionState.SHA = "abc123"
	shared.ActionState.Tag = "v1.0.0"
	shared.ActionState.Output = outPath
	shared.ActionState.GroupPattern = "n/y"

	cmd := &cobra.Command{}
	runAction(cmd, []string{inputPath})

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	var bench shared.Benchmark
	require.NoError(t, json.Unmarshal(data, &bench))
	assert.Equal(t, "example.com/foo", bench.Pkg)
	require.GreaterOrEqual(t, len(bench.Data), 1)
	assert.Equal(t, "v1.0.0", bench.Data[0].XAxis)
}

func TestActionCommandWithMerge(t *testing.T) {
	orig := shared.ActionState
	defer func() { shared.ActionState = orig }()

	tmpDir := t.TempDir()

	existing := shared.Benchmark{
		Name: "example.com/foo",
		Pkg:  "example.com/foo",
		Data: []shared.BenchmarkData{
			{Name: "Foo", XAxis: "v1.0.0", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 100}}},
		},
	}
	mergePath := filepath.Join(tmpDir, "existing.json")
	require.NoError(t, shared.WriteJSONFile(mergePath, existing))

	input := "goos: linux\ngoarch: amd64\nBenchmarkFoo-16    100   10 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	outPath := filepath.Join(tmpDir, "benchmarks.json")
	shared.ActionState.SHA = "newsha"
	shared.ActionState.Tag = "v1.1.0"
	shared.ActionState.Merge = mergePath
	shared.ActionState.Output = outPath
	shared.ActionState.GroupPattern = "n/y"

	cmd := &cobra.Command{}
	runAction(cmd, []string{inputPath})

	bench, err := shared.ReadJSONFile[shared.Benchmark](outPath)
	require.NoError(t, err)
	hasOld := false
	hasNew := false
	for _, d := range bench.Data {
		if d.XAxis == "v1.0.0" {
			hasOld = true
		}
		if d.XAxis == "v1.1.0" {
			hasNew = true
		}
	}
	assert.True(t, hasOld, "should preserve existing data")
	assert.True(t, hasNew, "should add new data")
}

func TestActionCommandWithPrune(t *testing.T) {
	orig := shared.ActionState
	defer func() { shared.ActionState = orig }()

	tmpDir := t.TempDir()

	existing := shared.Benchmark{
		Name: "example.com/foo",
		Pkg:  "example.com/foo",
		Runtimes: map[string]time.Time{
			"v1.0.0": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			"v1.1.0": time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			"v1.2.0": time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		Data: []shared.BenchmarkData{
			{Name: "Foo", XAxis: "v1.0.0", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 100}}},
			{Name: "Foo", XAxis: "v1.1.0", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 200}}},
			{Name: "Foo", XAxis: "v1.2.0", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 300}}},
		},
	}
	mergePath := filepath.Join(tmpDir, "existing.json")
	require.NoError(t, shared.WriteJSONFile(mergePath, existing))

	input := "goos: linux\ngoarch: amd64\nBenchmarkFoo-16    100   50 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	outPath := filepath.Join(tmpDir, "benchmarks.json")
	shared.ActionState.SHA = "newsha"
	shared.ActionState.Tag = "v1.3.0"
	shared.ActionState.Merge = mergePath
	shared.ActionState.Output = outPath
	shared.ActionState.Keep = 2
	shared.ActionState.GroupPattern = "n/y"

	cmd := &cobra.Command{}
	runAction(cmd, []string{inputPath})

	bench, err := shared.ReadJSONFile[shared.Benchmark](outPath)
	require.NoError(t, err)
	assert.Len(t, bench.Data, 2)
	assert.Len(t, bench.Runtimes, 2)

	tags := make(map[string]bool)
	for _, d := range bench.Data {
		tags[d.XAxis] = true
	}
	assert.True(t, tags["v1.3.0"], "should keep newest")
	assert.True(t, tags["v1.2.0"], "should keep second newest")
	assert.False(t, tags["v1.0.0"], "should drop oldest")
	assert.False(t, tags["v1.1.0"], "should drop second oldest")
}

func TestActionCommandNoInputFile(t *testing.T) {
	orig := shared.ActionState
	defer func() { shared.ActionState = orig }()

	oldOsExit := shared.OsExit
	defer func() { shared.OsExit = oldOsExit }()
	exitCalled := false
	shared.OsExit = func(code int) {
		exitCalled = true
		panic(fmt.Sprintf("OsExit(%d)", code))
	}

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "out.json")
	shared.ActionState.SHA = "abc"
	shared.ActionState.Tag = "v1.0.0"
	shared.ActionState.Output = outPath
	shared.ActionState.GroupPattern = "n/y"

	cmd := &cobra.Command{}
	assert.Panics(t, func() {
		runAction(cmd, []string{"/tmp/nonexistent_xyz_bench.txt"})
	})
	assert.True(t, exitCalled)
}

func TestActionCommandCustomPattern(t *testing.T) {
	orig := shared.ActionState
	defer func() { shared.ActionState = orig }()

	tmpDir := t.TempDir()
	input := "goos: linux\ngoarch: amd64\npkg: example.com/foo\nBenchmarkAdd/Queue-16    1000000    1234 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	outPath := filepath.Join(tmpDir, "benchmarks.json")
	shared.ActionState.Tag = "v1.0.0"
	shared.ActionState.Output = outPath
	shared.ActionState.GroupPattern = "x/y"

	cmd := &cobra.Command{}
	runAction(cmd, []string{inputPath})

	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	var bench shared.Benchmark
	require.NoError(t, json.Unmarshal(data, &bench))
	require.GreaterOrEqual(t, len(bench.Data), 1)
	assert.Equal(t, "Add", bench.Data[0].XAxis)
	assert.Equal(t, "Queue", bench.Data[0].YAxis)
	assert.Equal(t, "v1.0.0", bench.Data[0].Name)
}
