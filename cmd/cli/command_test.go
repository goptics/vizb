package cli_test

import (
	"testing"

	// Chart configs self-register so ChartCommands has specs to build from.
	_ "github.com/goptics/vizb/cmd/charts/bar"
	_ "github.com/goptics/vizb/cmd/charts/heatmap"
	_ "github.com/goptics/vizb/cmd/charts/line"
	_ "github.com/goptics/vizb/cmd/charts/pie"
	_ "github.com/goptics/vizb/cmd/charts/radar"
	_ "github.com/goptics/vizb/cmd/charts/sankey"
	_ "github.com/goptics/vizb/cmd/charts/scatter"
	"github.com/goptics/vizb/cmd/cli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// CommandSuite covers the generic chart-command builder.
type CommandSuite struct {
	suite.Suite
	byUse map[string]*cobra.Command
}

func (s *CommandSuite) SetupTest() {
	s.byUse = map[string]*cobra.Command{}
	for _, c := range cli.ChartCommands() {
		s.byUse[c.Name()] = c
	}
}

func (s *CommandSuite) TestBuildsOneCommandPerChart() {
	for _, name := range []string{"bar", "line", "scatter", "pie", "heatmap", "radar", "sankey"} {
		s.Contains(s.byUse, name, "missing %s subcommand", name)
	}
}

func (s *CommandSuite) TestVariableFlagsBoundPerChart() {
	// bar carries --scale (variable) and --swap (universal); pie carries neither
	// --scale nor --visualmap.
	bar := s.byUse["bar"]
	s.NotNil(bar.Flags().Lookup("scale"))
	s.NotNil(bar.Flags().Lookup("swap"))
	s.Nil(bar.Flags().Lookup("visualmap"))

	pie := s.byUse["pie"]
	s.Nil(pie.Flags().Lookup("scale"))
	s.NotNil(pie.Flags().Lookup("swap"))

	// sankey uses BaseChartFlags only (no scale/stack/3d/visualMap).
	sankey := s.byUse["sankey"]
	s.Nil(sankey.Flags().Lookup("scale"))
	s.Nil(sankey.Flags().Lookup("stack"))
	s.Nil(sankey.Flags().Lookup("3d"))
	s.Nil(sankey.Flags().Lookup("visualmap"))
	s.NotNil(sankey.Flags().Lookup("swap"))

	// scatter is the only chart with the 2D --visualmap flag.
	s.NotNil(s.byUse["scatter"].Flags().Lookup("visualmap"))
}

func TestCommandSuite(t *testing.T) {
	suite.Run(t, new(CommandSuite))
}
