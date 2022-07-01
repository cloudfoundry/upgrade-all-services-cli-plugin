package main

import (
	"os"
	"time"
	"upgrade-all-services-cli-plugin/internal/command"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/logger"

	"code.cloudfoundry.org/cli/plugin"
)

type UpgradePlugin struct{}

func (p *UpgradePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	l := logger.New(time.Minute)
	if err := command.UpgradeAll(cliConnection, args[1:], l); err != nil {
		l.Printf("upgrade-all-services plugin failed: %s", err.Error())
		os.Exit(1)
	}
}

func (p *UpgradePlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:          "UpgradeAllServiceInstances",
		Version:       plugin.VersionType{Major: 0, Minor: 1, Build: 0},
		MinCliVersion: plugin.VersionType{Major: 6, Minor: 53, Build: 0},
		Commands: []plugin.Command{
			{
				Name:         "upgrade-all-services",
				HelpText:     "Upgrade all service instances from a broker to the latest available version of their current service plans.",
				UsageDetails: plugin.Usage{Usage: config.Help()},
			},
		},
	}
}

func main() {
	plugin.Start(&UpgradePlugin{})
}
