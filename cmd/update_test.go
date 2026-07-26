package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UpdateCommandSuite struct {
	suite.Suite
}

func (s *UpdateCommandSuite) TestRegisteredOnRootCommand() {
	command, _, err := rootCmd.Find([]string{"update"})
	s.Require().NoError(err)
	s.Same(updateCmd, command)
}

func (s *UpdateCommandSuite) TestRunsWithCommandStreams() {
	called := false
	runner := func(_ context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
		called = true
		s.NotNil(stdin)
		_, _ = io.WriteString(stdout, "updated")
		_, _ = io.WriteString(stderr, "warning")
		return nil
	}
	command := newUpdateCommand(runner)
	var stdout, stderr bytes.Buffer
	command.SetIn(bytes.NewBufferString("input"))
	command.SetOut(&stdout)
	command.SetErr(&stderr)

	s.Require().NoError(command.Execute())
	s.True(called)
	s.Equal("updated", stdout.String())
	s.Equal("warning", stderr.String())
}

func (s *UpdateCommandSuite) TestRejectsArgumentsWithoutRunning() {
	called := false
	command := newUpdateCommand(func(context.Context, io.Reader, io.Writer, io.Writer) error {
		called = true
		return nil
	})
	command.SetArgs([]string{"v1.2.3"})

	s.Require().ErrorContains(command.Execute(), "unknown command")
	s.False(called)
}

func (s *UpdateCommandSuite) TestReturnsRunnerError() {
	want := errors.New("update failed")
	command := newUpdateCommand(func(context.Context, io.Reader, io.Writer, io.Writer) error {
		return want
	})

	s.ErrorIs(command.Execute(), want)
}

func TestUpdateCommandSuite(t *testing.T) {
	suite.Run(t, new(UpdateCommandSuite))
}
