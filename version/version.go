package version

import "runtime/debug"

var (
	Version      = "devel"
	Distribution string
)

func init() {
	if Version != "devel" {
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		Version = info.Main.Version
	}
}
