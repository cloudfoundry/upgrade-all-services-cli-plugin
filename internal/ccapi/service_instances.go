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

	// Can't be retrieves from CF API using `fields` query parameter
	// We populate this field in Upgrade function in internal/upgrader/upgrader.go
	PlanMaintenanceInfoVersion string `json:"-"`
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

func (c CCAPI) GetServiceInstances(planGUIDs []string) ([]ServiceInstance, error) {
	if len(planGUIDs) == 0 {
		return nil, fmt.Errorf("no service_plan_guids specified")
	}

	var receiver struct {
		Instances []ServiceInstance `json:"resources"`
		Included  struct {
			Plans            []includedPlan            `json:"service_plans"`
			Spaces           []includedSpace           `json:"spaces"`
			Organizations    []includedOrganization    `json:"organizations"`
			ServiceOfferings []includedServiceOffering `json:"service_offerings"`
		} `json:"included"`
	}

	if err := c.requester.Get("v3/service_instances?"+BuildQueryParams(planGUIDs), &receiver); err != nil {
		return nil, fmt.Errorf("error getting service instances: %s", err)
	}

	// Enrich with service plan and service offering data
	servicePlanGUIDLookup := computeServicePlanGUIDLookup(receiver.Included.Plans, receiver.Included.ServiceOfferings)
	for i := range receiver.Instances {
		planName, offeringGUID, offeringName := servicePlanGUIDLookup(receiver.Instances[i].ServicePlanGUID)
		receiver.Instances[i].ServicePlanName = planName
		receiver.Instances[i].ServiceOfferingGUID = offeringGUID
		receiver.Instances[i].ServiceOfferingName = offeringName
	}

	// Enrich with space and org data
	spaceGUIDLookup := computeSpaceGUIDLookup(receiver.Included.Spaces, receiver.Included.Organizations)
	for i := range receiver.Instances {
		spaceName, orgGUID, orgName := spaceGUIDLookup(receiver.Instances[i].SpaceGUID)
		receiver.Instances[i].SpaceName = spaceName
		receiver.Instances[i].OrganizationGUID = orgGUID
		receiver.Instances[i].OrganizationName = orgName
	}

	return receiver.Instances, nil
}

func computeServicePlanGUIDLookup(plans []includedPlan, offerings []includedServiceOffering) func(key string) (string, string, string) {
	offeringLookup := make(map[string]string)
	for _, o := range offerings {
		offeringLookup[o.GUID] = o.Name
	}

	type entry struct {
		planName     string
		offeringGUID string
		offeringName string
	}
	planLookup := make(map[string]entry)
	for _, p := range plans {
		planLookup[p.GUID] = entry{
			planName:     p.Name,
			offeringGUID: p.ServiceOfferingGUID,
			offeringName: offeringLookup[p.ServiceOfferingGUID],
		}
	}

	return func(planGUID string) (string, string, string) {
		e := planLookup[planGUID]
		return e.planName, e.offeringGUID, e.offeringName
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
