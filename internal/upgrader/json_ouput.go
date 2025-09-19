package upgrader

import (
	"upgrade-all-services-cli-plugin/internal/ccapi"

	"code.cloudfoundry.org/jsonry"
)

func newJSONOutputServiceInstance(instance ccapi.ServiceInstance) jsonOutputServiceInstance {
	return jsonOutputServiceInstance{
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

type jsonOutputServiceInstance struct {
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

func (m jsonOutputServiceInstance) MarshalJSON() ([]byte, error) {
	return jsonry.Marshal(m)
}
