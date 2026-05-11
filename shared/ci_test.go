package shared

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)
	run := Run{
		Version: "abc123def",
		Tag:     "v1.2.3",
		Date:    now,
		Branch:  "main",
		Goos:    "linux",
		Goarch:  "amd64",
		CPU:     "13th Gen Intel(R) Core(TM) i7-13700",
		Benchmarks: []BenchmarkResult{
			{Name: "BenchmarkTest", Pkg: "example.com/foo", NsPerOp: 1234.5, MBPerSec: 45.6, BytesPerOp: 567, AllocsPerOp: 10},
		},
	}
	data, err := json.Marshal(run)
	require.NoError(t, err)
	var decoded Run
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, run.Version, decoded.Version)
	assert.Equal(t, run.Tag, decoded.Tag)
	assert.True(t, run.Date.Equal(decoded.Date))
	assert.Equal(t, run.Benchmarks[0].Name, decoded.Benchmarks[0].Name)
}

func TestHistoryJSONRoundTrip(t *testing.T) {
	now := time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)
	history := History{
		Meta: HistoryMeta{Title: "My Benchmarks", Description: "Tracking performance"},
		Runs: []Run{
			{Version: "abc123", Date: now, Benchmarks: []BenchmarkResult{{Name: "BenchA", NsPerOp: 100}}},
			{Version: "def456", Date: now.Add(-time.Hour), Benchmarks: []BenchmarkResult{{Name: "BenchA", NsPerOp: 200}}},
		},
	}
	data, err := json.Marshal(history)
	require.NoError(t, err)
	var decoded History
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "My Benchmarks", decoded.Meta.Title)
	assert.Len(t, decoded.Runs, 2)
}

func TestHistoryEmptyRuns(t *testing.T) {
	h := History{Meta: HistoryMeta{Title: "Empty"}}
	data, err := json.Marshal(h)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"runs":null`)
	var decoded History
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Empty(t, decoded.Runs)
}

func TestMergeIntoEmptyHistory(t *testing.T) {
	history := &History{}
	run := Run{Version: "abc", Date: time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)}
	MergeRunIntoHistory(history, run, 0)
	require.Len(t, history.Runs, 1)
	assert.Equal(t, "abc", history.Runs[0].Version)
}

func TestMergePrependsNewest(t *testing.T) {
	older := Run{Version: "old", Date: time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)}
	newer := Run{Version: "new", Date: time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)}
	history := &History{Runs: []Run{older}}
	MergeRunIntoHistory(history, newer, 0)
	require.Len(t, history.Runs, 2)
	assert.Equal(t, "new", history.Runs[0].Version)
	assert.Equal(t, "old", history.Runs[1].Version)
}

func TestMergeSkipsDuplicateSHA(t *testing.T) {
	run := Run{Version: "dup", Date: time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC)}
	history := &History{Runs: []Run{run}}
	MergeRunIntoHistory(history, run, 0)
	require.Len(t, history.Runs, 1)
}

func TestMergePruneToCount(t *testing.T) {
	history := &History{Runs: []Run{}}
	for i := 0; i < 10; i++ {
		run := Run{
			Version: string(rune('a' + i)),
			Date:    time.Date(2026, 5, 11, 10, 0, 0, 0, time.UTC).Add(-time.Duration(i) * time.Hour),
		}
		MergeRunIntoHistory(history, run, 5)
	}
	require.Len(t, history.Runs, 5)
	assert.Equal(t, string(rune('a')), history.Runs[0].Version)
}
