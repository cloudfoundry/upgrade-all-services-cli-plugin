package config

// This defines the text for all the flags
const (
	parallelDefault     = 10
	parallelFlag        = "parallel"
	parallelDescription = "number of upgrades to run in parallel"

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
	checkUpToDateDescription = "checks and fails if any service instance is not up-to-date - implies a dry-run"

	checkDeactivatedPlansDefault     = false
	checkDeactivatedPlansFlag        = "check-deactivated-plans"
	checkDeactivatedPlansDescription = "checks whether any of the plans have been deactivated. If any deactivated plans are found, the command will fail"
)
