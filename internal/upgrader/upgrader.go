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
	UpgradeStarting(name string, guid string)
	UpgradeSucceeded(name string, guid string, duration time.Duration)
	UpgradeFailed(name string, guid string, duration time.Duration, err error)
	InitialTotals(totalServiceInstances, totalUpgradableServiceInstances int)
	FinalTotals()
}

func Upgrade(api CFClient, brokerName string, parallelUpgrades int, dryRun bool, log Logger) error {
	planVersions, err := discoverServicePlans(api, brokerName)
	if err != nil {
		return err
	}

	log.Printf("discovering service instances for broker: %s", brokerName)
	upgradableInstances, totalServiceInstances, err := discoverUpgradeableInstances(api, keys(planVersions))
	switch {
	case err != nil:
		return err
	case len(upgradableInstances) == 0:
		log.Printf("no instances available to upgrade")
		return nil
	case dryRun:
		return performDryRun(upgradableInstances, log)
	default:
		log.InitialTotals(totalServiceInstances, len(upgradableInstances))
		return performUpgrade(api, upgradableInstances, planVersions, parallelUpgrades, log)
	}
}

func performUpgrade(api CFClient, upgradableInstances []ccapi.ServiceInstance, planVersions map[string]string, parallelUpgrades int, log Logger) error {
	type upgradeTask struct {
		ServiceInstanceName    string
		ServiceInstanceGUID    string
		MaintenanceInfoVersion string
	}

	upgradeQueue := make(chan upgradeTask)
	go func() {
		for _, instance := range upgradableInstances {
			upgradeQueue <- upgradeTask{
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
			log.UpgradeStarting(instance.ServiceInstanceName, instance.ServiceInstanceGUID)
			err := api.UpgradeServiceInstance(instance.ServiceInstanceGUID, instance.MaintenanceInfoVersion)
			switch err {
			case nil:
				log.UpgradeSucceeded(instance.ServiceInstanceName, instance.ServiceInstanceGUID, time.Since(start))
			default:
				log.UpgradeFailed(instance.ServiceInstanceName, instance.ServiceInstanceGUID, time.Since(start), err)
			}
		}
	})

	log.FinalTotals()
	return nil
}

func performDryRun(upgradableInstances []ccapi.ServiceInstance, log Logger) error {
	log.Printf("the following service instances would be upgraded:")
	for _, i := range upgradableInstances {
		log.Printf(" - %s", i.GUID)
	}
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

func discoverUpgradeableInstances(api CFClient, planGUIDs []string) ([]ccapi.ServiceInstance, int, error) {
	serviceInstances, err := api.GetServiceInstances(planGUIDs)
	if err != nil {
		return nil, 0, err
	}

	var upgradableInstances []ccapi.ServiceInstance
	for _, i := range serviceInstances {
		if i.UpgradeAvailable {
			upgradableInstances = append(upgradableInstances, i)
		}
	}

	return upgradableInstances, len(serviceInstances), nil
}

func keys(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
