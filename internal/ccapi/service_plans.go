package ccapi

import (
	"fmt"
)

type ServicePlan struct {
	GUID                   string `json:"guid"`
	MaintenanceInfoVersion string `jsonry:"maintenance_info.version"`
}

func (c CCAPI) GetServicePlans(brokerName string) ([]ServicePlan, error) {
	var receiver struct {
		Plans []ServicePlan `json:"resources"`
	}
	if err := c.requester.Get(fmt.Sprintf("v3/service_plans?per_page=5000&service_broker_names=%s", brokerName), &receiver); err != nil {
		return nil, fmt.Errorf("error getting service plans: %s", err)
	}
	return receiver.Plans, nil
}
