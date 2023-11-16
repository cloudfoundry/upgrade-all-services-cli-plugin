package main

import (
	"fmt"
	"os"
	"time"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/logger"
	"upgrade-all-services-cli-plugin/internal/requester"
	"upgrade-all-services-cli-plugin/internal/upgrader"

	"code.cloudfoundry.org/cli/plugin"
	"github.com/blang/semver/v4"
)

// version will be set via -ldflags at build time
var version = "0.0.0"

type UpgradePlugin struct{}

func (p *UpgradePlugin) Run(cliConnection plugin.CliConnection, args []string) {
	if args[0] == "upgrade-all-services" {
		if err := upgradeAllServices(cliConnection, args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "upgrade-all-services plugin failed: %s", err.Error())
			os.Exit(1)
		}
	}
}

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

func main() {
	plugin.Start(&UpgradePlugin{})
}

func upgradeAllServices(cliConnection plugin.CliConnection, args []string) error {
	cfg, err := config.ParseConfig(cliConnection, args)
	if err != nil {
		return err
	}

	logr := logger.New(time.Minute)
	reqr := requester.NewRequester(cfg.APIEndpoint, cfg.APIToken, cfg.SkipSSLValidation)
	if cfg.HTTPLogging {
		reqr.Logger = logr
	}

	return upgrader.Upgrade(ccapi.NewCCAPI(reqr), cfg.BrokerName, cfg.ParallelUpgrades, cfg.DryRun, cfg.CheckUpToDate, logr)
}

func pluginVersion() plugin.VersionType {
	v := semver.MustParse(version)
	return plugin.VersionType{
		Major: int(v.Major),
		Minor: int(v.Minor),
		Build: int(v.Patch),
	}
}
