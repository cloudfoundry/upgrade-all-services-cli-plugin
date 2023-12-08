package ccapi

import (
	"fmt"
)

type ServicePlan struct {
	GUID                   string
	Available              bool
	Name                   string
	MaintenanceInfoVersion string
	ServiceOfferingGUID    string
	ServiceOfferingName    string
}

func (c CCAPI) GetServicePlans(brokerName string) ([]ServicePlan, error) {

	type plan struct {
		GUID                        string `json:"guid"`
		Available                   bool   `json:"available"`
		Name                        string `json:"name"`
		MaintenanceInfoVersion      string `jsonry:"maintenance_info.version"`
		IncludedServiceOfferingGUID string `jsonry:"relationships.service_offering.data.guid"`
	}

	type serviceOffering struct {
		GUID string `json:"guid"`
		Name string `json:"name"`
	}

	type includedServiceOfferings struct {
		ServiceOfferings []serviceOffering `json:"service_offerings"`
	}

	var receiver struct {
		Plans             []plan                   `json:"resources"`
		IncludedResources includedServiceOfferings `json:"included"`
	}

	if err := c.requester.Get(fmt.Sprintf("v3/service_plans?include=service_offering&per_page=5000&service_broker_names=%s", brokerName), &receiver); err != nil {
		return nil, fmt.Errorf("error getting service plans: %w", err)
	}

	var plans []ServicePlan

	serviceOfferingLookup := make(map[string]serviceOffering, len(receiver.IncludedResources.ServiceOfferings))
	for _, offering := range receiver.IncludedResources.ServiceOfferings {
		serviceOfferingLookup[offering.GUID] = offering
	}

	for _, p := range receiver.Plans {

		sp := ServicePlan{
			GUID:                   p.GUID,
			Available:              p.Available,
			Name:                   p.Name,
			MaintenanceInfoVersion: p.MaintenanceInfoVersion,
		}

		offering := serviceOfferingLookup[p.IncludedServiceOfferingGUID]
		sp.ServiceOfferingGUID = offering.GUID
		sp.ServiceOfferingName = offering.Name

		plans = append(plans, sp)
	}

	return plans, nil
}
