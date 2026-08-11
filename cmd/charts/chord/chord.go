package chord

import (
	"slices"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/charts"
	chordchart "github.com/goptics/vizb/internal/charts/chord"
)

func init() {
	charts.Register(charts.Spec{Type: chordchart.Type, Factory: chordchart.New})
	charts.SetFlags(chordchart.Type, slices.Clone(charts.BaseChartFlags))
	cli.SetChartMeta(cli.ChartMeta{
		Type:  chordchart.Type,
		Use:   "chord [target]",
		Short: "Generate a chord chart from data",
		Long:  "Generate an interactive chord chart (HTML or JSON) from benchmark output or tabular CSV/JSON data. Axis mapping: source=x, target=y, value=active stat; n (name) panels are supported. z is ignored for layout.",
	})
}
