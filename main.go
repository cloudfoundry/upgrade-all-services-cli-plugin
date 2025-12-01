package main

import (
	"os"
	"upgrade-all-services-cli-plugin/internal/config"

	"code.cloudfoundry.org/cli/v8/plugin"
)

func main() {
	plugin.Start(&UpgradePlugin{})
}

type UpgradePlugin struct{}

// Run implements a required method of the code.cloudfoundry.org/cli/plugin.Plugin interface.
// It is the entry point for running the plugin.
func (p *UpgradePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "upgrade-all-services" {
		exitCode := upgradeAllServices(cliConnection, args[1:])
		os.Exit(exitCode)
	}
}

// GetMetadata implements a required method of the code.cloudfoundry.org/cli/plugin.Plugin interface.
// It provides the CF CLI with information about this plugin.
func (p *UpgradePlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:          "UpgradeAllServices",
		Version:       pluginVersion(),
		MinCliVersion: plugin.VersionType{Major: 6, Minor: 53, Build: 0},
		Commands: []plugin.Command{
			{
				Name:     "upgrade-all-services",
				HelpText: "Upgrade all service instances from a broker to the latest available version of their current service plans.",
				UsageDetails: plugin.Usage{
					Usage:   config.Usage,
					Options: config.UsageOptions(),
				},
			},
		},
	}
}
