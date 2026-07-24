package updater

import (
	"runtime/debug"
	"strings"

	"golang.org/x/mod/semver"
)

const modulePath = "github.com/goptics/vizb"

type installationKind string

const (
	installationUnknown    installationKind = "unknown"
	installationStandalone installationKind = "standalone"
	installationManaged    installationKind = "managed"
)

type buildInfo struct {
	MainPath    string
	MainVersion string
}

type detectionContext struct {
	GOOS                string
	RawExecutable       string
	CanonicalExecutable string
	Distribution        string
	Build               buildInfo
}

type installation struct {
	Kind        installationKind
	Manager     string
	Instruction string
}

type installationDetector interface {
	Detect(detectionContext) (installation, bool)
}

type detectorFunc func(detectionContext) (installation, bool)

func (f detectorFunc) Detect(ctx detectionContext) (installation, bool) {
	return f(ctx)
}

var installationDetectors = []installationDetector{
	detectorFunc(detectWinGet),
	detectorFunc(detectGoToolchain),
	detectorFunc(detectDistribution),
}

func detectInstallation(ctx detectionContext) installation {
	for _, detector := range installationDetectors {
		if found, ok := detector.Detect(ctx); ok {
			return found
		}
	}
	return installation{Kind: installationUnknown}
}

func detectWinGet(ctx detectionContext) (installation, bool) {
	if ctx.GOOS != "windows" {
		return installation{}, false
	}

	for _, executable := range []string{ctx.RawExecutable, ctx.CanonicalExecutable} {
		path := strings.ToLower(strings.ReplaceAll(executable, `\`, "/"))
		if strings.Contains(path, "/microsoft/winget/links/") ||
			strings.Contains(path, "/microsoft/winget/packages/") {
			return installation{
				Kind:        installationManaged,
				Manager:     "WinGet",
				Instruction: "winget upgrade --id goptics.vizb --exact",
			}, true
		}
	}

	return installation{}, false
}

func detectGoToolchain(ctx detectionContext) (installation, bool) {
	if ctx.Distribution != "" || ctx.Build.MainPath != modulePath {
		return installation{}, false
	}

	buildVersion := normalizeVersion(ctx.Build.MainVersion)
	if !semver.IsValid(buildVersion) {
		return installation{}, false
	}

	return installation{
		Kind:        installationManaged,
		Manager:     "Go toolchain",
		Instruction: "go install github.com/goptics/vizb@latest",
	}, true
}

func detectDistribution(ctx detectionContext) (installation, bool) {
	distribution := strings.TrimSpace(ctx.Distribution)
	if distribution == "" {
		return installation{}, false
	}
	if distribution == "standalone" {
		return installation{Kind: installationStandalone, Manager: "Vizb installer"}, true
	}

	return installation{Kind: installationManaged, Manager: distribution}, true
}

func readBuildInfo() buildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return buildInfo{}
	}
	return buildInfo{MainPath: info.Main.Path, MainVersion: info.Main.Version}
}
