package main

import (
	"time"

	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/logger"
	"upgrade-all-services-cli-plugin/internal/requester"
	"upgrade-all-services-cli-plugin/internal/upgrader"

	"code.cloudfoundry.org/cli/plugin"
)

// upgradeAllServices is a coordination layer that connects the different pieces of this plugin together,
// while delegating the actual work to other packages. Principally it:
// - reads the endpoint and credentials for connecting to CF
// - reads and parses the command line arguments
// - requests the start of the upgrade process with the required data
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

	return upgrader.Upgrade(ccapi.NewCCAPI(reqr), logr, upgrader.UpgradeConfig{
		BrokerName:       cfg.BrokerName,
		ParallelUpgrades: cfg.ParallelUpgrades,
		Action:           cfg.Action,
		MinVersion:       cfg.MinVersion,
		JSONOutput:       cfg.JSONOutput,
		Limit:            cfg.Limit,
		Attempts:         cfg.Attempts,
		RetryInterval:    cfg.RetryInterval,
	})
}
