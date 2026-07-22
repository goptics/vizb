// Package updater detects how the running Vizb binary was installed and safely
// updates official standalone releases.
package updater

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/goptics/vizb/shared"
	"golang.org/x/mod/semver"
)

const (
	defaultLatestURL          = "https://github.com/goptics/vizb/releases/latest"
	defaultReleaseDownloadURL = "https://github.com/goptics/vizb/releases/download"
)

type windowsInstaller func(context.Context, string, string, io.Reader, io.Writer, io.Writer) error

//go:embed install.ps1
var embeddedWindowsInstaller []byte

// Updater contains the runtime and release dependencies for one update run.
type Updater struct {
	currentVersion      string
	distribution        string
	rawExecutable       string
	canonicalExecutable string
	build               buildInfo
	goos                string
	goarch              string
	client              *http.Client
	latestURL           string
	releaseDownloadURL  string
	replace             func(string, string) error
	installOnWindows    windowsInstaller
}

// New constructs an updater for the currently running executable.
func New(currentVersion, distribution string) (*Updater, error) {
	executable, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locate running vizb executable: %w", err)
	}
	canonical, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return nil, fmt.Errorf("resolve running vizb executable %q: %w", executable, err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute vizb executable path: %w", err)
	}

	return &Updater{
		currentVersion:      currentVersion,
		distribution:        distribution,
		rawExecutable:       executable,
		canonicalExecutable: canonical,
		build:               readBuildInfo(),
		goos:                runtime.GOOS,
		goarch:              runtime.GOARCH,
		client:              &http.Client{Timeout: 5 * time.Minute},
		latestURL:           defaultLatestURL,
		releaseDownloadURL:  defaultReleaseDownloadURL,
		replace:             replaceExecutable,
		installOnWindows:    runWindowsInstaller,
	}, nil
}

// Run updates a standalone installation or prints the package-manager command
// that owns the running executable.
func (u *Updater) Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) error {
	installedBy := detectInstallation(detectionContext{
		GOOS:                u.goos,
		RawExecutable:       u.rawExecutable,
		CanonicalExecutable: u.canonicalExecutable,
		Distribution:        u.distribution,
		Build:               u.build,
	})

	switch installedBy.Kind {
	case installationManaged:
		if installedBy.Instruction == "" {
			return fmt.Errorf("vizb is managed by %s; update it with that package manager", installedBy.Manager)
		}
		fmt.Fprintf(stdout, "vizb is managed by %s.\nUpdate with: %s\n", installedBy.Manager, installedBy.Instruction)
		return nil
	case installationUnknown:
		return fmt.Errorf("cannot safely update this development or unidentified vizb build; reinstall from https://vizb.goptics.org/installation")
	}

	currentVersion := normalizeVersion(u.currentVersion)
	if !semver.IsValid(currentVersion) {
		return fmt.Errorf("running vizb has invalid semantic version %q", u.currentVersion)
	}
	if _, err := assetFor(currentVersion, u.goos, u.goarch); err != nil {
		return err
	}

	latestVersion, err := latestReleaseVersion(ctx, u.client, u.latestURL, u.userAgent())
	if err != nil {
		return err
	}

	comparison := semver.Compare(currentVersion, latestVersion)
	if comparison == 0 {
		fmt.Fprintf(stdout, "vizb %s is already up to date.\n", currentVersion)
		return nil
	}
	if comparison > 0 {
		fmt.Fprintf(stdout, "vizb %s is newer than the latest release %s; no downgrade performed.\n", currentVersion, latestVersion)
		return nil
	}

	if u.goos == "windows" {
		fmt.Fprintf(stdout, "Updating vizb from %s to %s through the Windows installer...\n", currentVersion, latestVersion)
		if err := u.installOnWindows(ctx, latestVersion, u.canonicalExecutable, stdin, stdout, stderr); err != nil {
			return fmt.Errorf("run Windows installer: %w", err)
		}
		fmt.Fprintf(stdout, "Updated vizb to %s.\n", latestVersion)
		return nil
	}

	candidate, cleanup, err := u.prepareCandidate(ctx, latestVersion)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := u.replace(candidate, u.canonicalExecutable); err != nil {
		return fmt.Errorf("replace %s: %w", u.canonicalExecutable, err)
	}
	fmt.Fprintf(stdout, "Updated vizb from %s to %s.\n", currentVersion, latestVersion)
	return nil
}

func (u *Updater) prepareCandidate(ctx context.Context, version string) (string, func(), error) {
	asset, err := assetFor(version, u.goos, u.goarch)
	if err != nil {
		return "", func() {}, err
	}

	tempDir, err := os.MkdirTemp("", "vizb-update-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("create update workspace: %w", err)
	}
	shared.TempFiles.Store(tempDir)
	cleanup := func() { _ = os.RemoveAll(tempDir) }

	checksumsPath := filepath.Join(tempDir, "checksums.txt")
	checksumsURL := u.releaseURL(version, "checksums.txt")
	if err := downloadFile(ctx, u.client, checksumsURL, checksumsPath, u.userAgent(), maxChecksumSize); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("download release checksums: %w", err)
	}
	expected, err := expectedChecksum(checksumsPath, asset.Name)
	if err != nil {
		cleanup()
		return "", func() {}, err
	}

	archivePath := filepath.Join(tempDir, asset.Name)
	if err := downloadFile(ctx, u.client, u.releaseURL(version, asset.Name), archivePath, u.userAgent(), maxArchiveSize); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("download release archive: %w", err)
	}
	if err := verifyChecksum(archivePath, expected); err != nil {
		cleanup()
		return "", func() {}, err
	}

	binaryName := "vizb"
	if u.goos == "windows" {
		binaryName += ".exe"
	}
	candidate := filepath.Join(tempDir, binaryName)
	if err := extractBinary(archivePath, asset.Extension, candidate, binaryName); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return candidate, cleanup, nil
}

func (u *Updater) releaseURL(version, asset string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimRight(u.releaseDownloadURL, "/"), version, asset)
}

func (u *Updater) userAgent() string {
	return "vizb/" + normalizeVersion(u.currentVersion)
}

func replaceExecutable(candidate, target string) error {
	targetInfo, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("inspect running executable: %w", err)
	}
	if !targetInfo.Mode().IsRegular() {
		return fmt.Errorf("running executable is not a regular file")
	}

	input, err := os.Open(candidate)
	if err != nil {
		return fmt.Errorf("open verified replacement: %w", err)
	}
	defer input.Close()

	staged, err := os.CreateTemp(filepath.Dir(target), ".vizb-update-*")
	if err != nil {
		return fmt.Errorf("stage replacement beside executable: %w", err)
	}
	stagedPath := staged.Name()
	shared.TempFiles.Store(stagedPath)
	defer os.Remove(stagedPath)

	if _, err := io.Copy(staged, input); err != nil {
		_ = staged.Close()
		return fmt.Errorf("stage replacement: %w", err)
	}
	if err := staged.Chmod(targetInfo.Mode().Perm()); err != nil {
		_ = staged.Close()
		return fmt.Errorf("set replacement permissions: %w", err)
	}
	if err := staged.Sync(); err != nil {
		_ = staged.Close()
		return fmt.Errorf("sync replacement: %w", err)
	}
	if err := staged.Close(); err != nil {
		return fmt.Errorf("close replacement: %w", err)
	}
	if err := os.Rename(stagedPath, target); err != nil {
		return fmt.Errorf("atomically install replacement: %w", err)
	}
	return nil
}

func runWindowsInstaller(ctx context.Context, version, executable string, stdin io.Reader, stdout, stderr io.Writer) error {
	script, err := os.CreateTemp("", "vizb-update-*.ps1")
	if err != nil {
		return fmt.Errorf("create temporary Windows installer: %w", err)
	}
	scriptPath := script.Name()
	shared.TempFiles.Store(scriptPath)
	defer os.Remove(scriptPath)

	if _, err := script.Write(embeddedWindowsInstaller); err != nil {
		_ = script.Close()
		return fmt.Errorf("write temporary Windows installer: %w", err)
	}
	if err := script.Sync(); err != nil {
		_ = script.Close()
		return fmt.Errorf("sync temporary Windows installer: %w", err)
	}
	if err := script.Close(); err != nil {
		return fmt.Errorf("close temporary Windows installer: %w", err)
	}

	command := exec.CommandContext(
		ctx,
		"powershell.exe",
		"-NoProfile",
		"-NonInteractive",
		"-ExecutionPolicy",
		"Bypass",
		"-File",
		scriptPath,
		"-VersionTag",
		version,
		"-SourceExecutable",
		executable,
	)
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}
