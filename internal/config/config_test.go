package config_test

import (
	"fmt"

	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/config/configfakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		fakeCLIConnection *configfakes.FakeCLIConnection
		fakeArgs          []string
		cfg               config.Config
		cfgErr            error
	)

	BeforeEach(func() {
		fakeCLIConnection = &configfakes.FakeCLIConnection{}
		fakeCLIConnection.IsLoggedInReturns(true, nil)
		fakeCLIConnection.ApiVersionReturns("3.9999.0", nil)

		fakeArgs = []string{"fake-broker-name"}
	})

	JustBeforeEach(func() {
		cfg, cfgErr = config.ParseConfig(fakeCLIConnection, fakeArgs)
	})

	Describe("checking logged in", func() {
		When("logged in", func() {
			BeforeEach(func() {
				fakeCLIConnection.IsLoggedInReturns(true, nil)
			})

			It("succeeds", func() {
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})

		When("not logged in", func() {
			BeforeEach(func() {
				fakeCLIConnection.IsLoggedInReturns(false, nil)
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError("you must authenticate with the cf cli before running this command"))
			})
		})

		When("error getting logged in status", func() {
			BeforeEach(func() {
				fakeCLIConnection.IsLoggedInReturns(false, fmt.Errorf("boom"))
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError("error getting login status: boom"))
			})
		})
	})

	Describe("validating API version", func() {
		When("API version is valid v3", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiVersionReturns("3.99.0", nil)
			})

			It("succeeds", func() {
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})

		When("API version is too low v3", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiVersionReturns("3.98.0", nil)
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`plugin requires minimum API version 3.99.0 or 2.164.0, got "3.98.0"`))
			})
		})

		When("API version is valid v2", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiVersionReturns("2.164.0", nil)
			})

			It("succeeds", func() {
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})

		When("API version is too low v2", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiVersionReturns("2.163.0", nil)
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`plugin requires minimum API version 3.99.0 or 2.164.0, got "2.163.0"`))
			})
		})

		When("API version is too high", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiVersionReturns("4", nil)
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`plugin requires minimum API version 3.99.0 or 2.164.0, got "4.0.0"`))
			})
		})

		When("version is not parsable", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiVersionReturns("zenobia", nil)
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError("error parsing API version: Malformed version: zenobia"))
			})
		})

		When("error getting version", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiVersionReturns("", fmt.Errorf("boom"))
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError("error retrieving API version: boom"))
			})
		})
	})

	Describe("access token", func() {
		BeforeEach(func() {
			fakeCLIConnection.AccessTokenReturns("fake-token", nil)
		})

		It("gets the access token", func() {
			Expect(cfg.APIToken).To(Equal("fake-token"))
			Expect(cfgErr).NotTo(HaveOccurred())
		})

		When("error getting access token", func() {
			BeforeEach(func() {
				fakeCLIConnection.AccessTokenReturns("", fmt.Errorf("boom"))
			})

			It("returns the error", func() {
				Expect(cfgErr).To(MatchError("error reading access token: boom"))
				Expect(cfg.APIToken).To(Equal(""))
			})
		})
	})

	Describe("API endpoint", func() {
		BeforeEach(func() {
			fakeCLIConnection.ApiEndpointReturns("fake-api-endpoint", nil)
		})

		It("gets the API endpoint", func() {
			Expect(cfg.APIEndpoint).To(Equal("fake-api-endpoint"))
			Expect(cfgErr).NotTo(HaveOccurred())
		})

		When("error getting API endpoint", func() {
			BeforeEach(func() {
				fakeCLIConnection.ApiEndpointReturns("", fmt.Errorf("boom"))
			})

			It("returns the error", func() {
				Expect(cfgErr).To(MatchError("error reading API endpoint: boom"))
				Expect(cfg.APIToken).To(Equal(""))
			})
		})
	})

	Describe("skip SSL validation", func() {
		BeforeEach(func() {
			fakeCLIConnection.IsSSLDisabledReturns(true, nil)
		})

		It("gets the skip SSL validation", func() {
			Expect(cfg.SkipSSLValidation).To(Equal(true))
			Expect(cfgErr).NotTo(HaveOccurred())
		})

		When("error getting skip SSL validation", func() {
			BeforeEach(func() {
				fakeCLIConnection.IsSSLDisabledReturns(false, fmt.Errorf("boom"))
			})

			It("returns the error", func() {
				Expect(cfgErr).To(MatchError("error reading skip SSL validation: boom"))
				Expect(cfg.SkipSSLValidation).To(Equal(false))
			})
		})
	})

	Describe("verbose logging", func() {
		When("not specified", func() {
			It("defaults to false", func() {
				Expect(cfg.HTTPLogging).To(BeFalse())
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})

		When("set", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-loghttp")
			})

			It("is true", func() {
				Expect(cfgErr).NotTo(HaveOccurred())
				Expect(cfg.HTTPLogging).To(BeTrue())
			})
		})

		BeforeEach(func() {
			fakeCLIConnection.IsSSLDisabledReturns(true, nil)
		})

		It("gets the skip SSL validation", func() {
			Expect(cfg.SkipSSLValidation).To(Equal(true))
			Expect(cfgErr).NotTo(HaveOccurred())
		})

		When("error getting skip SSL validation", func() {
			BeforeEach(func() {
				fakeCLIConnection.IsSSLDisabledReturns(false, fmt.Errorf("boom"))
			})

			It("returns the error", func() {
				Expect(cfgErr).To(MatchError("error reading skip SSL validation: boom"))
				Expect(cfg.SkipSSLValidation).To(Equal(false))
			})
		})
	})

	Describe("parallel upgrades", func() {
		When("not specified", func() {
			It("defaults", func() {
				Expect(cfg.ParallelUpgrades).To(Equal(10))
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})

		When("specified", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-parallel", "42")
			})

			It("gets the value", func() {
				Expect(cfgErr).NotTo(HaveOccurred())
				Expect(cfg.ParallelUpgrades).To(Equal(42))
			})
		})

		When("not a number", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-parallel", "boudica")
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`invalid value "boudica" for flag -parallel: parse error`))
			})
		})

		When("too low", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-parallel", "0")
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`number of parallel upgrades must be in the range of 1 to 100`))
			})
		})

		When("too high", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-parallel", "101")
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`number of parallel upgrades must be in the range of 1 to 100`))
			})
		})
	})

	Describe("dry-run", func() {
		When("not specified", func() {
			It("is false", func() {
				Expect(cfg.DryRun).To(BeFalse())
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})

		When("specified", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-dry-run")
			})

			It("is true", func() {
				Expect(cfgErr).NotTo(HaveOccurred())
				Expect(cfg.DryRun).To(BeTrue())
			})
		})
	})

	Describe("min-version-required", func() {
		When("not specified", func() {
			It("is not set", func() {
				Expect(cfg.MinVersionRequired).To(BeEmpty())
			})
		})

		When("specified without version", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-min-version-required=")
			})

			It("an empty value is set", func() {
				Expect(cfg.MinVersionRequired).To(BeEmpty())
			})
		})

		When("specified with a non-semver version", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-min-version-required", "invalid version")
			})

			It("is set and the value is the version", func() {
				Expect(cfgErr).To(MatchError(ContainSubstring("error parsing check-up-to-date option: Malformed version: invalid version")))
			})
		})

		When("specified with version", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-min-version-required", "1.3.0")
			})

			It("is set and the value is the version", func() {
				Expect(cfg.MinVersionRequired).To(Equal("1.3.0"))
			})
		})
	})

	Describe("invalid combinations", func() {
		When("-dry-run and -parallel are specified together", func() {
			BeforeEach(func() {
				fakeArgs = append(fakeArgs, "-dry-run", "-parallel", "10")
			})

			It("succeeds", func() {
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})
	})

	Describe("broker name", func() {
		When("valid", func() {
			BeforeEach(func() {
				fakeArgs = []string{"lovely-broker-name"}
			})

			It("reads the name", func() {
				Expect(cfg.BrokerName).To(Equal("lovely-broker-name"))
				Expect(cfgErr).NotTo(HaveOccurred())
			})
		})

		When("not specified", func() {
			BeforeEach(func() {
				fakeArgs = nil
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`missing broker name`))
			})
		})

		When("invalid", func() {
			BeforeEach(func() {
				fakeArgs = []string{"fake**broker**name"}
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`broker name contains invalid characters`))
			})
		})

		When("more than one specified", func() {
			BeforeEach(func() {
				fakeArgs = []string{"fake-broker-name", "invalid-extra-parameter"}
			})

			It("returns an error", func() {
				Expect(cfgErr).To(MatchError(`too many parameters, did not parse: invalid-extra-parameter`))
			})
		})
	})
})
