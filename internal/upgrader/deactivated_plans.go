package upgrader

import (
	"encoding/json"
	"errors"
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
	return errors.New("discovered deactivated plans associated with instances")
}

func outputDeactivatedPlansJSON(instances []ccapi.ServiceInstance) error {
	output, err := json.MarshalIndent(slicex.Map(instances, newJSONOutputServiceInstance), "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))

	// In contrast to text output, we don't exit with an error code. The rationale is that JSON may be piped
	// to a processor command like `jq`, and a command failure would be more of a hindrance than a help.
	return nil
}
