package cli

import (
	"testing"

	// Chart configs self-register so ChartCommands has specs to build from.
	_ "github.com/goptics/vizb/config/charts/bar"
	_ "github.com/goptics/vizb/config/charts/heatmap"
	_ "github.com/goptics/vizb/config/charts/line"
	_ "github.com/goptics/vizb/config/charts/pie"
	_ "github.com/goptics/vizb/config/charts/radar"
	_ "github.com/goptics/vizb/config/charts/scatter"
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
	for _, c := range ChartCommands() {
		s.byUse[c.Name()] = c
	}
}

func (s *CommandSuite) TestBuildsOneCommandPerChart() {
	for _, name := range []string{"bar", "line", "scatter", "pie", "heatmap", "radar"} {
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

	// scatter is the only chart with the 2D --visualmap flag.
	s.NotNil(s.byUse["scatter"].Flags().Lookup("visualmap"))
}

func TestCommandSuite(t *testing.T) {
	suite.Run(t, new(CommandSuite))
}
