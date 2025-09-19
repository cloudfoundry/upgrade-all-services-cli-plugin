package upgrader

import (
	"fmt"
	"upgrade-all-services-cli-plugin/internal/ccapi"
)

func logServiceInstances(instances []ccapi.ServiceInstance) {
	for _, instance := range instances {
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
}
