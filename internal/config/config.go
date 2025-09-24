package config

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
)

// Config is the type that contains all the configuration data required to run the plugin
// Note that while in theory it could make sense to make some of the `int` types a `uint`,
// in practice this is messy because the error from flags.UintVar() look like an internal
// error rather than a user error, and you have to cast the `uint` to compare to a `len()`,
// which just looks more complicated than it has to. So we use an `int` for pragmatic reasons.
type Config struct {
	Action            Action
	BrokerName        string
	APIToken          string
	APIEndpoint       string
	SkipSSLValidation bool
	HTTPLogging       bool
	JSONOutput        bool
	MinVersion        *version.Version
	ParallelUpgrades  int
	Limit             int
	Attempts          int
	RetryInterval     time.Duration
}

// ParseConfig combines and validates data from the command line and CLIConnection object
func ParseConfig(conn CLIConnection, args []string) (Config, error) {
	var (
		cfg                   Config
		dryRun                bool
		checkUpToDate         bool
		minVersionRequired    string
		checkDeactivatedPlans bool
	)

	flagSet := flag.NewFlagSet("upgrade-all-services", flag.ContinueOnError)
	flagSet.IntVar(&cfg.ParallelUpgrades, parallelFlag, parallelDefault, parallelDescription)
	flagSet.BoolVar(&cfg.HTTPLogging, httpLoggingFlag, httpLoggingDefault, httpLoggingDescription)
	flagSet.BoolVar(&cfg.JSONOutput, jsonOutputFlag, jsonOutputDefault, jsonOutputDescription)
	flagSet.BoolVar(&dryRun, dryRunFlag, dryRunDefault, dryRunDescription)
	flagSet.BoolVar(&checkUpToDate, checkUpToDateFlag, checkUpToDateDefault, checkUpToDateDescription)
	flagSet.StringVar(&minVersionRequired, minVersionRequiredFlag, minVersionRequiredDefault, minVersionRequiredDescription)
	flagSet.BoolVar(&checkDeactivatedPlans, checkDeactivatedPlansFlag, checkDeactivatedPlansDefault, checkDeactivatedPlansDescription)
	flagSet.IntVar(&cfg.Limit, limitFlag, limitDefault, limitDescription)
	flagSet.IntVar(&cfg.Attempts, attemptsFlag, attemptsDefault, attemptsDescription)
	flagSet.DurationVar(&cfg.RetryInterval, retryIntervalFlag, retryIntervalDefault, retryIntervalDescription)

	// This ranges over a chain of functions, each of which performs a single action and may return an error.
	// The chain breaks at the first error received. It arguably reads better than repetitive error handling logic.
	for _, s := range []func() error{
		func() (err error) {
			cfg.BrokerName, err = parseCommandLine(flagSet, args)
			return
		},
		func() (err error) {
			cfg.Action, err = determineAction(checkDeactivatedPlans, checkUpToDate, dryRun, minVersionRequired)
			return
		},
		func() error { return validateLoginStatus(conn) },
		func() error { return validateAPIVersion(conn) },
		func() error { return read("access token", conn.AccessToken, &cfg.APIToken) },
		func() error { return read("API endpoint", conn.ApiEndpoint, &cfg.APIEndpoint) },
		func() error { return read("skip SSL validation", conn.IsSSLDisabled, &cfg.SkipSSLValidation) },
		func() error { return validateParallelUpgrades(cfg.ParallelUpgrades) },
		func() error { return validateBrokerName(cfg.BrokerName) },
		func() (err error) {
			cfg.MinVersion, err = validateMinVersionRequired(minVersionRequired)
			return
		},
		func() error { return validateJSONFlag(cfg.JSONOutput, cfg.Action) },
		func() error { return validateLimit(cfg.Limit) },
		func() error { return validateAttempts(cfg.Attempts) },
		func() error { return validateRetryInterval(cfg.RetryInterval) },
	} {
		if err := s(); err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

// parseCommandLine reads the command line argument, with validation handled later by validation functions
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

// read calls a function (typically on the object that implements CLIConnection) and assuming
// no error it stores it in the specified location. This arguably reads better than repetitive logic.
func read[T any](desc string, get func() (T, error), set *T) error {
	data, err := get()
	if err != nil {
		return fmt.Errorf("error reading %s: %w", desc, err)
	}

	*set = data
	return nil
}
