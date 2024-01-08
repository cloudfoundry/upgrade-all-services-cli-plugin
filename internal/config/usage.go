package config

import (
	"fmt"
	"os"
)

// Usage is the short usage string
const Usage = "cf upgrade-all-services <broker-name> [options]"

// UsageOptions documents the available options. It is called by the CF CLI plugin infrastructure.
func UsageOptions() map[string]string {
	return map[string]string{
		parallelFlag:              parallelDescription,
		httpLoggingFlag:           httpLoggingDescription,
		dryRunFlag:                dryRunDescription,
		minVersionRequiredFlag:    minVersionRequiredDescription,
		checkUpToDateFlag:         checkUpToDateDescription,
		checkDeactivatedPlansFlag: checkDeactivatedPlansDescription,
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n", Usage)
}
