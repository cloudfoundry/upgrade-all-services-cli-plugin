package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

const (
	Usage = "cf upgrade-all-services <broker-name> [options]"

	parallelDefault     = 10
	parallelFlag        = "parallel"
	parallelDescription = "number of upgrades to run in parallel"

	// Ideally we would have used "-v" as the flag as the CF CLI does,
	// but unfortunately the CF CLI swallows this flag, and the value
	// is not available to plugins
	httpLoggingDefault     = false
	httpLoggingFlag        = "loghttp"
	httpLoggingDescription = "enable HTTP request logging"
)

func ParseConfig(conn CLIConnection, args []string) (Config, error) {
	var cfg Config

	flagSet := flag.NewFlagSet("upgrade-all-services", flag.ContinueOnError)
	flagSet.IntVar(&cfg.ParallelUpgrades, parallelFlag, parallelDefault, parallelDescription)
	flagSet.BoolVar(&cfg.HTTPLogging, httpLoggingFlag, httpLoggingDefault, httpLoggingDescription)

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
		func() (err error) {
			cfg.BrokerName, err = parseCommandLine(flagSet, args)
			return
		},
		func() error {
			return validateParallelUpgrades(cfg.ParallelUpgrades)
		},
		func() error {
			return validateBrokerName(cfg.BrokerName)
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
	HTTPLogging       bool
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
		fmt.Sprintf("-%s", parallelFlag):    parallelDescription,
		fmt.Sprintf("-%s", httpLoggingFlag): httpLoggingDescription,
	}
}

func parseCommandLine(flagSet *flag.FlagSet, args []string) (string, error) {
	if len(args) == 0 {
		printUsage()
		return "", fmt.Errorf("missing broker name")
	}

	if err := flagSet.Parse(args[1:]); err != nil {
		return "", err
	}

	if len(flagSet.Args()) > 0 {
		printUsage()
		return "", fmt.Errorf("too many parameters, did not parse: %s", strings.Join(flagSet.Args(), " "))
	}

	return args[0], nil
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

func validateBrokerName(name string) error {
	if valid := regexp.MustCompile(`^[\w_.-]+$`).MatchString(name); !valid {
		printUsage()
		return fmt.Errorf("broker name contains invalid characters")
	}

	return nil
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n", Usage)
}
