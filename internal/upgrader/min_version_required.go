package upgrader

import (
	"encoding/json"
	"fmt"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/slicex"
	"upgrade-all-services-cli-plugin/internal/versionchecker"

	"code.cloudfoundry.org/jsonry"
	"github.com/hashicorp/go-version"
)

func performMinimumVersionRequiredCheck(api CFClient, cfg UpgradeConfig) error {
	serviceInstances, err := getAllServiceInstances(api, cfg.BrokerName)
	if err != nil {
		return err
	}

	filteredInstances, err := filterInstancesVersionLessThanMinimumVersionRequired(serviceInstances, cfg.MinVersion)
	if err != nil {
		return err
	}

	switch cfg.JSONOutput {
	case true:
		return outputMinimumVersionJSON(filteredInstances)
	default:
		return outputMinimumVersionText(filteredInstances, len(serviceInstances), cfg.BrokerName, cfg.MinVersion.String())
	}
}

func outputMinimumVersionText(filteredInstances []ccapi.ServiceInstance, totalServiceInstances int, brokerName, minVersion string) error {
	fmt.Printf("Discovering service instances for broker: %s\n", brokerName)
	fmt.Printf("Total number of service instances: %d\n", totalServiceInstances)
	if len(filteredInstances) == 0 {
		fmt.Printf("No instances found with version lower than %q\n", minVersion)
		return nil
	}

	fmt.Printf("Number of service instances with a version lower than %q: %d\n", minVersion, len(filteredInstances))
	fmt.Println()
	for _, instance := range filteredInstances {
		fmt.Printf("  Service Instance Name: %q\n", instance.Name)
		fmt.Printf("  Service Instance GUID: %q\n", instance.GUID)
		fmt.Printf("  Service Instance Version: %q\n", instance.MaintenanceInfoVersion)
		fmt.Printf("  Service Plan Name: %q\n", instance.ServicePlanName)
		fmt.Printf("  Service Plan GUID: %q\n", instance.ServicePlanGUID)
		fmt.Printf("  Service Plan Version: %q\n", instance.ServicePlanMaintenanceInfoVersion)
		fmt.Printf("  Service Offering Name: %q\n", instance.ServiceOfferingName)
		fmt.Printf("  Service Offering GUID: %q\n", instance.ServiceOfferingGUID)
		fmt.Printf("  Space Name: %q\n", instance.SpaceName)
		fmt.Printf("  Space GUID: %q\n", instance.SpaceGUID)
		fmt.Printf("  Organization Name: %q\n", instance.OrganizationName)
		fmt.Printf("  Organization GUID: %q\n", instance.OrganizationGUID)
		fmt.Println()
	}

	return fmt.Errorf("found %d service instances with a version less than the minimum required", len(filteredInstances))
}

func outputMinimumVersionJSON(filteredInstances []ccapi.ServiceInstance) error {
	lines := slicex.Map(filteredInstances, newMinVersionRequiredLineFromServiceInstance)

	output, err := json.MarshalIndent(lines, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(output))

	// In contrast to text output, we don't exit with an error code. The rationale is that JSON may be piped
	// to a processor command like `jq`, and a command failure would be more of a hindrance than a help.
	return nil
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

func newMinVersionRequiredLineFromServiceInstance(instance ccapi.ServiceInstance) minVersionRequiredLine {
	return minVersionRequiredLine{
		Name:         instance.Name,
		GUID:         instance.GUID,
		Version:      instance.MaintenanceInfoVersion,
		SpaceName:    instance.SpaceName,
		SpaceGUID:    instance.SpaceGUID,
		OrgName:      instance.OrganizationName,
		OrgGUID:      instance.OrganizationGUID,
		PlanName:     instance.ServicePlanName,
		PlanGUID:     instance.ServicePlanGUID,
		OfferingName: instance.ServiceOfferingName,
		OfferingGUID: instance.ServiceOfferingGUID,
	}
}

type minVersionRequiredLine struct {
	Name         string `json:"name"`
	GUID         string `json:"guid"`
	Version      string `jsonry:"maintenance_info.version"`
	SpaceName    string `jsonry:"space.name"`
	SpaceGUID    string `jsonry:"space.guid"`
	OrgName      string `jsonry:"organization.name"`
	OrgGUID      string `jsonry:"organization.guid"`
	PlanName     string `jsonry:"service_plan.name"`
	PlanGUID     string `jsonry:"service_plan.guid"`
	OfferingName string `jsonry:"service_offering.name"`
	OfferingGUID string `jsonry:"service_offering.guid"`
}

func (m minVersionRequiredLine) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(m)
}
