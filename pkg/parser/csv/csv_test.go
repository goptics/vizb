package csv

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goptics/vizb/shared"
	"github.com/stretchr/testify/assert"
)

func writeCSVTestFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.csv")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// resetFlags clears the CSV-relevant FlagState fields and restores them after the test.
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

func TestParseCSV(t *testing.T) {
	t.Run("Numeric columns become charts, no group", func(t *testing.T) {
		resetFlags(t)
		csv := "name,sells,stocks,date\na,10,5,2024-01\nb,20,7,2025-02\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 2)
		assert.Equal(t, []string{"sells", "stocks"}, statTypes(results[0].Stats))
		assert.Equal(t, 10.0, results[0].Stats[0].Value)
		assert.Equal(t, 5.0, results[0].Stats[1].Value)
		// no -g → empty labels
		assert.Empty(t, results[0].Name)
		assert.Empty(t, results[0].XAxis)
		assert.Empty(t, results[0].YAxis)
	})

	t.Run("Group single column → xAxis (default pattern)", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name"}
		csv := "name,sells,date\nalpha,10,2024-01\nbeta,20,2025-02\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 2)
		assert.Equal(t, "alpha", results[0].XAxis)
		assert.Equal(t, "beta", results[1].XAxis)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
	})

	t.Run("Group multi column joined and routed by -p name/x", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name", "date"}
		shared.FlagState.GroupPattern = "name/x"
		csv := "name,sells,date\nalpha,10,2024-01\nbeta,20,2025-02\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 2)
		assert.Equal(t, "alpha", results[0].Name)
		assert.Equal(t, "2024-01", results[0].XAxis)
	})

	t.Run("Group column excluded from charts even if numeric", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"id"}
		csv := "id,sells\n1,10\n2,20\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 2)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
		assert.Equal(t, "1", results[0].XAxis)
	})

	t.Run("Any-one-parses: stray number makes a junk chart column", func(t *testing.T) {
		resetFlags(t)
		csv := "name,mostlytext\na,hello\nb,42\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		// mostlytext qualifies as a chart column (>=1 numeric cell);
		// row a has no numeric cell → dropped, row b kept.
		assert.Len(t, results, 1)
		assert.Equal(t, []string{"mostlytext"}, statTypes(results[0].Stats))
		assert.Equal(t, 42.0, results[0].Stats[0].Value)
	})

	t.Run("NaN and Inf cells skipped", func(t *testing.T) {
		resetFlags(t)
		csv := "name,v\na,NaN\nb,Inf\nc,3\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		// only c has a finite value
		assert.Len(t, results, 1)
		assert.Equal(t, 3.0, results[0].Stats[0].Value)
	})

	t.Run("Pure non-numeric column ignored", func(t *testing.T) {
		resetFlags(t)
		csv := "label,sells\nfoo,10\nbar,20\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 2)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
	})

	t.Run("BOM stripped from first header", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name"}
		csv := "\ufeffname,sells\nalpha,10\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 1)
		assert.Equal(t, "alpha", results[0].XAxis)
	})

	t.Run("Whitespace trimmed in headers and group values", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name"}
		csv := " name , sells \n alpha , 10 \n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 1)
		assert.Equal(t, "alpha", results[0].XAxis)
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
		assert.Equal(t, 10.0, results[0].Stats[0].Value)
	})

	t.Run("Ragged rows tolerated", func(t *testing.T) {
		resetFlags(t)
		csv := "name,sells,stocks\na,10\nb,20,7\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 2)
		// row a missing stocks cell → only sells stat
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
		assert.Equal(t, []string{"sells", "stocks"}, statTypes(results[1].Stats))
	})

	t.Run("Duplicate headers suffixed", func(t *testing.T) {
		resetFlags(t)
		csv := "sells,sells\n10,20\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 1)
		assert.Equal(t, []string{"sells", "sells (2)"}, statTypes(results[0].Stats))
	})

	t.Run("Empty header column ignored", func(t *testing.T) {
		resetFlags(t)
		csv := "name,,sells\na,99,10\n"
		shared.FlagState.Group = []string{"name"}

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 1)
		// the empty-named column (value 99) is not charted
		assert.Equal(t, []string{"sells"}, statTypes(results[0].Stats))
	})

	t.Run("Empty -g entry filtered out", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name", ""}
		csv := "name,sells\nalpha,10\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 1)
		assert.Equal(t, "alpha", results[0].XAxis)
	})

	t.Run("Filter regex on group label", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"name"}
		shared.FlagState.FilterRegex = "keep"
		csv := "name,sells\nkeep_a,10\ndrop_b,20\nkeep_c,30\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 2)
		for _, r := range results {
			assert.Contains(t, r.XAxis, "keep")
		}
	})

	t.Run("Number unit scaling", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.NumberUnit = "M"
		csv := "name,sells\na,2000000\n"

		results := ParseCSV(writeCSVTestFile(t, csv))

		assert.Len(t, results, 1)
		assert.Equal(t, "sells (M)", results[0].Stats[0].Type)
		assert.Equal(t, 2.0, results[0].Stats[0].Value)
	})

	t.Run("Less than two rows returns nil", func(t *testing.T) {
		resetFlags(t)

		assert.Nil(t, ParseCSV(writeCSVTestFile(t, "name,sells\n")))
		assert.Nil(t, ParseCSV(writeCSVTestFile(t, "")))
	})
}

func TestParseCSV_FatalErrors(t *testing.T) {
	origOsExit := shared.OsExit
	defer func() { shared.OsExit = origOsExit }()
	shared.OsExit = func(code int) { panic("exit") }

	t.Run("Missing -g column is fatal", func(t *testing.T) {
		resetFlags(t)
		shared.FlagState.Group = []string{"nope"}
		csv := "name,sells\na,10\n"
		path := writeCSVTestFile(t, csv)

		assert.PanicsWithValue(t, "exit", func() {
			ParseCSV(path)
		})
	})

	t.Run("No numeric columns is fatal", func(t *testing.T) {
		resetFlags(t)
		csv := "name,label\na,foo\nb,bar\n"
		path := writeCSVTestFile(t, csv)

		assert.PanicsWithValue(t, "exit", func() {
			ParseCSV(path)
		})
	})
}
