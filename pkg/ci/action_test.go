package ci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunActionBasic(t *testing.T) {
	tmpDir := t.TempDir()
	input := `goos: linux
goarch: amd64
pkg: example.com/foo
BenchmarkAdd/Queue-16    1000000    1234 ns/op    567 B/op    10 allocs/op
BenchmarkAdd/Priority-16 1000000    2345 ns/op    890 B/op    15 allocs/op
`
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:         inputPath,
		IdentifyValue: "v1.0.0",
		Date:          time.Now(),
		GroupPattern:  "n/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)
	require.NotNil(t, bench)

	assert.Equal(t, "example.com/foo", bench.Pkg)
	assert.GreaterOrEqual(t, len(bench.Data), 2)

	for _, d := range bench.Data {
		assert.Equal(t, "v1.0.0", d.XAxis)
	}

	var queueItem *shared.BenchmarkData
	for i := range bench.Data {
		if bench.Data[i].YAxis == "Queue" {
			queueItem = &bench.Data[i]
			break
		}
	}
	require.NotNil(t, queueItem)
	assert.Equal(t, "Add", queueItem.Name)
	assert.Equal(t, "Queue", queueItem.YAxis)

	var hasNsOp, hasBOp, hasAllocs bool
	for _, s := range queueItem.Stats {
		if strings.Contains(s.Type, "Execution Time") {
			hasNsOp = true
		}
		if strings.Contains(s.Type, "Memory Usage") {
			hasBOp = true
		}
		if strings.Contains(s.Type, "Allocations") {
			hasAllocs = true
		}
	}
	assert.True(t, hasNsOp)
	assert.True(t, hasBOp)
	assert.True(t, hasAllocs)
}

func TestRunActionAppendReplaceByTag(t *testing.T) {
	tmpDir := t.TempDir()

	existing := shared.Benchmark{
		Name: "example.com/foo",
		Pkg:  "example.com/foo",
		Runtimes: map[string]time.Time{
			"v1.0.0": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			"v1.1.0": time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Data: []shared.BenchmarkData{
			{Name: "Add", XAxis: "v1.0.0", YAxis: "Queue", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 100}}},
			{Name: "Add", XAxis: "v1.1.0", YAxis: "Queue", Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 200}}},
		},
	}
	appendPath := filepath.Join(tmpDir, "existing.json")
	require.NoError(t, shared.WriteJSONFile(appendPath, existing))

	input := "goos: linux\ngoarch: amd64\npkg: example.com/foo\nBenchmarkAdd/Queue-16    100   50 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:         inputPath,
		IdentifyValue: "v1.0.0",
		Date:          time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		AppendFile:    appendPath,
		GroupPattern:  "n/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)

	assert.Len(t, bench.Data, 2)

	var v10Count, v11Count int
	for _, d := range bench.Data {
		if d.XAxis == "v1.0.0" {
			v10Count++
			if len(d.Stats) > 0 {
				assert.InDelta(t, 50.0, d.Stats[0].Value, 0.001)
			}
		}
		if d.XAxis == "v1.1.0" {
			v11Count++
		}
	}
	assert.Equal(t, 1, v10Count)
	assert.Equal(t, 1, v11Count)

	assert.Contains(t, bench.Runtimes, "v1.0.0")
	assert.Contains(t, bench.Runtimes, "v1.1.0")
}

func TestRunActionPruneOldRuns(t *testing.T) {
	tmpDir := t.TempDir()

	existingTags := []struct {
		tag  string
		date time.Time
	}{
		{"v1.0.0", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"v1.1.0", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
		{"v1.2.0", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
		{"v1.3.0", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)},
		{"v1.4.0", time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
	}

	existing := shared.Benchmark{
		Name:     "example.com/foo",
		Pkg:      "example.com/foo",
		Runtimes: make(map[string]time.Time),
	}
	for _, t := range existingTags {
		existing.Runtimes[t.tag] = t.date
		existing.Data = append(existing.Data, shared.BenchmarkData{
			Name: "Add", XAxis: t.tag, YAxis: "Queue",
			Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 100}},
		})
	}
	appendPath := filepath.Join(tmpDir, "existing.json")
	require.NoError(t, shared.WriteJSONFile(appendPath, existing))

	input := "goos: linux\ngoarch: amd64\npkg: example.com/foo\nBenchmarkAdd/Queue-16    100   50 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:         inputPath,
		IdentifyValue: "v1.5.0",
		Date:          time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		AppendFile:    appendPath,
		KeepCount:     3,
		GroupPattern:  "n/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)

	assert.Len(t, bench.Data, 3)
	assert.Len(t, bench.Runtimes, 3)

	tags := make(map[string]bool)
	for _, d := range bench.Data {
		tags[d.XAxis] = true
	}
	assert.True(t, tags["v1.5.0"], "should keep v1.5.0")
	assert.True(t, tags["v1.4.0"], "should keep v1.4.0")
	assert.True(t, tags["v1.3.0"], "should keep v1.3.0")
	assert.False(t, tags["v1.0.0"], "should drop oldest v1.0.0")
	assert.False(t, tags["v1.1.0"], "should drop v1.1.0")
	assert.False(t, tags["v1.2.0"], "should drop v1.2.0")
}

func TestRunActionFileNotFound(t *testing.T) {
	opts := ActionOpts{
		Input: "/nonexistent/bench.txt",
		Date:  time.Now(),
	}

	_, err := RunAction(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input file")
}

func TestRunActionAppendFileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	input := "goos: linux\ngoarch: amd64\nBenchmarkFoo-16    100   10 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:         inputPath,
		IdentifyValue: "v1.0.0",
		Date:          time.Now(),
		AppendFile:    "/nonexistent/benchmarks.json",
		GroupPattern:  "n/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)
	require.NotNil(t, bench)
	assert.Len(t, bench.Data, 1)
}

func TestRunActionEmptyIdentifyValue(t *testing.T) {
	tmpDir := t.TempDir()
	input := "goos: linux\ngoarch: amd64\nBenchmarkFoo-16    100   10 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:        inputPath,
		Date:         time.Now(),
		GroupPattern: "n/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)
	require.NotNil(t, bench)
	assert.Len(t, bench.Data, 1)
	// When identify value is empty, data is returned as-is
	assert.Equal(t, "", bench.Data[0].XAxis)
}

func TestTagDimension(t *testing.T) {
	tests := []struct {
		pattern string
		regex   string
		want    Dimension
		wantErr bool
		errMsg  string
	}{
		{pattern: "n/y", want: DimXAxis},
		{pattern: "n/x", want: DimYAxis},
		{pattern: "x/y", want: DimName},
		{pattern: "x/n", want: DimYAxis},
		{pattern: "y/n", want: DimXAxis},
		{pattern: "y/x", want: DimName},
		{pattern: "n/y/x", wantErr: true, errMsg: "exactly 2 dimensions"},
		{pattern: "n", wantErr: true, errMsg: "exactly 2 dimensions"},
		{pattern: "", wantErr: true},
		{regex: "(?P<name>.*)/(?P<yAxis>.*)", want: DimXAxis},
		{regex: "(?P<n>.*)/(?P<y>.*)", want: DimXAxis},
		{regex: "(?P<x>.*)/(?P<y>.*)", want: DimName},
		{regex: "bad(regex", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+tt.regex, func(t *testing.T) {
			got, err := TagDimension(tt.pattern, tt.regex)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInjectTag(t *testing.T) {
	tests := []struct {
		name    string
		data    []shared.BenchmarkData
		tag     string
		pattern string
		regex   string
		want    []shared.BenchmarkData
		wantErr bool
		errMsg  string
	}{
		{
			name: "n/y pattern: tag fills xAxis",
			data: []shared.BenchmarkData{
				{Name: "Add", YAxis: "Queue", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
				{Name: "Add", YAxis: "Priority", Stats: []shared.Stat{{Type: "ns/op", Value: 200}}},
			},
			tag:     "v1.0.0",
			pattern: "n/y",
			want: []shared.BenchmarkData{
				{Name: "Add", YAxis: "Queue", XAxis: "v1.0.0", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
				{Name: "Add", YAxis: "Priority", XAxis: "v1.0.0", Stats: []shared.Stat{{Type: "ns/op", Value: 200}}},
			},
		},
		{
			name: "n/x pattern: tag fills yAxis",
			data: []shared.BenchmarkData{
				{Name: "Add", XAxis: "Queue", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
			tag:     "v1.0.0",
			pattern: "n/x",
			want: []shared.BenchmarkData{
				{Name: "Add", XAxis: "Queue", YAxis: "v1.0.0", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
		},
		{
			name: "x/y pattern: tag fills name",
			data: []shared.BenchmarkData{
				{XAxis: "Add", YAxis: "Queue", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
			tag:     "v1.0.0",
			pattern: "x/y",
			want: []shared.BenchmarkData{
				{XAxis: "Add", YAxis: "Queue", Name: "v1.0.0", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
		},
		{
			name: "1D data: n/y pattern, yAxis empty, tag fills xAxis",
			data: []shared.BenchmarkData{
				{Name: "Foo", Stats: []shared.Stat{{Type: "ns/op", Value: 50}}},
			},
			tag:     "v1.0.0",
			pattern: "n/y",
			want: []shared.BenchmarkData{
				{Name: "Foo", XAxis: "v1.0.0", Stats: []shared.Stat{{Type: "ns/op", Value: 50}}},
			},
		},
		{
			name: "3D+ bench name data: n/y, tag fills xAxis",
			data: []shared.BenchmarkData{
				{Name: "Add", YAxis: "Queue/Sub", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
			tag:     "v1.0.0",
			pattern: "n/y",
			want: []shared.BenchmarkData{
				{Name: "Add", YAxis: "Queue/Sub", XAxis: "v1.0.0", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
		},
		{
			name:    "empty tag is no-op",
			data:    []shared.BenchmarkData{{Name: "Foo", YAxis: "Bar"}},
			tag:     "",
			pattern: "n/y",
			want:    []shared.BenchmarkData{{Name: "Foo", YAxis: "Bar"}},
		},
		{
			name: "regex mode: tag fills xAxis",
			data: []shared.BenchmarkData{
				{Name: "Add", YAxis: "Queue", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
			tag:   "v1.0.0",
			regex: "(?P<name>.*)/(?P<yAxis>.*)",
			want: []shared.BenchmarkData{
				{Name: "Add", YAxis: "Queue", XAxis: "v1.0.0", Stats: []shared.Stat{{Type: "ns/op", Value: 100}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagDim, err := TagDimension(tt.pattern, tt.regex)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)

			got := InjectTag(tt.data, tt.tag, tagDim)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunActionWithCustomPattern(t *testing.T) {
	tmpDir := t.TempDir()
	input := `goos: linux
goarch: amd64
pkg: example.com/foo
BenchmarkAdd/Queue-16    1000000    1234 ns/op
`
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:         inputPath,
		IdentifyValue: "v1.0.0",
		Date:          time.Now(),
		GroupPattern:  "x/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)
	require.NotNil(t, bench)

	require.GreaterOrEqual(t, len(bench.Data), 1)
	assert.Equal(t, "Add", bench.Data[0].XAxis)
	assert.Equal(t, "Queue", bench.Data[0].YAxis)
	assert.Equal(t, "v1.0.0", bench.Data[0].Name)
}

func TestRunActionAppendWithCustomPattern(t *testing.T) {
	tmpDir := t.TempDir()

	existing := shared.Benchmark{
		Name: "example.com/foo",
		Pkg:  "example.com/foo",
		Runtimes: map[string]time.Time{
			"v1.0.0": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		Data: []shared.BenchmarkData{
			{Name: "v1.0.0", XAxis: "Add", YAxis: "Queue",
				Stats: []shared.Stat{{Type: "Execution Time (ns/op)", Value: 100}}},
		},
	}
	appendPath := filepath.Join(tmpDir, "existing.json")
	require.NoError(t, shared.WriteJSONFile(appendPath, existing))

	input := "goos: linux\ngoarch: amd64\npkg: example.com/foo\nBenchmarkAdd/Queue-16    100   50 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:         inputPath,
		IdentifyValue: "v1.0.0",
		Date:          time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		AppendFile:    appendPath,
		GroupPattern:  "x/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)
	assert.Len(t, bench.Data, 1)
	assert.Equal(t, "Add", bench.Data[0].XAxis)
	assert.Equal(t, "Queue", bench.Data[0].YAxis)
	assert.Equal(t, "v1.0.0", bench.Data[0].Name)
	assert.InDelta(t, 50.0, bench.Data[0].Stats[0].Value, 0.001)
}
