package upgrader

import (
	"errors"
	"fmt"
	"upgrade-all-services-cli-plugin/internal/ccapi"
)

type checkDeactivatedPlansLogger interface {
	DeactivatedPlan(ccapi.ServiceInstance)
	Printf(string, ...any)
}

func checkDeactivatedPlans(api CFClient, log checkDeactivatedPlansLogger, brokerName string) error {
	servicePlans, err := api.GetServicePlans(brokerName)
	if err != nil {
		return err
	}

	if len(servicePlans) == 0 {
		return fmt.Errorf("no service plans available for broker: %s", brokerName)
	}

	log.Printf("discovering service instances for broker: %s", brokerName)
	instances, err := api.GetServiceInstancesForServicePlans(servicePlans)
	if err != nil {
		return err
	}

	var deactivatedPlanFound bool
	for _, instance := range instances {
		if instance.ServicePlanDeactivated {
			deactivatedPlanFound = true
			log.DeactivatedPlan(instance)
		}
	}

	if deactivatedPlanFound {
		return errors.New(
			"discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services",
		)
	}
	return nil
}
