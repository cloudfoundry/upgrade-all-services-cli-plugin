package upgrader

import (
	"errors"
	"fmt"
	"time"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/versionchecker"
	"upgrade-all-services-cli-plugin/internal/workers"

	"github.com/hashicorp/go-version"
)

//go:generate go tool counterfeiter -generate
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
	HasUpgradeSucceeded() bool
	FinalTotals()
}

type UpgradeConfig struct {
	BrokerName       string
	ParallelUpgrades int
	Action           config.Action
	MinVersion       *version.Version
}

func Upgrade(api CFClient, log Logger, cfg UpgradeConfig) error {
	servicePlans, err := api.GetServicePlans(cfg.BrokerName)
	if err != nil {
		return err
	}

	if len(servicePlans) == 0 {
		return fmt.Errorf("no service plans available for broker: %s", cfg.BrokerName)
	}

	log.Printf("discovering service instances for broker: %s", cfg.BrokerName)
	serviceInstances, err := api.GetServiceInstancesForServicePlans(servicePlans)
	if err != nil {
		return err
	}

	upgradableInstances := discoverInstancesWithPendingUpgrade(log, serviceInstances)

	switch {
	case cfg.Action == config.CheckDeactivatedPlansAction:
		return checkDeactivatedPlans(log, serviceInstances)
	case cfg.Action == config.CheckUpToDateAction:
		log.InitialTotals(len(serviceInstances), len(upgradableInstances))
		defer log.FinalTotals()
		var errs MultiError
		performDryRun(upgradableInstances, log)
		if len(upgradableInstances) > 0 {
			errs.Append(fmt.Errorf("found %d instances which are not up-to-date", len(upgradableInstances)))
		}
		errs.Append(checkDeactivatedPlans(log, serviceInstances))
		if len(errs.Errors) > 0 {
			return &errs
		}

		return nil
	case cfg.Action == config.MinVersionCheckAction:
		filteredInstances, err := filterInstancesVersionLessThanMinimumVersionRequired(serviceInstances, cfg.MinVersion)
		if err != nil {
			return err
		}

		if len(filteredInstances) == 0 {
			log.Printf("no instances found with version less than required")
			return nil
		}

		log.InitialTotals(len(serviceInstances), len(filteredInstances))
		defer log.FinalTotals()
		performDryRun(filteredInstances, log)
		return fmt.Errorf("found %d service instances with a version less than the minimum required", len(filteredInstances))
	case len(upgradableInstances) == 0:
		log.Printf("no instances available to upgrade")
		return nil
	case cfg.Action == config.DryRunAction:
		log.InitialTotals(len(serviceInstances), len(upgradableInstances))
		defer log.FinalTotals()
		performDryRun(upgradableInstances, log)
		return nil
	default:
		log.InitialTotals(len(serviceInstances), len(upgradableInstances))
		defer log.FinalTotals()
		return performUpgrade(api, upgradableInstances, cfg.ParallelUpgrades, log)
	}
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

func checkDeactivatedPlans(log Logger, instances []ccapi.ServiceInstance) error {
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

	if !log.HasUpgradeSucceeded() {
		return errors.New("there were failures upgrading one or more instances. Review the logs for more information")
	}
	return nil
}

func performDryRun(serviceInstances []ccapi.ServiceInstance, log Logger) {
	for _, i := range serviceInstances {
		dryRunErr := fmt.Errorf("dry-run prevented upgrade instance guid %s", i.GUID)
		log.UpgradeFailed(i, time.Duration(0), dryRunErr)
	}
}

func discoverInstancesWithPendingUpgrade(log Logger, serviceInstances []ccapi.ServiceInstance) []ccapi.ServiceInstance {
	var instancesWithPendingUpgrade []ccapi.ServiceInstance
	for _, i := range serviceInstances {
		if !i.UpgradeAvailable {
			continue
		}

		if ccapi.HasInstanceCreateFailedStatus(i) {
			log.SkippingInstance(i)
			continue
		}

		instancesWithPendingUpgrade = append(instancesWithPendingUpgrade, i)
	}

	return instancesWithPendingUpgrade
}
