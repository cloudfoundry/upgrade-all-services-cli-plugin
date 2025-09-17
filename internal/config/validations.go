package config

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
)

// validateLoginStatus checks that the CF CLI is logged in
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

// validateAPIVersion checks that the CAPI API version is recent enough
func validateAPIVersion(conn CLIConnection) error {
	ver, err := conn.ApiVersion()
	if err != nil {
		return fmt.Errorf("error retrieving API version: %w", err)
	}

	var (
		v3    = version.Must(version.NewVersion("3"))
		v4    = version.Must(version.NewVersion("4"))
		v2min = version.Must(version.NewVersion("2.164"))
		v3min = version.Must(version.NewVersion("3.99"))
	)

	v, err := version.NewVersion(ver)
	switch {
	case err != nil:
		return fmt.Errorf("error parsing API version: %w", err)
	case v.GreaterThanOrEqual(v3min) && v.LessThan(v4):
		return nil
	case v.GreaterThanOrEqual(v2min) && v.LessThan(v3):
		// There's a bug in CF CLI v6 where the API version is sometimes reported as v3 and sometimes as v2,
		// depending on whether "cf login" or "cf api" was used. CAPI release 1.109.0 shipped with both
		// API v3.99 and CF API v2.164, so if we have at least v2.164 then we know that v3.99 is also available
		return nil
	default:
		return fmt.Errorf("plugin requires minimum API version %s or %s, got %q", v3min.String(), v2min.String(), v.String())
	}
}

// validateParallelUpgrades checks that the parallelisation parameter is within bounds
func validateParallelUpgrades(p int) error {
	if p <= 0 || p > 100 {
		printUsage()
		return fmt.Errorf("number of parallel upgrades must be in the range of 1 to 100")
	}
	return nil
}

// validateBrokerName checks that the specified broker name is viable
func validateBrokerName(name string) error {
	if valid := regexp.MustCompile(`^[\w_.-]+$`).MatchString(name); !valid {
		printUsage()
		return fmt.Errorf("broker name contains invalid characters")
	}

	return nil
}

func validateMinVersionRequired(ver string) (*version.Version, error) {
	if ver == "" {
		return nil, nil
	}

	v, err := version.NewVersion(ver)
	if err != nil {
		return nil, fmt.Errorf("error parsing min-version-required option: %w", err)
	}
	return v, nil
}

func validateJSONFlag(value bool, action Action) error {
	if !value {
		return nil
	}

	if action != MinVersionCheckAction {
		return fmt.Errorf("the --%s flag can only be used with the --%s flag", jsonOutputFlag, minVersionRequiredFlag)
	}

	return nil
}
