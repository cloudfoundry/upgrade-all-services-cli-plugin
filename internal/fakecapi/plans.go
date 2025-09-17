package fakecapi

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"upgrade-all-services-cli-plugin/internal/slicex"

	"code.cloudfoundry.org/jsonry"
)

func WithServicePlan(plan ServicePlan, opts ...func(*FakeCAPI, ServicePlan)) func(*FakeCAPI, ServiceOffering) {
	return func(f *FakeCAPI, offering ServiceOffering) {
		if plan.Name == "" {
			plan.Name = f.fakeName("plan")
		}
		if plan.GUID == "" {
			plan.GUID = stableGUID(plan.Name)
		}
		plan.ServiceOfferingName = offering.Name
		plan.ServiceOfferingGUID = offering.GUID

		f.plans[plan.GUID] = plan

		for _, opt := range opts {
			opt(f, plan)
		}
	}
}

type ServicePlan struct {
	Name                string `json:"name"`
	GUID                string `json:"guid"`
	Version             string `jsonry:"maintenance_info.version"`
	Available           bool   `json:"available"`
	ServiceOfferingName string `json:"-"`
	ServiceOfferingGUID string `jsonry:"relationships.service_offering.data.guid"`
}

func (f *FakeCAPI) listServicePlansHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		includeServiceOffering := false

		plans := slices.Collect(maps.Values(f.plans))
		for k := range r.URL.Query() {
			v := r.URL.Query().Get(k)
			switch {
			case k == "per_page": // ignore
			case k == "include" && v == "service_offering":
				includeServiceOffering = true
			case k == "service_broker_names":
				plans = slicex.Filter(plans, func(p ServicePlan) bool {
					return slices.Contains(strings.Split(v, ","), f.offerings[p.ServiceOfferingGUID].ServiceBrokerName)
				})
			default:
				http.Error(w, fmt.Sprintf("unknown query filter %q with value %q", k, v), http.StatusBadRequest)
				return
			}
		}

		includedOfferings := make(map[string]ServiceOffering)
		if includeServiceOffering {
			for _, p := range plans {
				includedOfferings[p.ServiceOfferingGUID] = f.offerings[p.ServiceOfferingGUID]
			}
		}

		slices.SortStableFunc(plans, func(a, b ServicePlan) int { return strings.Compare(a.Name, b.Name) })

		payload, err := jsonry.Marshal(struct {
			Resources         []ServicePlan     `json:"resources"`
			IncludedOfferings []ServiceOffering `jsonry:"included.service_offerings,omitempty"`
		}{Resources: plans, IncludedOfferings: slices.Collect(maps.Values(includedOfferings))})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(payload)
	}
}
