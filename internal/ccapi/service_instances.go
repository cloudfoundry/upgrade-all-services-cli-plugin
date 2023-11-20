package ccapi

import (
	"fmt"
	"strings"
)

type ServiceInstance struct {
	GUID             string `json:"guid"`
	Name             string `json:"name"`
	UpgradeAvailable bool   `json:"upgrade_available"`
	ServicePlanGUID  string `jsonry:"relationships.service_plan.data.guid"`
	SpaceGUID        string `jsonry:"relationships.space.data.guid"`

	LastOperation LastOperation `json:"last_operation"`

	MaintenanceInfoVersion string `jsonry:"maintenance_info.version"`
	Included               EmbeddedInclude

	// Can't be retrieves from CF API using `fields` query parameter
	// We populate this field in Upgrade function in internal/upgrader/upgrader.go
	PlanMaintenanceInfoVersion string `json:"-"`
}

type LastOperation struct {
	Type        string `json:"type"`
	State       string `json:"state"`
	Description string `json:"description"`
}

type serviceInstances struct {
	Instances []ServiceInstance `json:"resources"`
	Included  struct {
		Plans            []IncludedPlan    `json:"service_plans"`
		Spaces           []Space           `json:"spaces"`
		Organizations    []Organization    `json:"organizations"`
		ServiceOfferings []ServiceOffering `json:"service_offerings"`
	} `json:"included"`
}

type EmbeddedInclude struct {
	Plan            IncludedPlan
	ServiceOffering ServiceOffering
	Space           Space
	Organization    Organization
}

type IncludedPlan struct {
	GUID                string `json:"guid"`
	Name                string `json:"name"`
	ServiceOfferingGUID string `jsonry:"relationships.service_offering.data.guid"`
}

type Space struct {
	GUID             string `json:"guid"`
	Name             string `json:"name"`
	OrganizationGUID string `jsonry:"relationships.organization.data.guid"`
}

type Organization struct {
	GUID string `json:"guid"`
	Name string `json:"name"`
}

type ServiceOffering struct {
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

	var si serviceInstances
	if err := c.requester.Get("v3/service_instances?"+BuildQueryParams(planGUIDs), &si); err != nil {
		return nil, fmt.Errorf("error getting service instances: %s", err)
	}
	embedIncludes(si)
	return si.Instances, nil
}

func embedIncludes(si serviceInstances) []ServiceInstance {
	orgs := make(map[string]Organization, len(si.Included.Organizations))
	for _, org := range si.Included.Organizations {
		orgs[org.GUID] = org
	}
	spaces := make(map[string]Space, len(si.Included.Spaces))
	for _, space := range si.Included.Spaces {
		spaces[space.GUID] = space
	}
	soffers := make(map[string]ServiceOffering, len(si.Included.ServiceOfferings))
	for _, soffer := range si.Included.ServiceOfferings {
		soffers[soffer.GUID] = soffer
	}
	plans := make(map[string]IncludedPlan, len(si.Included.Plans))
	for _, plan := range si.Included.Plans {
		plans[plan.GUID] = plan
	}

	for i, instance := range si.Instances {
		emb := EmbeddedInclude{}
		emb.Plan = plans[instance.ServicePlanGUID]
		emb.Space = spaces[instance.SpaceGUID]
		emb.ServiceOffering = soffers[plans[instance.ServicePlanGUID].ServiceOfferingGUID]
		emb.Organization = orgs[spaces[instance.SpaceGUID].OrganizationGUID]

		si.Instances[i].Included = emb
	}

	return si.Instances
}
