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

type includedSpace struct {
	GUID             string `json:"guid"`
	Name             string `json:"name"`
	OrganizationGUID string `jsonry:"relationships.organization.data.guid"`
}

type includedOrganization struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}

func BuildQueryParams(planGUIDs []string) string {
	return fmt.Sprintf("per_page=5000&fields[space]=name,guid,relationships.organization&fields[space.organization]=name,guid&service_plan_guids=%s", strings.Join(planGUIDs, ","))
}

func (c CCAPI) GetServiceInstancesForServicePlans(plans []ServicePlan) ([]ServiceInstance, error) {

	var receiver struct {
		Instances []ServiceInstance `json:"resources"`
		Included  struct {
			Spaces        []includedSpace        `json:"spaces"`
			Organizations []includedOrganization `json:"organizations"`
		} `json:"included"`
	}

	if err := c.requester.Get("v3/service_instances?"+BuildQueryParams(getPlansGUIDs(plans)), &receiver); err != nil {
		return nil, fmt.Errorf("error getting service instances: %s", err)
	}

	// Enrich with service plan, service offering space, and org data
	spaceGUIDLookup := computeSpaceGUIDLookup(receiver.Included.Spaces, receiver.Included.Organizations)
	planGUIDLookup := computePlanGUIDLookup(plans)
	for i := range receiver.Instances {
		plan := planGUIDLookup(receiver.Instances[i].ServicePlanGUID)
		receiver.Instances[i].ServicePlanName = plan.Name
		receiver.Instances[i].ServiceOfferingGUID = plan.ServiceOffering.GUID
		receiver.Instances[i].ServiceOfferingName = plan.ServiceOffering.Name
		receiver.Instances[i].ServicePlanMaintenanceInfoVersion = plan.MaintenanceInfoVersion
		receiver.Instances[i].ServicePlanDeactivated = !plan.Available

		spaceName, orgGUID, orgName := spaceGUIDLookup(receiver.Instances[i].SpaceGUID)
		receiver.Instances[i].SpaceName = spaceName
		receiver.Instances[i].OrganizationGUID = orgGUID
		receiver.Instances[i].OrganizationName = orgName
	}

	return receiver.Instances, nil
}

func computePlanGUIDLookup(plans []ServicePlan) func(guid string) ServicePlan {
	plansLookup := make(map[string]ServicePlan, len(plans))
	for _, plan := range plans {
		plansLookup[plan.GUID] = plan
	}
	return func(guid string) ServicePlan {
		return plansLookup[guid]
	}
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
