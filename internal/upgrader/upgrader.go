package upgrader

import (
	"fmt"
	"time"

	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/workers"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . CFClient
type CFClient interface {
	GetServiceInstances([]string) ([]ccapi.ServiceInstance, error)
	GetServicePlans(string) ([]ccapi.Plan, error)
	UpgradeServiceInstance(string, string) error
}

//counterfeiter:generate . Logger
type Logger interface {
	Printf(format string, a ...any)
	SkippingInstance(instance ccapi.ServiceInstance)
	UpgradeStarting(instance ccapi.ServiceInstance)
	UpgradeSucceeded(instance ccapi.ServiceInstance, duration time.Duration)
	UpgradeFailed(instance ccapi.ServiceInstance, duration time.Duration, err error)
	InitialTotals(totalServiceInstances, totalUpgradableServiceInstances int)
	FinalTotals()
}

func Upgrade(api CFClient, brokerName string, parallelUpgrades int, dryRun, checkUpToDate bool, log Logger) error {
	planVersions, err := discoverServicePlans(api, brokerName)
	if err != nil {
		return err
	}

	log.Printf("discovering service instances for broker: %s", brokerName)
	upgradableInstances, totalServiceInstances, err := discoverUpgradeableInstances(api, keys(planVersions), log)

	// See internal/ccapi/service_instances.go to understand why we are setting this value here
	for i := range upgradableInstances {
		upgradableInstances[i].PlanMaintenanceInfoVersion = planVersions[upgradableInstances[i].PlanGUID]
	}

	switch {
	case err != nil:
		return err
	case len(upgradableInstances) == 0:
		log.Printf("no instances available to upgrade")
		return nil
	case checkUpToDate:
		log.InitialTotals(totalServiceInstances, len(upgradableInstances))
		return performCheckUpToDate(upgradableInstances, log)
	case dryRun:
		log.InitialTotals(totalServiceInstances, len(upgradableInstances))
		return performDryRun(upgradableInstances, log)
	default:
		log.InitialTotals(totalServiceInstances, len(upgradableInstances))
		return performUpgrade(api, upgradableInstances, planVersions, parallelUpgrades, log)
	}
}

func performUpgrade(api CFClient, upgradableInstances []ccapi.ServiceInstance, planVersions map[string]string, parallelUpgrades int, log Logger) error {
	type upgradeTask struct {
		UpgradeableIndex       int
		ServiceInstanceName    string
		ServiceInstanceGUID    string
		MaintenanceInfoVersion string
	}

	upgradeQueue := make(chan upgradeTask)
	go func() {
		for i, instance := range upgradableInstances {
			upgradeQueue <- upgradeTask{
				UpgradeableIndex:       i,
				ServiceInstanceName:    instance.Name,
				ServiceInstanceGUID:    instance.GUID,
				MaintenanceInfoVersion: planVersions[instance.PlanGUID],
			}
		}
		close(upgradeQueue)
	}()

	workers.Run(parallelUpgrades, func() {
		for instance := range upgradeQueue {
			start := time.Now()
			log.UpgradeStarting(upgradableInstances[instance.UpgradeableIndex])
			err := api.UpgradeServiceInstance(instance.ServiceInstanceGUID, instance.MaintenanceInfoVersion)
			switch err {
			case nil:
				log.UpgradeSucceeded(upgradableInstances[instance.UpgradeableIndex], time.Since(start))
			default:
				log.UpgradeFailed(upgradableInstances[instance.UpgradeableIndex], time.Since(start), err)
			}
		}
	})

	log.FinalTotals()
	return nil
}

func performCheckUpToDate(upgradableInstances []ccapi.ServiceInstance, log Logger) error {
	err := performDryRun(upgradableInstances, log)
	if err != nil {
		return fmt.Errorf("check up-to-date failed because dry-run returned the following error: %w", err)
	}
	if len(upgradableInstances) > 0 {
		return fmt.Errorf("check up-to-date failed: found %d instances which are not up-to-date", len(upgradableInstances))
	}
	return nil
}

func performDryRun(upgradableInstances []ccapi.ServiceInstance, log Logger) error {
	log.Printf("the following service instances would be upgraded:")
	for _, i := range upgradableInstances {
		log.UpgradeFailed(i, time.Duration(0), fmt.Errorf("dry-run prevented upgrade"))
	}
	log.FinalTotals()
	return nil
}

func discoverServicePlans(api CFClient, brokerName string) (map[string]string, error) {
	plans, err := api.GetServicePlans(brokerName)
	if err != nil {
		return nil, err
	}

	if len(plans) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("no service plans available for broker: %s", brokerName))
	}

	planVersions := make(map[string]string)

	for _, plan := range plans {
		planVersions[plan.GUID] = plan.MaintenanceInfoVersion
	}

	return planVersions, nil
}

func discoverUpgradeableInstances(api CFClient, planGUIDs []string, log Logger) ([]ccapi.ServiceInstance, int, error) {
	serviceInstances, err := api.GetServiceInstances(planGUIDs)
	if err != nil {
		return nil, 0, err
	}

	var upgradableInstances []ccapi.ServiceInstance
	for _, i := range serviceInstances {
		if i.UpgradeAvailable && isCreateFailed(i.LastOperation.Type, i.LastOperation.State) {
			log.SkippingInstance(i)
		} else if i.UpgradeAvailable {
			upgradableInstances = append(upgradableInstances, i)
		}
	}

	return upgradableInstances, len(serviceInstances), nil
}

func isCreateFailed(operationType, operationState string) bool {
	return operationType == "create" && operationState == "failed"
}

func keys(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
