package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/hashicorp/go-version"
)

const (
	Usage               = "cf upgrade-all-services <broker-name>"
	parallelDefault     = 10
	parallelFlag        = "parallel"
	parallelDescription = "number of upgrades to run in parallel (defaults to 10)"
)

func ParseConfig(conn CLIConnection, args []string) (Config, error) {
	var cfg Config

	flagSet := flag.NewFlagSet("upgrade-all-services", flag.ContinueOnError)
	flagSet.IntVar(&cfg.ParallelUpgrades, "parallel", parallelDefault, parallelDescription)

	for _, s := range []func() error{
		func() error {
			return validateLoginStatus(conn)
		},
		func() error {
			return validateAPIVersion(conn)
		},
		func() error {
			return read("access token", conn.AccessToken, &cfg.APIToken)
		},
		func() error {
			return read("API endpoint", conn.ApiEndpoint, &cfg.APIEndpoint)
		},
		func() error {
			return read("skip SSL validation", conn.IsSSLDisabled, &cfg.SkipSSLValidation)
		},
		func() error {
			return flagSet.Parse(args)
		},
		func() error {
			return validateParallelUpgrades(cfg.ParallelUpgrades)
		},
		func() (err error) {
			cfg.BrokerName, err = readBrokerName(flagSet.Args())
			return
		},
	} {
		if err := s(); err != nil {
			return Config{}, err
		}
	}
	return cfg, nil
}

type Config struct {
	BrokerName        string
	APIToken          string
	APIEndpoint       string
	SkipSSLValidation bool
	ParallelUpgrades  int
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . CLIConnection
type CLIConnection interface {
	IsLoggedIn() (bool, error)
	AccessToken() (string, error)
	ApiVersion() (string, error)
	ApiEndpoint() (string, error)
	IsSSLDisabled() (bool, error)
}

func Options() map[string]string {
	return map[string]string{
		fmt.Sprintf("-%s", parallelFlag): parallelDescription,
	}
}

func read[T any](desc string, get func() (T, error), set *T) error {
	data, err := get()
	if err != nil {
		return fmt.Errorf("error reading %s: %w", desc, err)
	}

	*set = data
	return nil
}

func validateLoginStatus(conn CLIConnection) error {
	loggedIn, err := conn.IsLoggedIn()
	switch {
	case err != nil:
		return fmt.Errorf("error getting login status: %w", err)
	case !loggedIn:
		return fmt.Errorf("you must authenticate with the cf cli before running this command")
	default:
		return nil
	}
}

func validateAPIVersion(conn CLIConnection) error {
	ver, err := conn.ApiVersion()
	if err != nil {
		return fmt.Errorf("error retrieving API version: %w", err)
	}

	v, err := version.NewVersion(ver)
	switch {
	case err != nil:
		return fmt.Errorf("error parsing API version: %w", err)
	case v.GreaterThanOrEqual(version.Must(version.NewVersion("4"))):
		return fmt.Errorf("plugin requires API major version v3, got: %q", v.String())
	case v.LessThan(version.Must(version.NewVersion("3.99"))):
		return fmt.Errorf("plugin requires minimum API version v3.99, got: %q", v.String())
	default:
		return nil
	}
}

func validateParallelUpgrades(p int) error {
	if p <= 0 || p > 100 {
		printUsage()
		return fmt.Errorf("number of parallel upgrades must be in the range of 1 to 100")
	}
	return nil
}

func readBrokerName(args []string) (string, error) {
	switch len(args) {
	case 0:
		printUsage()
		return "", fmt.Errorf("missing broker name")
	case 1: // OK
	default:
		printUsage()
		return "", fmt.Errorf("too many parameters")
	}

	if valid := regexp.MustCompile(`^[\w_.-]+$`).MatchString(args[0]); !valid {
		printUsage()
		return "", fmt.Errorf("broker name contains invalid characters")
	}

	return args[0], nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n", Usage)
}
