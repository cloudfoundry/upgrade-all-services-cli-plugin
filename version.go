package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver/v4"
)

// version will be set via -ldflags at build time
var version = "0.0.0"

func pluginVersion() plugin.VersionType {
	v := semver.MustParse(version)
	return plugin.VersionType{
		Major: int(v.Major),
		Minor: int(v.Minor),
		Build: int(v.Patch),
	}
}
