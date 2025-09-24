package upgrader

import (
	"encoding/json"
	"fmt"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/slicex"
)

// performDeactivatedPlansCheck lists service instances associated with deactivated plans
func performDeactivatedPlansCheck(api CFClient, cfg UpgradeConfig) error {
	instances, err := getAllServiceInstances(api, cfg.BrokerName)
	if err != nil {
		return err
	}

	instancesWithDeactivatedPlans := slicex.Filter(instances, func(instance ccapi.ServiceInstance) bool { return instance.ServicePlanDeactivated })

	switch {
	case cfg.JSONOutput:
		return outputDeactivatedPlansJSON(instancesWithDeactivatedPlans)
	default:
		return outputDeactivatedPlansText(instancesWithDeactivatedPlans, cfg.BrokerName, len(instances))
	}
}

func outputDeactivatedPlansText(instancesWithDeactivatedPlans []ccapi.ServiceInstance, brokerName string, totalServiceInstances int) error {
	fmt.Printf("Discovering service instances for broker: %s\n", brokerName)
	fmt.Printf("Total number of service instances: %d\n", totalServiceInstances)
	if len(instancesWithDeactivatedPlans) == 0 {
		fmt.Println("No instances found associated with deactivated plans")
		return nil
	}

	fmt.Printf("Number of service instances associated with deactivated plans: %d\n", len(instancesWithDeactivatedPlans))
	fmt.Println()
	logServiceInstances(instancesWithDeactivatedPlans)
	return newInstanceError("discovered deactivated plans associated with instances")
}

func outputDeactivatedPlansJSON(instancesWithDeactivatedPlans []ccapi.ServiceInstance) error {
	output, err := json.MarshalIndent(slicex.Map(instancesWithDeactivatedPlans, newJSONOutputServiceInstance), "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))

	if len(instancesWithDeactivatedPlans) > 0 {
		return newInstanceError("discovered deactivated plans associated with instances")
	}

	return nil
}
