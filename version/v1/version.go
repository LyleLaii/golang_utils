package version

import (
"fmt"
"runtime"
)

// Build information. Populated at build-time.
var (
	Version   string
	Revision  string
	Branch    string
	BuildUser string
	BuildDate string
	GoVersion = runtime.Version()
)

// InfoContext returns version, branch and revision information.
func InfoContext() string {
	return fmt.Sprintf("(version=%s, branch=%s, revision=%s)", Version, Branch, Revision)
}

// BuildContext returns goVersion, buildUser and buildDate information.
func BuildContext() string {
	return fmt.Sprintf("(go=%s, user=%s, date=%s)", GoVersion, BuildUser, BuildDate)
}

func Print() string {
	return fmt.Sprintf("VERSION: %s\nBUILD: %s", InfoContext(), BuildContext())
}