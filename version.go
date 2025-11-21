package main

import (
	"code.cloudfoundry.org/cli/v8/plugin"
	"github.com/blang/semver/v4"
)

// version will be set via -ldflags at build time
var version = "0.0.0"

func pluginVersion() plugin.VersionType {
	// NOTE: we use library "github.com/hashicorp/go-version" elsewhere, but it doesn't provide the
	// ability to easily parse out major/minor/fix, so we use "github.com/blang/semver/v4" here
	// and only here
	v := semver.MustParse(version)
	return plugin.VersionType{
		Major: int(v.Major),
		Minor: int(v.Minor),
		Build: int(v.Patch),
	}
}
