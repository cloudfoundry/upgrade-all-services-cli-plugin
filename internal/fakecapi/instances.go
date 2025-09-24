package fakecapi

import (
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"
	"upgrade-all-services-cli-plugin/internal/slicex"

	"code.cloudfoundry.org/jsonry"
)

const (
	spaceGUID = "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
	orgGUID   = "1a2f43b5-1594-4247-a888-e8843ebd1b03"
)

func WithServiceInstances(instances ...ServiceInstance) func(*FakeCAPI, ServicePlan) {
	return func(f *FakeCAPI, plan ServicePlan) {
		for _, instance := range instances {
			if instance.Name == "" {
				instance.Name = f.fakeName("instance")
			}
			if instance.GUID == "" {
				instance.GUID = stableGUID(instance.Name)
			}
			if instance.LastOperationType == "" {
				instance.LastOperationType = "update"
			}
			if instance.LastOperationState == "" {
				instance.LastOperationState = "succeeded"
			}
			instance.ServicePlanName = plan.Name
			instance.ServicePlanGUID = plan.GUID
			instance.ServiceOfferingName = plan.ServiceOfferingName
			instance.ServiceOfferingGUID = plan.ServiceOfferingGUID
			instance.SpaceGUID = spaceGUID

			f.instances[instance.GUID] = &instance
		}
	}
}

type ServiceInstance struct {
	Name                     string        `json:"name"`
	GUID                     string        `json:"guid"`
	ServicePlanGUID          string        `jsonry:"relationships.service_plan.data.guid"`
	SpaceGUID                string        `jsonry:"relationships.space.data.guid"`
	ServicePlanName          string        `json:"-"`
	ServiceOfferingGUID      string        `json:"-"`
	ServiceOfferingName      string        `json:"-"`
	Version                  string        `jsonry:"maintenance_info.version"`
	UpgradeAvailable         bool          `json:"upgrade_available"`
	LastOperationType        string        `jsonry:"last_operation.type"`
	LastOperationState       string        `jsonry:"last_operation.state"`
	LastOperationDescription string        `jsonry:"last_operation.description"`
	UpdateTime               time.Duration `json:"-"`
	UpdateCount              int           `json:"-"`
	FailTimes                int           `json:"-"`
	Callback                 func()        `json:"-"`
}

type Space struct {
	Name             string `json:"name"`
	GUID             string `json:"guid"`
	OrganizationName string `jsonry:"relationships.organization.date.name"`
	OrganizationGUID string `jsonry:"relationships.organization.data.guid"`
}

type Org struct {
	Name string `json:"name"`
	GUID string `json:"guid"`
}

func (f *FakeCAPI) listServiceInstancesHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		includeSpaces, includeOrgs := false, false

		instances := slices.Collect(maps.Values(f.instances))
		for k := range r.URL.Query() {
			v := r.URL.Query().Get(k)
			switch {
			case k == "per_page": // ignore
			case k == "fields[space]" && v == "name,guid,relationships.organization":
				includeSpaces = true
			case k == "fields[space.organization]" && v == "name,guid":
				includeOrgs = true
			case k == "service_plan_guids":
				instances = slicex.Filter(instances, func(p *ServiceInstance) bool { return slices.Contains(strings.Split(v, ","), p.ServicePlanGUID) })
			default:
				http.Error(w, fmt.Sprintf("unknown query filter %q with value %q", k, v), http.StatusBadRequest)
				return
			}
		}

		var includedSpaces []Space
		if includeSpaces {
			includedSpaces = append(includedSpaces, Space{
				Name:             "fake-space",
				GUID:             spaceGUID,
				OrganizationName: "fake-org",
				OrganizationGUID: orgGUID,
			})
		}

		var includedOrgs []Org
		if includeOrgs {
			includedOrgs = append(includedOrgs, Org{
				Name: "fake-org",
				GUID: orgGUID,
			})
		}

		slices.SortStableFunc(instances, func(a, b *ServiceInstance) int { return strings.Compare(a.Name, b.Name) })

		payload, err := jsonry.Marshal(struct {
			Resources      []*ServiceInstance `json:"resources"`
			IncludedSpaces []Space            `jsonry:"included.spaces,omitempty"`
			IncludedOrgs   []Org              `jsonry:"included.organizations,omitempty"`
		}{Resources: instances, IncludedSpaces: includedSpaces, IncludedOrgs: includedOrgs})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(payload)
	}
}

func (f *FakeCAPI) getServiceInstanceHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		guid := r.PathValue("guid")
		instance, ok := f.instances[guid]
		if !ok {
			http.Error(w, fmt.Sprintf("instance with guid %q not found", guid), http.StatusNotFound)
			return
		}

		response, err := jsonry.Marshal(instance)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling service instance: %s", err), http.StatusInternalServerError)
		}

		w.Write(response)
	}
}

func (f *FakeCAPI) updateServiceInstanceHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		guid := r.PathValue("guid")
		instance, ok := f.instances[guid]
		if !ok {
			http.Error(w, fmt.Sprintf("instance with guid %q not found", guid), http.StatusNotFound)
			return
		}

		if instance.Callback != nil {
			instance.Callback()
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading body: %s", err), http.StatusInternalServerError)
			return
		}

		var receiver struct {
			Version string `jsonry:"maintenance_info.version"`
		}
		if err := jsonry.Unmarshal(data, &receiver); err != nil {
			http.Error(w, fmt.Sprintf("error parsing body: %s", err), http.StatusBadRequest)
		}

		if planVersion := f.plans[instance.ServicePlanGUID].Version; receiver.Version != planVersion {
			http.Error(w, "plan version %q does not match requested version %q", http.StatusBadRequest)
			return
		}

		f.startOperation()
		instance.LastOperationType = "update"
		instance.LastOperationState = "in progress"
		instance.LastOperationDescription = "update operation started"

		response, err := jsonry.Marshal(instance)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshaling service instance: %s", err), http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write(response)

		go func() {
			time.Sleep(instance.UpdateTime)
			f.stopOperation()
			instance.UpdateCount++

			if instance.FailTimes > 0 {
				instance.FailTimes--
				instance.LastOperationState = "failed"
				instance.LastOperationDescription = "failed as requested by test setup"
				return
			}

			instance.LastOperationState = "succeeded"
			instance.LastOperationDescription = "succeeded as requested by test setup"
		}()
	}
}
