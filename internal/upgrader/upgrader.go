package upgrader

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/slicex"
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
	InitialTotals(totalServiceInstances, totalUpgradableServiceInstances int)
	HasUpgradeSucceeded() bool
	FinalTotals()
}

type UpgradeConfig struct {
	BrokerName       string
	ParallelUpgrades int
	Action           config.Action
	MinVersion       *version.Version
	JSONOutput       bool
	Limit            int
}

func Upgrade(api CFClient, log Logger, cfg UpgradeConfig) error {
	switch cfg.Action {
	case config.MinVersionCheckAction:
		return performMinimumVersionRequiredCheck(api, cfg)
	case config.CheckDeactivatedPlansAction:
		return performDeactivatedPlansCheck(api, cfg)
	case config.CheckUpToDateAction:
		return performUpToDateCheck(api, cfg)
	default: // continue function
	}

	instances, err := getGroupedServiceInstances(api, cfg.BrokerName, cfg.Limit)
	if err != nil {
		return err
	}

	switch {
	case cfg.Action == config.DryRunAction && cfg.JSONOutput:
		return outputDryRunJSON(instances.upgradeable, instances.createFailed)
	case cfg.Action == config.DryRunAction && !cfg.JSONOutput:
		return outputDryRunText(instances, log, cfg.BrokerName)
	default:
		return performUpgrade(api, instances, cfg.ParallelUpgrades, cfg.BrokerName, log)
	}
}

func performUpgrade(api CFClient, instances groupedServiceInstances, parallelUpgrades int, brokerName string, log Logger) error {
	log.Printf("discovering service instances for broker: %s", brokerName)
	log.InitialTotals(len(instances.all), len(instances.upgradeable))
	defer log.FinalTotals()
	for _, instance := range instances.createFailed {
		log.SkippingInstance(instance)
	}
	if len(instances.upgradeable) == 0 {
		log.Printf("no instances available to upgrade")
		return nil
	}

	type upgradeTask struct {
		UpgradeableIndex       int
		ServiceInstanceName    string
		ServiceInstanceGUID    string
		MaintenanceInfoVersion string
	}

	upgradeQueue := make(chan upgradeTask)
	go func() {
		for i, instance := range instances.upgradeable {
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
			log.UpgradeStarting(instances.upgradeable[instance.UpgradeableIndex])
			err := api.UpgradeServiceInstance(instance.ServiceInstanceGUID, instance.MaintenanceInfoVersion)
			switch err {
			case nil:
				log.UpgradeSucceeded(instances.upgradeable[instance.UpgradeableIndex], time.Since(start))
			default:
				log.UpgradeFailed(instances.upgradeable[instance.UpgradeableIndex], time.Since(start), err)
			}
		}
	})

	if !log.HasUpgradeSucceeded() {
		return errors.New("there were failures upgrading one or more instances. Review the logs for more information")
	}
	return nil
}

func outputDryRunText(instances groupedServiceInstances, log Logger, brokerName string) error {
	log.Printf("discovering service instances for broker: %s", brokerName)
	for _, instance := range instances.createFailed {
		log.SkippingInstance(instance)
	}

	if len(instances.upgradeable) == 0 {
		log.Printf("no instances available to upgrade")
		return nil
	}

	log.InitialTotals(len(instances.all), len(instances.upgradeable))
	defer log.FinalTotals()

	for _, i := range instances.upgradeable {
		dryRunErr := fmt.Errorf("dry-run prevented upgrade instance guid %s", i.GUID)
		log.UpgradeFailed(i, time.Duration(0), dryRunErr)
	}

	return nil
}

// outputDryRunJSON produces a JSON version of the dry run output. Unlike --check-up-to-date we do not
// output deactivated plans. This is to match existing behavior.
func outputDryRunJSON(upgradableInstances, createFailedInstances []ccapi.ServiceInstance) error {
	type formatter struct {
		UpgradePending []jsonOutputServiceInstance `json:"upgrade"`
		CreateFailed   []jsonOutputServiceInstance `json:"skip"`
	}

	data := formatter{
		UpgradePending: slicex.Map(upgradableInstances, newJSONOutputServiceInstance),
		CreateFailed:   slicex.Map(createFailedInstances, newJSONOutputServiceInstance),
	}

	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))

	return nil
}

func getAllServiceInstances(api CFClient, brokerName string) ([]ccapi.ServiceInstance, error) {
	servicePlans, err := api.GetServicePlans(brokerName)
	if err != nil {
		return nil, err
	}

	if len(servicePlans) == 0 {
		return nil, fmt.Errorf("no service plans available for broker: %s", brokerName)
	}

	return api.GetServiceInstancesForServicePlans(servicePlans)
}

type groupedServiceInstances struct {
	all, upgradeable, deactivatedPlan, createFailed []ccapi.ServiceInstance
}

// getGroupedServiceInstances will fetch all the service instances for a broker and group them into the following categories:
// - all: all service instances
// - deactivatedPlan - all service instances associated with a deactivated plan
// - createFailed - all service instances for which the UpgradeAvailable flag is set, but the instance failed to create
// - upgradeable - all service instances for which the UpgradeAvailable flag is set, bit the instance has been created successfully
func getGroupedServiceInstances(api CFClient, brokerName string, limit int) (groupedServiceInstances, error) {
	instances, err := getAllServiceInstances(api, brokerName)
	if err != nil {
		return groupedServiceInstances{}, err
	}

	deactivatedPlan := slicex.Filter(instances, func(instance ccapi.ServiceInstance) bool { return instance.ServicePlanDeactivated })
	upgradeAvailable := slicex.Filter(instances, func(instance ccapi.ServiceInstance) bool { return instance.UpgradeAvailable })
	createFailed, upgradeable := slicex.Partition(upgradeAvailable, ccapi.HasInstanceCreateFailedStatus)

	// If we have been asked to limit the number of instances upgraded, then apply that here
	if limit > 0 && len(upgradeable) > limit {
		upgradeable = upgradeable[:limit]
	}

	return groupedServiceInstances{
		all:             instances,
		deactivatedPlan: deactivatedPlan,
		createFailed:    createFailed,
		upgradeable:     upgradeable,
	}, nil
}
