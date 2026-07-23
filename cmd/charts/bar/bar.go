// Package bar registers the bar chart type: it plugs the typed Config factory
// into the charts registry, stores the flag descriptors, and advertises cobra
// metadata for the `vizb bar` subcommand.
package bar

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	barchart "github.com/goptics/vizb/internal/charts/bar"
)

func init() {
	charts.Register(charts.Spec{Type: "bar", Factory: barchart.New})
	charts.SetFlags("bar", append(slices.Clone(charts.BaseChartFlags),
		charts.ScaleFlag, charts.StackFlag, charts.ThreeDFlag, charts.ThreeDRotateFlag, charts.ThreeDVisualMapFlag,
		charts.HorizontalFlag,
	))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "bar",
		Use:   "bar [target]",
		Short: "Generate a bar chart from data",
		Long: `Generate an interactive bar chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.

Continuous coordinates (value axes) require solo --select, e.g.:
  vizb bar data.csv --select x,y,z -o out.html
  vizb bar data.csv --select x,y,z,value --3d-visualmap -o out.html

All-numeric files with no flags use auto col-axis x (column names as series),
not continuous coordinates. Use -A / --col-axis for series-on-axis without --select.`,
	})
}
