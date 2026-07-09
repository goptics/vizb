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
		Long:  "Generate an interactive line chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
	})
}
