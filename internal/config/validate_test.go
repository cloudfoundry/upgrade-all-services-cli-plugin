package config_test

import (
	"fmt"

	"upgrade-all-services-cli-plugin/internal/command"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/upgrader/upgraderfakes"

	"code.cloudfoundry.org/cli/plugin/pluginfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validate", func() {

	var (
		fakeCliConnection pluginfakes.FakeCliConnection
		fakeLogger        *upgraderfakes.FakeLogger
	)

	BeforeEach(func() {
		fakeCliConnection = pluginfakes.FakeCliConnection{}
		fakeLogger = &upgraderfakes.FakeLogger{}
	})

	Describe("brokername validation", func() {
		When("no broker name is given", func() {
			It("fails to run the upgrade", func() {
				err := config.ValidateInput(&fakeCliConnection, nil)

				Expect(err).To(MatchError(fmt.Errorf("broker name must be specifed\nusage:\ncf upgrade-all-service-instances <broker-name>")))
			})
		})

		When("invalid brokername is given", func() {
			It("returns an error", func() {
				err := config.ValidateInput(&fakeCliConnection, []string{"*inValid'Broker/Name"})

				Expect(err).To(MatchError(fmt.Errorf("invalid brokername format")))
			})
		})
	})

	Describe("validateAPIVersion", func() {
		When("cf api version < 3.99.0", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("2.0.0", nil)

				err := config.ValidateInput(&fakeCliConnection, []string{"broker-name"})

				Expect(err).To(MatchError("plugin requires CF API version >= 3.99.0"))
			})
		})
		When("unable to get API version", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("", fmt.Errorf("ApiVersion error"))

				err := config.ValidateInput(&fakeCliConnection, []string{"broker-name"})

				Expect(err).To(MatchError("error retrieving api version: ApiVersion error"))
			})
		})
	})

	Describe("validateLoggedIn", func() {
		When("not authenticated", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("3.99.0", nil)
				fakeCliConnection.IsLoggedInReturns(false, nil)

				err := command.UpgradeAll(&fakeCliConnection, []string{"broker-name"}, fakeLogger)

				Expect(err).To(MatchError("you must authenticate with the cf cli before running this command"))
			})
		})
		When("unable to check if logged in", func() {
			It("outputs the error", func() {
				fakeCliConnection.ApiVersionReturns("3.99.0", nil)
				fakeCliConnection.IsLoggedInReturns(false, fmt.Errorf("isLoggedIn error"))

				err := command.UpgradeAll(&fakeCliConnection, []string{"broker-name"}, fakeLogger)

				Expect(err).To(MatchError("error validating user authentication: isLoggedIn error"))
			})
		})
	})
})
