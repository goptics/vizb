package json

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func writeJSONTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func resetFlags(t *testing.T) {
	t.Helper()
	orig := shared.FlagState
	t.Cleanup(func() { shared.FlagState = orig })

	shared.FlagState.Group = nil
	shared.FlagState.GroupPattern = "x"
	shared.FlagState.GroupRegex = ""
	shared.FlagState.FilterRegex = ""
	shared.FlagState.NumberUnit = ""
}

func statTypes(stats []shared.Stat) []string {
	out := make([]string, len(stats))
	for i, s := range stats {
		out[i] = s.Type
	}
	return out
}

func TestParseJSON(t *testing.T) {
	t.Run("Numeric fields become charts, no group", func(t *testing.T) {
		resetFlags(t)
		j := `[{"name":"a","sells":10,"stocks":5,"date":"2024-01"},{"name":"b","sells":20,"stocks":7,"date":"2025-02"}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 2)
		assert.Equal(t, []string{"sells", "stocks"}, statTypes(results[0].Stats))
		assert.Equal(t, 10.0, results[0].Stats[0].Value)
		assert.Equal(t, 5.0, results[0].Stats[1].Value)
		assert.Empty(t, results[0].XAxis)
	})

	t.Run("First-seen column order preserved", func(t *testing.T) {
		resetFlags(t)
		j := `[{"zeta":1,"alpha":2,"mid":3}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, []string{"zeta", "alpha", "mid"}, statTypes(results[0].Stats))
	})

	t.Run("Numeric string parsed", func(t *testing.T) {
		resetFlags(t)
		j := `[{"name":"a","sells":"42"}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
		assert.Equal(t, 42.0, results[0].Stats[0].Value)
	})

	t.Run("Nested object flattened to dotted keys", func(t *testing.T) {
		resetFlags(t)
		j := `[{"name":"a","mem":{"alloc":5,"bytes":100}}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, []string{"mem.alloc", "mem.bytes"}, statTypes(results[0].Stats))
		assert.Equal(t, 5.0, results[0].Stats[0].Value)
	})

	t.Run("Array-valued field skipped", func(t *testing.T) {
		resetFlags(t)
		j := `[{"name":"a","sells":10,"tags":[1,2,3]}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
	})

	t.Run("bool and null skipped", func(t *testing.T) {
		resetFlags(t)
		j := `[{"name":"a","sells":10,"active":true,"note":null}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
	})

	t.Run("Heterogeneous rows: missing key is a gap", func(t *testing.T) {
		resetFlags(t)
		j := `[{"name":"a","sells":10},{"name":"b","stocks":7}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 2)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
		assert.Equal(t, []string{"stocks"}, statTypes(results[1].Stats))
	})

	t.Run("Mixed type per key: numeric where parseable", func(t *testing.T) {
		resetFlags(t)
		// v is a number in row 1, non-numeric string in row 2
		j := `[{"v":3},{"v":"foo"}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		// v qualifies as a chart column (>=1 numeric); row 2 has no stats → dropped
		assert.Len(t, results, 1)
		assert.Equal(t, 3.0, results[0].Stats[0].Value)
	})

	t.Run("Group single field → xAxis", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name"}
		j := `[{"name":"alpha","sells":10},{"name":"beta","sells":20}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 2)
		assert.Equal(t, "alpha", results[0].XAxis)
		assert.Equal(t, "beta", results[1].XAxis)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
	})

	t.Run("Group multi field + -p name/x", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name", "date"}
		shared.FlagState.GroupPattern = "name/x"
		j := `[{"name":"alpha","sells":10,"date":"2024-01"}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, "alpha", results[0].Name)
		assert.Equal(t, "2024-01", results[0].XAxis)
	})

	t.Run("Group on numeric field stringified", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"id"}
		j := `[{"id":7,"sells":10}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, "7", results[0].XAxis)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
	})

	t.Run("Filter regex on group label", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name"}
		shared.FlagState.FilterRegex = "keep"
		j := `[{"name":"keep_a","sells":10},{"name":"drop_b","sells":20},{"name":"keep_c","sells":30}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 2)
		for _, r := range results {
			assert.Contains(t, r.XAxis, "keep")
		}
	})

	t.Run("Number unit scaling", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.NumberUnit = "M"
		j := `[{"name":"a","sells":2000000}]`

		results := ParseJSON(writeJSONTestFile(t, j))

		assert.Len(t, results, 1)
		assert.Equal(t, "sells (M)", results[0].Stats[0].Type)
		assert.Equal(t, 2.0, results[0].Stats[0].Value)
	})

	t.Run("Non-array input returns nil", func(t *testing.T) {
		resetFlags(t)

		assert.Nil(t, ParseJSON(writeJSONTestFile(t, `{"name":"a","sells":10}`)))
		assert.Nil(t, ParseJSON(writeJSONTestFile(t, `[]`)))
		assert.Nil(t, ParseJSON(writeJSONTestFile(t, ``)))
	})
}

func TestParseJSON_FatalErrors(t *testing.T) {
	origOsExit := shared.OsExit
	defer func() { shared.OsExit = origOsExit }()
	shared.OsExit = func(code int) { panic("exit") }

	t.Run("Missing -g field is fatal", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"nope"}
		path := writeJSONTestFile(t, `[{"name":"a","sells":10}]`)

		assert.PanicsWithValue(t, "exit", func() {
			ParseJSON(path)
		})
	})

	t.Run("No numeric fields is fatal", func(t *testing.T) {
		resetFlags(t)
		path := writeJSONTestFile(t, `[{"name":"a","label":"foo"}]`)

		assert.PanicsWithValue(t, "exit", func() {
			ParseJSON(path)
		})
	})
}
