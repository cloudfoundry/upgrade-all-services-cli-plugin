package command

import (
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/requester"
	"upgrade-all-services-cli-plugin/internal/upgrader"

	"code.cloudfoundry.org/cli/plugin"
)

func UpgradeAll(cliConnection plugin.CliConnection, args []string, log upgrader.Logger) error {
	c, err := config.ParseConfig(cliConnection, args)
	if err != nil {
		return err
	}

	r := requester.NewRequester(c.APIEndpoint, c.AccessToken, c.SkipSSLValidation)
	api := ccapi.NewCCAPI(r)

	if err := upgrader.Upgrade(api, c.BrokerName, c.ParallelUpgrades, log); err != nil {
		return err
	}

	return nil
}
