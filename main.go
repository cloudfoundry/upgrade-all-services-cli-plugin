package main

import (
	"os"
	"time"
	"upgrade-all-services-cli-plugin/internal/command"
	"upgrade-all-services-cli-plugin/internal/logger"
	"upgrade-all-services-cli-plugin/internal/validate"

	"code.cloudfoundry.org/cli/plugin"
)

type UpgradePlugin struct{}

func (p *UpgradePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "upgrade-all-services" {
		l := logger.New(time.Minute)
		err := command.UpgradeAll(cliConnection, args[1:], l)
		if err != nil {
			l.Printf("upgrade-all-services plugin failed: %s", err.Error())
			os.Exit(1)
		}
	}
}

func (p *UpgradePlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:          "UpgradeAllServices",
		Version:       plugin.VersionType{Major: 0, Minor: 1, Build: 0},
		MinCliVersion: plugin.VersionType{Major: 6, Minor: 53, Build: 0},
		Commands: []plugin.Command{
			{
				Name:     "upgrade-all-services",
				HelpText: "Upgrade all service instances from a broker to the latest available version of their current service plans.",
				UsageDetails: plugin.Usage{
					Usage: validate.Usage,
					Options: map[string]string{
						"-batch-size": "The number of concurrent upgrades (defaults to 10)",
					},
				},
			},
		},
	}
}

func main() {
	plugin.Start(&UpgradePlugin{})
}
