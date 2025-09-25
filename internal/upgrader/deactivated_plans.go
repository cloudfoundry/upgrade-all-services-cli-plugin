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
		if err := outputDeactivatedPlansJSON(instancesWithDeactivatedPlans); err != nil {
			return err
		}
	default:
		outputDeactivatedPlansText(instancesWithDeactivatedPlans, cfg.BrokerName, len(instances))
	}

	if len(instancesWithDeactivatedPlans) > 0 {
		return newInstanceError("discovered deactivated plans associated with instances")
	}

	return nil
}

func outputDeactivatedPlansText(instancesWithDeactivatedPlans []ccapi.ServiceInstance, brokerName string, totalServiceInstances int) {
	fmt.Printf("Discovering service instances for broker: %s\n", brokerName)
	fmt.Printf("Total number of service instances: %d\n", totalServiceInstances)
	if len(instancesWithDeactivatedPlans) == 0 {
		fmt.Println("No instances found associated with deactivated plans")
		return
	}

	fmt.Printf("Number of service instances associated with deactivated plans: %d\n", len(instancesWithDeactivatedPlans))
	fmt.Println()
	logServiceInstances(instancesWithDeactivatedPlans)
}

func outputDeactivatedPlansJSON(instancesWithDeactivatedPlans []ccapi.ServiceInstance) error {
	output, err := json.MarshalIndent(slicex.Map(instancesWithDeactivatedPlans, newJSONOutputServiceInstance), "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}
