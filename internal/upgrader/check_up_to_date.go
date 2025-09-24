package upgrader

import (
	"encoding/json"
	"fmt"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/slicex"
)

// performUpToDateCheck performs multiple checks:
// - it lists service instances associated with deactivated plans (the same as performDeactivatedPlansCheck)
// - it lists service instances that have an upgrade available and failed to create
// - it lists service instances that have an upgrade available and did not fail to create (similar to performing a dry run)
func performUpToDateCheck(api CFClient, cfg UpgradeConfig) error {
	instances, err := getGroupedServiceInstances(api, cfg.BrokerName, 0)
	if err != nil {
		return err
	}

	switch cfg.JSONOutput {
	case true:
		return outputUpToDateJSON(instances.deactivatedPlan, instances.upgradeable, instances.createFailed)
	default:
		return outputUpToDateText(instances.deactivatedPlan, instances.upgradeable, instances.createFailed, len(instances.all), cfg.BrokerName)
	}
}

func outputUpToDateText(instancesWithDeactivatedPlans, upgradableInstances, createFailedInstances []ccapi.ServiceInstance, totalServiceInstances int, brokerName string) error {
	fmt.Printf("Discovering service instances for broker: %s\n", brokerName)
	fmt.Printf("Total number of service instances: %d\n", totalServiceInstances)

	fmt.Printf("Number of service instances associated with deactivated plans: %d\n", len(instancesWithDeactivatedPlans))
	fmt.Println()
	logServiceInstances(instancesWithDeactivatedPlans)

	fmt.Printf("Number of service instances with an upgrade available: %d\n", len(upgradableInstances))
	fmt.Println()
	logServiceInstances(upgradableInstances)

	fmt.Printf("Number of service instances which failed to create: %d\n", len(createFailedInstances))
	fmt.Println()
	logServiceInstances(createFailedInstances)

	if len(instancesWithDeactivatedPlans) > 0 || len(upgradableInstances) > 0 {
		return newInstanceError("discovered service instances associated with deactivated plans or with an upgrade available")
	}

	fmt.Println("No instances found associated with deactivated plans or with an upgrade available")
	return nil
}

func outputUpToDateJSON(instancesWithDeactivatedPlans, upgradableInstances, createFailedInstances []ccapi.ServiceInstance) error {
	type formatter struct {
		DeactivatedPlans []jsonOutputServiceInstance `json:"plan_deactivated"`
		UpgradePending   []jsonOutputServiceInstance `json:"upgrade_pending"`
		CreateFailed     []jsonOutputServiceInstance `json:"create_failed"`
	}

	data := formatter{
		DeactivatedPlans: slicex.Map(instancesWithDeactivatedPlans, newJSONOutputServiceInstance),
		UpgradePending:   slicex.Map(upgradableInstances, newJSONOutputServiceInstance),
		CreateFailed:     slicex.Map(createFailedInstances, newJSONOutputServiceInstance),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))

	if len(instancesWithDeactivatedPlans) > 0 || len(upgradableInstances) > 0 {
		return newInstanceError("discovered service instances associated with deactivated plans or with an upgrade available")
	}

	return nil
}
