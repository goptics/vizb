package sankey

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	sankeychart "github.com/goptics/vizb/internal/charts/sankey"
)

func init() {
	charts.Register(charts.Spec{Type: "sankey", Factory: sankeychart.New})
	charts.SetFlags("sankey", slices.Clone(charts.BaseChartFlags))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  "sankey",
		Use:   "sankey [target]",
		Short: "Generate a sankey chart from data",
		Long:  "Generate an interactive sankey chart (HTML or JSON) from benchmark output or tabular CSV/JSON data. Axis mapping: source=x, target=y, value=active stat; n (name) panels are supported. z is ignored for layout.",
	})
}
