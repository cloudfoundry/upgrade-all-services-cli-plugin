package upgrader

import (
	"encoding/json"
	"fmt"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/slicex"
	"upgrade-all-services-cli-plugin/internal/versionchecker"

	"github.com/hashicorp/go-version"
)

// performMinimumVersionRequiredCheck lists service instances whose version is lower than the specified version
func performMinimumVersionRequiredCheck(api CFClient, cfg UpgradeConfig) error {
	serviceInstances, err := getAllServiceInstances(api, cfg.BrokerName)
	if err != nil {
		return err
	}

	filteredInstances, err := filterInstancesVersionLessThanMinimumVersionRequired(serviceInstances, cfg.MinVersion)
	if err != nil {
		return err
	}

	switch cfg.JSONOutput {
	case true:
		return outputMinimumVersionJSON(filteredInstances)
	default:
		return outputMinimumVersionText(filteredInstances, len(serviceInstances), cfg.BrokerName, cfg.MinVersion.String())
	}
}

func outputMinimumVersionText(filteredInstances []ccapi.ServiceInstance, totalServiceInstances int, brokerName, minVersion string) error {
	fmt.Printf("Discovering service instances for broker: %s\n", brokerName)
	fmt.Printf("Total number of service instances: %d\n", totalServiceInstances)
	if len(filteredInstances) == 0 {
		fmt.Printf("No instances found with version lower than %q\n", minVersion)
		return nil
	}

	fmt.Printf("Number of service instances with a version lower than %q: %d\n", minVersion, len(filteredInstances))
	fmt.Println()
	logServiceInstances(filteredInstances)
	return fmt.Errorf("found %d service instances with a version less than the minimum required", len(filteredInstances))
}

func outputMinimumVersionJSON(filteredInstances []ccapi.ServiceInstance) error {
	lines := slicex.Map(filteredInstances, newJSONOutputServiceInstance)

	output, err := json.MarshalIndent(lines, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))

	// In contrast to text output, we don't exit with an error code. The rationale is that JSON may be piped
	// to a processor command like `jq`, and a command failure would be more of a hindrance than a help.
	return nil
}

func filterInstancesVersionLessThanMinimumVersionRequired(instances []ccapi.ServiceInstance, minVersion *version.Version) ([]ccapi.ServiceInstance, error) {
	checker, err := versionchecker.New(minVersion)
	if err != nil {
		return nil, err
	}

	var filteredInstances []ccapi.ServiceInstance
	for _, instance := range instances {
		is, err := checker.IsInstanceVersionLessThanMinimumRequired(instance.MaintenanceInfoVersion)
		if err != nil {
			return nil, err
		}

		if is {
			filteredInstances = append(filteredInstances, instance)
		}
	}
	return filteredInstances, nil
}
