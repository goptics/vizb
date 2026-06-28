package radar

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/config/charts"
	radarchart "github.com/goptics/vizb/config/charts/radar"
)

func init() {
	charts.Register(charts.Spec{Type: "radar", Factory: radarchart.New})
	charts.SetFlags("radar", slices.Clone(charts.BaseChartFlags))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "radar",
		Use:   "radar [target]",
		Short: "Generate a radar chart from data",
		Long:  "Generate an interactive radar chart (HTML or JSON) from benchmark output or tabular CSV/JSON data.",
	})
}
