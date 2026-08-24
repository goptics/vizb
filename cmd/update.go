package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/goptics/vizb/cmd/cli"
	"github.com/goptics/vizb/internal/updater"
	"github.com/goptics/vizb/version"
	"github.com/spf13/cobra"
	"golang.org/x/mod/module"
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

const (
	updateDemoBytes = 8 << 20
	updateDemoEvery = 512 << 10
	updateDemoPace  = 80 * time.Millisecond
)

func runUpdate(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	service, err := updater.New(version.Version, version.Distribution)
	if err != nil {
		return err
	}
	bar := cli.NewDownloadBar(stderr)
	defer func() { _ = bar.Finish() }()
	if isDevBuild(version.Version, version.Distribution) && bar.IsTTY() {
		return previewUpdateDownload(bar, stdout, updateDemoBytes, updateDemoPace)
	}
	service.ProgressWrap = bar.Wrap
	return service.Run(ctx, stdin, stdout, stderr)
}

func isDevBuild(ver, distribution string) bool {
	if strings.TrimSpace(distribution) != "" {
		return false
	}
	v := strings.TrimSpace(ver)
	switch v {
	case "", "devel", "(devel)":
		return true
	}
	// Local `go build` / `task build:cli` often VCS-stamps a pseudo-version
	// and/or +dirty. Release and `go install` stay on a clean tag.
	if strings.Contains(v, "+dirty") {
		return true
	}
	base, _, _ := strings.Cut(v, "+")
	return module.IsPseudoVersion(base)
}

func previewUpdateDownload(bar *cli.DownloadBar, stdout io.Writer, total int64, pace time.Duration) error {
	src := bar.Wrap(newPacedReader(total, pace), total, "(devel)")
	_, err := io.Copy(io.Discard, src)
	_ = bar.Finish()
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, "Development build: download progress demo only; binary not updated.")
	return nil
}

type pacedReader struct {
	remain     int64
	pace       time.Duration
	sinceSleep int64
}

func newPacedReader(total int64, pace time.Duration) *pacedReader {
	return &pacedReader{remain: total, pace: pace}
}

func (p *pacedReader) Read(b []byte) (int, error) {
	if p.remain <= 0 {
		return 0, io.EOF
	}
	n := len(b)
	if int64(n) > p.remain {
		n = int(p.remain)
	}
	p.remain -= int64(n)
	p.sinceSleep += int64(n)
	if p.pace > 0 && p.sinceSleep >= updateDemoEvery && p.remain > 0 {
		time.Sleep(p.pace)
		p.sinceSleep = 0
	}
	return n, nil
}
