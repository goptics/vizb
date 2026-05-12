# CI Group-Pattern Support — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add configurable `--group-pattern` / `--group-regex` support to the `vizb action` subcommand, reusing the existing parser package instead of duplicating the benchfmt/counting loop.

**Architecture:** `pkg/ci/action.go` stops having its own benchfmt reader, `parseBenchName`, `splitBenchName`, and `resultToStats`. Instead it calls `parser.ParseBenchmarkData()` through `shared.FlagState`, then `ci.InjectTag()` fills the missing dimension with the tag. CI-specific merge/prune/runtimes logic stays in `pkg/ci/`.

**Tech Stack:** Go 1.24, cobra, benchfmt, testify

---

### Task 1: Export `ParsePatternParts` from parser

**Files:**
- Modify: `pkg/parser/parse_pattern.go:63-72`

- [ ] **Step 1: Rename unexported function to exported**

Change line 64 from `func parsePatternParts(pattern string) []string {` to `func ParsePatternParts(pattern string) []string {`.

Update the two callers inside the same file (lines 18 and 32) from `parsePatternParts` to `ParsePatternParts`.

- [ ] **Step 2: Build to verify**

Run: `cd /home/fahim/Projects/goptics/vizb && go build ./pkg/parser/...`
Expected: success

- [ ] **Step 3: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add -p pkg/parser/parse_pattern.go && git commit -m "refactor: export ParsePatternParts from parser"
```

---

### Task 2: Write `TagDimension` and `InjectTag` functions

**Files:**
- Modify: `pkg/ci/action.go` (append new functions)
- Test: `pkg/ci/action_test.go` (append new test functions)

- [ ] **Step 1: Write the failing tests**

Add to `pkg/ci/action_test.go`:

```go
func TestTagDimension(t *testing.T) {
	tests := []struct {
		pattern string
		regex   string
		want    string
		wantErr bool
		errMsg  string
	}{
		{pattern: "n/y", want: "xAxis"},
		{pattern: "n/x", want: "yAxis"},
		{pattern: "x/y", want: "name"},
		{pattern: "x/n", want: "yAxis"},
		{pattern: "y/n", want: "xAxis"},
		{pattern: "y/x", want: "name"},
		{pattern: "n/y/x", wantErr: true, errMsg: "exactly 2 dimensions"},
		{pattern: "n", wantErr: true, errMsg: "exactly 2 dimensions"},
		{pattern: "", wantErr: true},
		{regex: "(?P<name>.*)/(?P<yAxis>.*)", want: "xAxis"},
		{regex: "(?P<n>.*)/(?P<y>.*)", want: "xAxis"},
		{regex: "(?P<x>.*)/(?P<y>.*)", want: "name"},
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
			name:    "3D pattern rejected",
			data:    []shared.BenchmarkData{},
			tag:     "v1.0.0",
			pattern: "n/y/x",
			wantErr: true,
			errMsg:  "exactly 2 dimensions",
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
			got, err := InjectTag(tt.data, tt.tag, tt.pattern, tt.regex)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/fahim/Projects/goptics/vizb && go test ./pkg/ci/ -run "TestTagDimension|TestInjectTag" -v`
Expected: FAIL — `TagDimension` not defined

- [ ] **Step 3: Write implementation**

Add to `pkg/ci/action.go` (before the `pruneBenchData` function):

```go
import (
	"regexp"
	"github.com/goptics/vizb/pkg/parser"
)

// TagDimension returns which dimension (name, xAxis, yAxis) is NOT covered
// by the 2D CI pattern or regex — this is where the tag/commit goes.
func TagDimension(pattern, regexStr string) (string, error) {
	all := []string{"name", "xAxis", "yAxis"}
	var present []string

	if regexStr != "" {
		re, err := regexp.Compile(regexStr)
		if err != nil {
			return "", fmt.Errorf("invalid regex: %w", err)
		}
		for _, name := range re.SubexpNames() {
			if name == "" {
				continue
			}
			present = append(present, expandCIShorthand(name))
		}
	} else {
		if err := validateCIPattern(pattern); err != nil {
			return "", err
		}
		present = parser.ParsePatternParts(pattern)
	}

	presentSet := make(map[string]bool, len(present))
	for _, p := range present {
		presentSet[p] = true
	}

	var missing []string
	for _, d := range all {
		if !presentSet[d] {
			missing = append(missing, d)
		}
	}

	if len(missing) != 1 {
		return "", fmt.Errorf("CI pattern must define exactly 2 dimensions, got %d defined (%v), missing %d (%v)",
			len(presentSet), present, len(missing), missing)
	}

	return missing[0], nil
}

// InjectTag fills the missing dimension in each BenchmarkData with the tag value.
func InjectTag(data []shared.BenchmarkData, tag, pattern, regexStr string) ([]shared.BenchmarkData, error) {
	if tag == "" {
		return data, nil
	}

	tagDim, err := TagDimension(pattern, regexStr)
	if err != nil {
		return nil, err
	}

	result := make([]shared.BenchmarkData, len(data))
	for i, d := range data {
		tagged := d
		setDim(&tagged, tagDim, tag)
		result[i] = tagged
	}
	return result, nil
}

func validateCIPattern(pattern string) error {
	if err := parser.ValidateGroupPattern(pattern); err != nil {
		return err
	}
	parts := parser.ParsePatternParts(pattern)
	if len(parts) != 2 {
		return fmt.Errorf("CI patterns must have exactly 2 dimensions, got %d: use n/y, n/x, x/y, etc.", len(parts))
	}
	return nil
}

func setDim(d *shared.BenchmarkData, dim, value string) {
	switch dim {
	case "name":
		d.Name = value
	case "xAxis":
		d.XAxis = value
	case "yAxis":
		d.YAxis = value
	}
}

func getDim(d shared.BenchmarkData, dim string) string {
	switch dim {
	case "name":
		return d.Name
	case "xAxis":
		return d.XAxis
	case "yAxis":
		return d.YAxis
	}
	return ""
}

func expandCIShorthand(part string) string {
	shortcuts := map[string]string{"n": "name", "x": "xAxis", "y": "yAxis"}
	if expanded, exists := shortcuts[part]; exists {
		return expanded
	}
	return part
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/fahim/Projects/goptics/vizb && go test ./pkg/ci/ -run "TestTagDimension|TestInjectTag" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add pkg/ci/action.go pkg/ci/action_test.go && git commit -m "feat: add TagDimension and InjectTag functions"
```

---

### Task 3: Update `ActionOpts` and `shared.ActionState`

**Files:**
- Modify: `shared/action_state.go:3-11`
- Modify: `pkg/ci/action.go:15-24`

- [ ] **Step 1: Add fields to `shared/action_state.go`**

```go
type actionState struct {
	SHA          string
	Tag          string
	Branch       string
	Merge        string
	Output       string
	HTML         bool
	Keep         int
	GroupPattern string
	GroupRegex   string
}
```

- [ ] **Step 2: Add fields to `ActionOpts`**

```go
type ActionOpts struct {
	Input        string
	Version      string
	Tag          string
	Branch       string
	Date         time.Time
	MergeFile    string
	Output       string
	KeepCount    int
	GroupPattern string
	GroupRegex   string
}
```

- [ ] **Step 3: Build to verify**

Run: `cd /home/fahim/Projects/goptics/vizb && go build ./...`
Expected: success

- [ ] **Step 4: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add shared/action_state.go pkg/ci/action.go && git commit -m "feat: add GroupPattern/GroupRegex to ActionOpts and ActionState"
```

---

### Task 4: Refactor `RunAction` to use `parser.ParseBenchmarkData` + `InjectTag`

**Files:**
- Modify: `pkg/ci/action.go` (rewrite `RunAction`, delete dead code, update `pruneBenchData`)

- [ ] **Step 1: Rewrite RunAction**

Replace `RunAction` (currently lines 26-139) with:

```go
func RunAction(opts ActionOpts) (*shared.Benchmark, error) {
	if _, err := os.Stat(opts.Input); err != nil {
		return nil, fmt.Errorf("input file: %w", err)
	}

	prevPattern := shared.FlagState.GroupPattern
	prevRegex := shared.FlagState.GroupRegex
	shared.FlagState.GroupPattern = opts.GroupPattern
	shared.FlagState.GroupRegex = opts.GroupRegex
	defer func() {
		shared.FlagState.GroupPattern = prevPattern
		shared.FlagState.GroupRegex = prevRegex
	}()

	data := parser.ParseBenchmarkData(opts.Input)

	data, err := InjectTag(data, opts.Tag, opts.GroupPattern, opts.GroupRegex)
	if err != nil {
		return nil, err
	}

	tagDim, err := TagDimension(opts.GroupPattern, opts.GroupRegex)
	if err != nil {
		return nil, err
	}

	bench := shared.Benchmark{
		Name: shared.Pkg,
		Pkg:  shared.Pkg,
		OS:   shared.OS,
		Arch: shared.Arch,
		Data: data,
	}
	bench.CPU.Name = shared.CPU
	bench.Settings.Charts = []string{"bar", "line", "pie"}
	bench.Settings.ShowLabels = true

	if opts.MergeFile != "" {
		existing, err := shared.ReadJSONFile[shared.Benchmark](opts.MergeFile)
		if err == nil {
			var filteredData []shared.BenchmarkData
			for _, d := range existing.Data {
				if getDim(d, tagDim) != opts.Tag {
					filteredData = append(filteredData, d)
				}
			}
			bench.Data = append(filteredData, data...)
			bench.Name = existing.Name
			bench.Description = existing.Description
			bench.CPU = existing.CPU
			bench.OS = existing.OS
			bench.Arch = existing.Arch
			bench.Settings = existing.Settings
			bench.Runtimes = existing.Runtimes
		} else if !errors.Is(err, os.ErrNotExist) {
			// ignore file-not-found for first run
		}
	}

	if bench.Runtimes == nil {
		bench.Runtimes = make(map[string]time.Time)
	}
	bench.Runtimes[opts.Tag] = opts.Date

	if opts.KeepCount > 0 {
		pruneBenchData(&bench, opts.KeepCount, tagDim)
	}

	return &bench, nil
}
```

- [ ] **Step 2: Update `pruneBenchData` signature and logic**

Replace the current `pruneBenchData` (lines 141-172) with:

```go
func pruneBenchData(bench *shared.Benchmark, keep int, tagDim string) {
	tags := make([]string, 0, len(bench.Runtimes))
	for tag := range bench.Runtimes {
		tags = append(tags, tag)
	}
	if len(tags) <= keep {
		return
	}

	sort.Slice(tags, func(i, j int) bool {
		return bench.Runtimes[tags[i]].After(bench.Runtimes[tags[j]])
	})

	keepSet := make(map[string]bool, keep)
	for i := 0; i < keep; i++ {
		keepSet[tags[i]] = true
	}

	var filtered []shared.BenchmarkData
	for _, d := range bench.Data {
		if keepSet[getDim(d, tagDim)] {
			filtered = append(filtered, d)
		}
	}
	bench.Data = filtered

	for tag := range bench.Runtimes {
		if !keepSet[tag] {
			delete(bench.Runtimes, tag)
		}
	}
}
```

- [ ] **Step 3: Delete dead code**

Remove these functions entirely:
- `parseBenchName` (lines 174-187)
- `splitBenchName` (lines 189-196)
- `resultToStats` (lines 198-213)

- [ ] **Step 4: Clean up imports**

```go
import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/goptics/vizb/pkg/parser"
	"github.com/goptics/vizb/shared"
)
```

Remove `"strings"` and `"golang.org/x/perf/benchfmt"`.

- [ ] **Step 5: Build to verify**

Run: `cd /home/fahim/Projects/goptics/vizb && go build ./pkg/ci/...`
Expected: success

- [ ] **Step 6: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add pkg/ci/action.go && git commit -m "refactor: RunAction uses parser.ParseBenchmarkData + InjectTag"
```

---

### Task 5: Update CI action unit tests

**Files:**
- Modify: `pkg/ci/action_test.go` (update existing 3 tests + add new test)

- [ ] **Step 1: Fix all calls to `RunAction`**

In all 3 existing tests + new test, change `_, bench, err := RunAction(opts)` to `bench, err := RunAction(opts)`. Remove the `_` for the removed `*Run` return.

- [ ] **Step 2: Add `GroupPattern: "n/y"` to every `ActionOpts` literal**

In `TestRunActionBasic`, `TestRunActionMergeReplaceByTag`, `TestRunActionPruneOldRuns`.

- [ ] **Step 3: Add custom pattern integration test**

```go
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
		Input:        inputPath,
		Version:      "abc123",
		Tag:          "v1.0.0",
		Date:         time.Now(),
		GroupPattern: "x/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)
	require.NotNil(t, bench)

	require.GreaterOrEqual(t, len(bench.Data), 1)
	assert.Equal(t, "Add", bench.Data[0].XAxis)
	assert.Equal(t, "Queue", bench.Data[0].YAxis)
	assert.Equal(t, "v1.0.0", bench.Data[0].Name)
}
```

- [ ] **Step 4: Add merge-with-proper-dimension test**

Verify that merge correctly filters by the tag dimension even when the tag is NOT in xAxis:

```go
func TestRunActionMergeWithCustomPattern(t *testing.T) {
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
	mergePath := filepath.Join(tmpDir, "existing.json")
	require.NoError(t, shared.WriteJSONFile(mergePath, existing))

	input := "goos: linux\ngoarch: amd64\npkg: example.com/foo\nBenchmarkAdd/Queue-16    100   50 ns/op\n"
	inputPath := filepath.Join(tmpDir, "bench.txt")
	require.NoError(t, os.WriteFile(inputPath, []byte(input), 0644))

	opts := ActionOpts{
		Input:        inputPath,
		Version:      "newsha",
		Tag:          "v1.0.0",
		Date:         time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		MergeFile:    mergePath,
		GroupPattern: "x/y",
	}

	bench, err := RunAction(opts)
	require.NoError(t, err)
	// Should have replaced v1.0.0 data (had name=v1.0.0, now should have fresh stats)
	assert.Len(t, bench.Data, 1)
	assert.Equal(t, "Add", bench.Data[0].XAxis)
	assert.Equal(t, "Queue", bench.Data[0].YAxis)
	assert.Equal(t, "v1.0.0", bench.Data[0].Name)
	assert.InDelta(t, 50.0, bench.Data[0].Stats[0].Value, 0.001)
}
```

- [ ] **Step 5: Run CI tests**

Run: `cd /home/fahim/Projects/goptics/vizb && go test ./pkg/ci/ -v`
Expected: all tests PASS

- [ ] **Step 6: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add pkg/ci/action_test.go && git commit -m "test: update CI action tests for pattern support"
```

---

### Task 6: Wire `--group-pattern` and `--group-regex` flags to CLI

**Files:**
- Modify: `cmd/action.go`

- [ ] **Step 1: Add flags**

Insert after the `--keep` flag in `init()`:

```go
	actionCmd.Flags().StringVar(&shared.ActionState.GroupPattern, "group-pattern", "n/y", "Pattern to parse benchmark names in CI mode (n/y, n/x, x/y, etc.). The missing dimension gets the tag.")
	actionCmd.Flags().StringVar(&shared.ActionState.GroupRegex, "group-regex", "", "Regex to parse benchmark names in CI mode. Overrides group-pattern.")
```

- [ ] **Step 2: Pass to ActionOpts**

Add to the `ci.ActionOpts{}` struct literal in `runAction`:

```go
		GroupPattern: shared.ActionState.GroupPattern,
		GroupRegex:   shared.ActionState.GroupRegex,
```

- [ ] **Step 3: Fix return handling**

Change `_, bench, err := ci.RunAction(opts)` to `bench, err := ci.RunAction(opts)`.

- [ ] **Step 4: Build**

Run: `cd /home/fahim/Projects/goptics/vizb && go build ./cmd/...`
Expected: success

- [ ] **Step 5: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add cmd/action.go && git commit -m "feat: add --group-pattern and --group-regex flags to action subcommand"
```

---

### Task 7: Update CLI action tests

**Files:**
- Modify: `cmd/action_test.go`

- [ ] **Step 1: Set `GroupPattern` in each test function**

Add `shared.ActionState.GroupPattern = "n/y"` to `TestActionCommandBasic`, `TestActionCommandWithMerge`, `TestActionCommandWithPrune`, `TestActionCommandNoInputFile` alongside the other state fields.

- [ ] **Step 2: Add custom pattern test**

```go
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
```

- [ ] **Step 3: Run CLI tests**

Run: `cd /home/fahim/Projects/goptics/vizb && go test ./cmd/ -v`
Expected: all tests PASS

- [ ] **Step 4: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add cmd/action_test.go && git commit -m "test: update CLI tests for pattern support"
```

---

### Task 8: Update `action.yml`

**Files:**
- Modify: `action.yml`

- [ ] **Step 1: Add inputs**

Insert after `generate-html` input:

```yaml
  group-pattern:
    description: 'Pattern to parse benchmark names in CI mode (n/y, n/x, x/y, etc.). The missing dimension gets the tag.'
    required: false
    default: 'n/y'
  group-regex:
    description: 'Regex to parse benchmark names in CI mode. Overrides group-pattern.'
    required: false
    default: ''
```

- [ ] **Step 2: Pass flags to CLI**

Update the "Generate history" step:

```yaml
    - name: Generate history
      run: |
        HTML_FLAG=""
        if [ "${{ inputs.generate-html }}" = "true" ]; then
          HTML_FLAG="--html"
        fi
        GROUP_REGEX_FLAG=""
        if [ "${{ inputs.group-regex }}" != "" ]; then
          GROUP_REGEX_FLAG="--group-regex ${{ inputs.group-regex }}"
        fi
        vizb action bench.txt \
          --sha ${{ github.sha }} \
          --tag ${{ github.ref_name }} \
          --branch ${{ github.ref_name }} \
          --merge benchmarks.json \
          --output benchmarks.json \
          --group-pattern ${{ inputs.group-pattern }} \
          $GROUP_REGEX_FLAG \
          $HTML_FLAG
      shell: bash
```

- [ ] **Step 3: Commit**

```bash
cd /home/fahim/Projects/goptics/vizb && git add action.yml && git commit -m "feat: add group-pattern and group-regex inputs to action.yml"
```

---

### Task 9: Full build and test run

- [ ] **Step 1: Build entire project**

Run: `cd /home/fahim/Projects/goptics/vizb && go build ./...`
Expected: success

- [ ] **Step 2: Run all tests**

Run: `cd /home/fahim/Projects/goptics/vizb && go test ./... -v`
Expected: all packages PASS

- [ ] **Step 3: Run full build (include UI)**

Run: `cd /home/fahim/Projects/goptics/vizb && task build`
Expected: success

- [ ] **Step 4: Commit if any cleanup needed**

```bash
cd /home/fahim/Projects/goptics/vizb && git status
```
