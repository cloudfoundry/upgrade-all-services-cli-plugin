package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/logger"
	"upgrade-all-services-cli-plugin/internal/requester"
	"upgrade-all-services-cli-plugin/internal/upgrader"

	"code.cloudfoundry.org/cli/v8/plugin"
)

// The CF CLI doesn't maintain exit codes, so we can't pass on any information with them
const (
	exitSuccess = 0
	exitError   = 1
)

// upgradeAllServices is a coordination layer that connects the different pieces of this plugin together,
// while delegating the actual work to other packages. Principally it:
// - reads the endpoint and credentials for connecting to CF
// - reads and parses the command line arguments
// - requests the start of the upgrade process with the required data
func upgradeAllServices(cliConnection plugin.CliConnection, args []string) int {
	cfg, err := config.ParseConfig(cliConnection, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "upgrade-all-services plugin failed: %s", err)
		return exitError
	}

	logr := logger.New(time.Minute)
	reqr := requester.NewRequester(cfg.APIEndpoint, cfg.APIToken, cfg.SkipSSLValidation)
	if cfg.HTTPLogging {
		reqr.Logger = logr
	}

	err = upgrader.Upgrade(ccapi.NewCCAPI(reqr, cfg.InstancePollingInterval), logr, upgrader.UpgradeConfig{
		BrokerName:       cfg.BrokerName,
		ParallelUpgrades: cfg.ParallelUpgrades,
		Action:           cfg.Action,
		MinVersion:       cfg.MinVersion,
		JSONOutput:       cfg.JSONOutput,
		Limit:            cfg.Limit,
		Attempts:         cfg.Attempts,
		RetryInterval:    cfg.RetryInterval,
	})

	isInstanceError := errors.As(err, &upgrader.InstanceError{})

	switch {
	case err == nil, isInstanceError && cfg.IgnoreInstanceErrors:
		return exitSuccess
	case isInstanceError && cfg.JSONOutput:
		// We don't pollute the JSON with an error message in case STDOUT and STDERR streams are merged
		return exitError
	case isInstanceError:
		fmt.Fprintf(os.Stderr, "upgrade-all-services plugin failed: %s", err)
		return exitError
	default:
		fmt.Fprintf(os.Stderr, "upgrade-all-services plugin error: %s", err)
		return exitError
	}
}
