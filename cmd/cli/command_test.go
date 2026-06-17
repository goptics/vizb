package cli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

// CommandSuite covers the chart command registry.
type CommandSuite struct {
	suite.Suite
}

func (s *CommandSuite) TestRegisterAndCommands() {
	before := len(Commands())

	Register(func() *cobra.Command {
		return &cobra.Command{Use: "test-cmd"}
	})

	cmds := Commands()
	s.Greater(len(cmds), before)
	s.Equal("test-cmd", cmds[len(cmds)-1].Use)
}

func TestCommandSuite(t *testing.T) {
	suite.Run(t, new(CommandSuite))
}
