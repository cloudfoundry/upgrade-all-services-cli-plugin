package upgrader

import (
	"errors"
	"fmt"
	"time"

	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/workers"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . CFClient
type CFClient interface {
	GetServiceInstancesForServicePlans([]ccapi.ServicePlan) ([]ccapi.ServiceInstance, error)
	GetServicePlans(string) ([]ccapi.ServicePlan, error)
	UpgradeServiceInstance(string, string) error
}

//counterfeiter:generate . Logger
type Logger interface {
	Printf(format string, a ...any)
	SkippingInstance(instance ccapi.ServiceInstance)
	UpgradeStarting(instance ccapi.ServiceInstance)
	UpgradeSucceeded(instance ccapi.ServiceInstance, duration time.Duration)
	UpgradeFailed(instance ccapi.ServiceInstance, duration time.Duration, err error)
	DeactivatedPlan(instance ccapi.ServiceInstance)
	InitialTotals(totalServiceInstances, totalUpgradableServiceInstances int)
	FinalTotals()
}

type UpgradeConfig struct {
	BrokerName            string
	ParallelUpgrades      int
	DryRun                bool
	CheckUpToDate         bool
	CheckDeactivatedPlans bool
}

func Upgrade(api CFClient, log Logger, cfg UpgradeConfig) error {
	servicePlans, err := api.GetServicePlans(cfg.BrokerName)
	if err != nil {
		return err
	}

	if len(servicePlans) == 0 {
		return fmt.Errorf(fmt.Sprintf("no service plans available for broker: %s", cfg.BrokerName))
	}

	log.Printf("discovering service instances for broker: %s", cfg.BrokerName)
	upgradableInstances, totalServiceInstances, err := discoverUpgradeableInstances(api, servicePlans, log)
	if err != nil {
		return err
	}

	if cfg.CheckDeactivatedPlans {
		if err := checkDeactivatedPlans(log, upgradableInstances); err != nil {
			return err
		}
	}

	switch {
	case len(upgradableInstances) == 0:
		log.Printf("no instances available to upgrade")
		return nil
	case cfg.CheckUpToDate:
		log.InitialTotals(totalServiceInstances, len(upgradableInstances))
		return performCheckUpToDate(upgradableInstances, log)
	case cfg.DryRun:
		log.InitialTotals(totalServiceInstances, len(upgradableInstances))
		return performDryRun(upgradableInstances, log)
	default:
		log.InitialTotals(totalServiceInstances, len(upgradableInstances))
		return performUpgrade(api, upgradableInstances, cfg.ParallelUpgrades, log)
	}
}

func checkDeactivatedPlans(log Logger, upgradableInstances []ccapi.ServiceInstance) error {
	var deactivatedPlanFound bool
	for _, instance := range upgradableInstances {
		if instance.ServicePlanDeactivated {
			deactivatedPlanFound = true
			log.DeactivatedPlan(instance)
		}
	}

	if deactivatedPlanFound {
		return errors.New(
			"discovered deactivated plans associated with upgradable instances. Review the log to collect information and restore the deactivated plans or create user provided services",
		)
	}
	return nil
}

func performUpgrade(api CFClient, upgradableInstances []ccapi.ServiceInstance, parallelUpgrades int, log Logger) error {
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
				MaintenanceInfoVersion: instance.ServicePlanMaintenanceInfoVersion,
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

func discoverUpgradeableInstances(api CFClient, servicePlans []ccapi.ServicePlan, log Logger) ([]ccapi.ServiceInstance, int, error) {
	serviceInstances, err := api.GetServiceInstancesForServicePlans(servicePlans)
	if err != nil {
		return nil, 0, err
	}

	var upgradableInstances []ccapi.ServiceInstance
	for _, i := range serviceInstances {
		if i.UpgradeAvailable && isCreateFailed(i.LastOperationType, i.LastOperationState) {
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
