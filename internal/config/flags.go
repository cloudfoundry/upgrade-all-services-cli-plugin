package config

import "time"

// This defines the text for all the flags
const (
	parallelDefault     = 10
	parallelFlag        = "parallel"
	parallelDescription = "number of upgrades to run in parallel"
	parallelMaximum     = 100

	// Ideally we would have used "-v" as the flag as the CF CLI does,
	// but unfortunately the CF CLI swallows this flag, and the value
	// is not available to plugins
	httpLoggingDefault     = false
	httpLoggingFlag        = "loghttp"
	httpLoggingDescription = "enable HTTP request logging"

	dryRunDefault     = false
	dryRunFlag        = "dry-run"
	dryRunDescription = "print the service instances that would be upgraded"

	checkUpToDateDefault     = false
	checkUpToDateFlag        = "check-up-to-date"
	checkUpToDateDescription = "checks and fails if any service instance is not up-to-date. An instance is not up-to-date if it is marked as upgradable or belongs to a deactivated plan"

	minVersionRequiredDefault     = ""
	minVersionRequiredFlag        = "min-version-required"
	minVersionRequiredDescription = "--min-version-required <major.minor.patch>. Checks and fails if any service instance has a version lower than the specified"

	checkDeactivatedPlansDefault     = false
	checkDeactivatedPlansFlag        = "check-deactivated-plans"
	checkDeactivatedPlansDescription = "checks whether any of the plans have been deactivated. If any deactivated plans are found, the command will fail"

	jsonOutputDefault     = false
	jsonOutputFlag        = "json"
	jsonOutputDescription = "output as JSON"

	limitDefault     = 0
	limitFlag        = "limit"
	limitDescription = "stop after attempting to upgrade the specified number of service instances. 0 means no limit"

	attemptsDefault     = 1
	attemptsFlag        = "attempts"
	attemptsDescription = "maximum number of attempts to perform an upgrade operation, between 1 and 10. Default is 1."
	attemptsMaximum     = 10 // If someone is having to retry more than 10 times, then they should probably be investigating other avenues

	retryIntervalDefault     = 0
	retryIntervalFlag        = "retry-interval"
	retryIntervalDescription = "time to wait after a failure before a retry, e.g. '10s', '2m'. Maximum 10m, default 0."
	retryIntervalMaximum     = 10 * time.Minute

	ignoreInstanceErrorsDefault     = false
	ignoreInstanceErrorsFlag        = "ignore-instance-errors"
	ignoreInstanceErrorsDescription = "exit with code 0 even when the -min-version-required, -check-deactivated-plans, or -check-up-to-date detect outdated service instances"

	instancePollingIntervalDefault     = 10 * time.Second
	instancePollingIntervalFlag        = "instance-polling-interval"
	instancePollingIntervalDescription = "polling interval for service instances during the upgrade process. Default is 10s"
	instancePollingIntervalMinimum     = time.Millisecond
	instancePollingIntervalMaximum     = time.Minute

	instanceTimeoutDefault     = 10 * time.Minute
	instanceTimeoutFlag        = "instance-timeout"
	instanceTimeoutDescription = "timeout for service instance upgrade operations. Default is 10m"
	instanceTimeoutMinimum     = time.Second
	instanceTimeoutMaximum     = 24 * time.Hour
)
