package line

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	linechart "github.com/goptics/vizb/internal/charts/line"
)

func init() {
	charts.Register(charts.Spec{Type: "line", Factory: linechart.New})
	charts.SetFlags("line", append(slices.Clone(charts.BaseChartFlags),
		charts.ScaleFlag, charts.StackFlag, charts.ThreeDFlag, charts.ThreeDRotateFlag, charts.ThreeDVisualMapFlag,
		charts.SymbolFlag, charts.SymbolSizeFlag, charts.SmoothFlag,
	))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "line",
		Use:   "line [target]",
		Short: "Generate a line chart from data",
		Long: `Generate an interactive line chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.

Continuous coordinates (value axes) require solo --select, e.g.:
  vizb line data.csv --select x,y,z -o out.html

All-numeric files with no flags use auto col-axis x (column names as series),
not continuous coordinates. Use -A / --col-axis for series-on-axis without --select.`,
	})
}
