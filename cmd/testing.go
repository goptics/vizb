package cmd

import "github.com/goptics/vizb/shared"

// ResetTestState restores root, ui, and merge flag globals to their defaults.
// Chart slices get a fresh copy so tests do not alias the package-level defaults.
// Tests that pass explicit -c should set rootOpts.Charts = nil before Execute so
// cobra replaces the slice instead of appending to the reset copy.
// Test-only entrypoint — production code must not call this.
func ResetTestState() {
	rootOpts.Name = ""
	rootOpts.Description = ""
	rootOpts.Tag = ""
	rootOpts.OutputFile = ""
	rootOpts.Parser = "auto"
	rootOpts.Group = nil
	rootOpts.GroupPattern = "x"
	rootOpts.GroupRegex = ""
	rootOpts.Filter = ""
	rootOpts.MemUnit = "B"
	rootOpts.TimeUnit = "ns"
	rootOpts.NumberUnit = ""
	rootOpts.Sort = ""
	rootOpts.ShowLabels = false
	rootOpts.ChartSpecs = nil
	rootOpts.Charts = append([]string(nil), defaultChartTypes...)

	uiOpts.OutputFile = ""
	uiOpts.ChartSpecs = nil
	uiOpts.DataURL = ""
	uiOpts.Enable3D = false
	uiOpts.Charts = append([]string(nil), shared.DefaultChartTypes...)

	mergeOpts.OutputFile = ""
	mergeOpts.TagAxis = "n"
}
