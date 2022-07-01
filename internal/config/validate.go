package config

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
)

func ParseConfig(conn CLIConnection, args []string) (Config, error) {
	var (
		o opts
		c Config
	)

	for _, s := range []func() error{
		func() (err error) {
			c.BrokerName, c.ParallelUpgrades, err = Parse(args)
			return
		},
		func() error {
			return validateBrokerName(c.BrokerName)
		},
		func() (err error) {
			return validateParallelUpgrades(c.ParallelUpgrades)
		},
		func() error {
			return validateAPIVersion(conn)
		},
		func() error {
			return validateLoggedIn(conn)
		},
		func() (err error) {
			c.APIEndpoint, err = readAPIEndpoint(conn)
			return
		},
		func() (err error) {
			c.SkipSSLValidation, err = readSkipSSLValidation(conn)
			return
		},
	} {
		if err := s(); err != nil {
			return Config{}, err
		}
	}

	return c, nil
}

type Config struct {
	BrokerName        string
	AccessToken       string
	APIEndpoint       string
	SkipSSLValidation bool
	ParallelUpgrades  int
	Verbose           bool
}

type CLIConnection interface {
	AccessToken() (string, error)
	ApiEndpoint() (string, error)
	ApiVersion() (string, error)
	IsSSLDisabled() (bool, error)
	IsLoggedIn() (bool, error)
}

func validateAPIVersion(conn CLIConnection) error {
	rawAPIVersion, err := conn.ApiVersion()
	if err != nil {
		return fmt.Errorf("error retrieving api version: %s", err)
	}

	apiVersion, err := version.NewVersion(rawAPIVersion)
	switch {
	case err != nil:
		return err
	case apiVersion.LessThan(version.Must(version.NewVersion("3.99.0"))):
		return fmt.Errorf("plugin requires CF API version >= 3.99.0")
	default:
		return nil
	}
}

func validateLoggedIn(conn CLIConnection) error {
	isLoggedIn, err := conn.IsLoggedIn()
	switch {
	case err != nil:
		return fmt.Errorf("error validating user authentication: %s", err)
	case !isLoggedIn:
		return fmt.Errorf("you must authenticate with the cf cli before running this command")
	default:
		return nil
	}
}

func validateBrokerName(name string) error {
	switch regexp.MustCompile(`^[\w_-]+$`).MatchString(name) {
	case true:
		return nil
	default:
		return fmt.Errorf("invalid broker name format")
	}
}

func validateParallelUpgrades(p int) error {
	if p <= 0 || p > 100 {
		return fmt.Errorf("number of parallel upgrades must be in the range of 1 to 100")
	}
	return nil
}

func readAPIEndpoint(conn CLIConnection) (string, error) {
	apiEndPoint, err := conn.ApiEndpoint()
	if err != nil {
		return "", fmt.Errorf("error retrieving api endpoint: %s", err)
	}
	return apiEndPoint, nil
}

func readSkipSSLValidation(conn CLIConnection) (bool, error) {
	skipSSLValidation, err := conn.IsSSLDisabled()
	if err != nil {
		return false, fmt.Errorf("error retrieving api ssl validation status: %s", err)
	}
	return skipSSLValidation, nil
}
