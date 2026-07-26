package cmd

import (
	"context"
	"io"

	"github.com/goptics/vizb/internal/updater"
	"github.com/goptics/vizb/version"
	"github.com/spf13/cobra"
)

type updateRunner func(context.Context, io.Reader, io.Writer, io.Writer) error

var updateCmd = newUpdateCommand(runUpdate)

func init() {
	rootCmd.AddCommand(updateCmd)
}

func newUpdateCommand(run updateRunner) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update vizb to the latest release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

func runUpdate(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	service, err := updater.New(version.Version, version.Distribution)
	if err != nil {
		return err
	}
	return service.Run(ctx, stdin, stdout, stderr)
}
