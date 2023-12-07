package ccapi

import (
	"fmt"
	"strings"
)

type ServiceInstance struct {
	// These elements are retrieved directly from the service instance object
	GUID                     string `json:"guid"`
	Name                     string `json:"name"`
	UpgradeAvailable         bool   `json:"upgrade_available"`
	ServicePlanGUID          string `jsonry:"relationships.service_plan.data.guid"`
	SpaceGUID                string `jsonry:"relationships.space.data.guid"`
	LastOperationType        string `jsonry:"last_operation.type"`
	LastOperationState       string `jsonry:"last_operation.state"`
	LastOperationDescription string `jsonry:"last_operation.description"`
	MaintenanceInfoVersion   string `jsonry:"maintenance_info.version"`

	// These elements are retrieved from other resources returned by the API
	ServicePlanName     string `json:"-"`
	ServiceOfferingGUID string `json:"-"`
	ServiceOfferingName string `json:"-"`
	SpaceName           string `json:"-"`
	OrganizationGUID    string `json:"-"`
	OrganizationName    string `json:"-"`

	ServicePlanMaintenanceInfoVersion string `json:"-"`
	ServicePlanDeactivated            bool   `json:"-"`
}

type includedPlan struct {
	GUID                string `json:"guid"`
	Name                string `json:"name"`
	ServiceOfferingGUID string `jsonry:"relationships.service_offering.data.guid"`
}

type includedSpace struct {
	GUID             string `json:"guid"`
	Name             string `json:"name"`
	OrganizationGUID string `jsonry:"relationships.organization.data.guid"`
}

type includedOrganization struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}

type includedServiceOffering struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}

func BuildQueryParams(planGUIDs []string) string {
	return fmt.Sprintf("per_page=5000&fields[space]=name,guid,relationships.organization&fields[space.organization]=name,guid&fields[service_plan]=name,guid,relationships.service_offering&fields[service_plan.service_offering]=guid,name&service_plan_guids=%s", strings.Join(planGUIDs, ","))
}

func (c CCAPI) GetServiceInstancesByServicePlans(plans []ServicePlan) ([]ServiceInstance, error) {

	var receiver struct {
		Instances []ServiceInstance `json:"resources"`
		Included  struct {
			Plans            []includedPlan            `json:"service_plans"`
			Spaces           []includedSpace           `json:"spaces"`
			Organizations    []includedOrganization    `json:"organizations"`
			ServiceOfferings []includedServiceOffering `json:"service_offerings"`
		} `json:"included"`
	}

	if err := c.requester.Get("v3/service_instances?"+BuildQueryParams(getPlansGUIDs(plans)), &receiver); err != nil {
		return nil, fmt.Errorf("error getting service instances: %s", err)
	}

	instances := make([]ServiceInstance, 0, len(receiver.Instances))

	// Enrich with service plan, service offering space, and org data
	spaceGUIDLookup := computeSpaceGUIDLookup(receiver.Included.Spaces, receiver.Included.Organizations)
	for _, instance := range receiver.Instances {
		plan := getPlanByGUID(plans, instance.ServicePlanGUID)
		instance.ServicePlanName = plan.Name
		instance.ServiceOfferingGUID = plan.ServiceOffering.GUID
		instance.ServiceOfferingName = plan.ServiceOffering.Name
		instance.ServicePlanMaintenanceInfoVersion = plan.MaintenanceInfoVersion
		instance.ServicePlanDeactivated = !plan.Available

		spaceName, orgGUID, orgName := spaceGUIDLookup(instance.SpaceGUID)
		instance.SpaceName = spaceName
		instance.OrganizationGUID = orgGUID
		instance.OrganizationName = orgName
		instances = append(instances, instance)
	}

	return instances, nil
}

func getPlanByGUID(plans []ServicePlan, guid string) ServicePlan {
	for _, plan := range plans {
		if guid == plan.GUID {
			return plan
		}
	}
	return ServicePlan{}
}

func computeSpaceGUIDLookup(spaces []includedSpace, orgs []includedOrganization) func(key string) (string, string, string) {
	orgLookup := make(map[string]string)
	for _, o := range orgs {
		orgLookup[o.GUID] = o.Name
	}

	type entry struct {
		spaceName string
		orgGUID   string
		orgName   string
	}
	spaceLookup := make(map[string]entry)
	for _, s := range spaces {
		spaceLookup[s.GUID] = entry{
			spaceName: s.Name,
			orgGUID:   s.OrganizationGUID,
			orgName:   orgLookup[s.OrganizationGUID],
		}
	}

	return func(spaceGUID string) (string, string, string) {
		e := spaceLookup[spaceGUID]
		return e.spaceName, e.orgGUID, e.orgName
	}
}

func getPlansGUIDs(plans []ServicePlan) []string {
	guids := make([]string, 0, len(plans))
	for _, plan := range plans {
		guids = append(guids, plan.GUID)
	}
	return guids
}
