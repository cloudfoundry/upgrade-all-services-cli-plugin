package upgrader_test

import (
	"fmt"
	"testing"
	"upgrade-all-services-cli-plugin/internal/ccapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUpgrader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Upgrader Suite")
}

const fakeBrokerName = "fake-broker-name"

var (
	fakePlanCounter     int
	fakeInstanceCounter int
)

var _ = BeforeEach(func() {
	fakePlanCounter = 0
	fakeInstanceCounter = 0
})

// fakePlan generates a fake plan, using the specified plan as a guide
func fakePlan(plan ccapi.ServicePlan) ccapi.ServicePlan {
	fakePlanCounter++

	if plan.GUID == "" {
		plan.GUID = fmt.Sprintf("fake-plan-guid-%d", fakePlanCounter)
	}
	if plan.Name == "" {
		plan.Name = fmt.Sprintf("fake-plan-name-%d", fakePlanCounter)
	}
	if plan.ServiceOfferingGUID == "" {
		plan.ServiceOfferingGUID = fmt.Sprintf("fake-offering-guid-for-fake-plan-%d", fakePlanCounter)
	}
	if plan.ServiceOfferingName == "" {
		plan.ServiceOfferingGUID = fmt.Sprintf("fake-offering-name-for-fake-plan-%d", fakePlanCounter)
	}
	if plan.MaintenanceInfoVersion == "" {
		plan.MaintenanceInfoVersion = "1.2.3"
	}

	return plan
}

func fakeServiceInstance(instance ccapi.ServiceInstance) ccapi.ServiceInstance {
	fakeInstanceCounter++

	if instance.GUID == "" {
		instance.GUID = fmt.Sprintf("fake-instance-guid-%d", fakeInstanceCounter)
	}
	if instance.Name == "" {
		instance.Name = fmt.Sprintf("fake-instance-name-%d", fakeInstanceCounter)
	}
	if instance.ServicePlanGUID == "" {
		instance.ServicePlanGUID = fmt.Sprintf("fake-plan-guid-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.ServicePlanName == "" {
		instance.ServicePlanName = fmt.Sprintf("fake-plan-name-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.ServiceOfferingGUID == "" {
		instance.ServiceOfferingGUID = fmt.Sprintf("fake-offering-guid-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.ServiceOfferingName == "" {
		instance.ServiceOfferingName = fmt.Sprintf("fake-offering-name-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.SpaceGUID == "" {
		instance.SpaceGUID = fmt.Sprintf("fake-space-guid-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.SpaceName == "" {
		instance.SpaceName = fmt.Sprintf("fake-space-name-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.OrganizationGUID == "" {
		instance.OrganizationGUID = fmt.Sprintf("fake-org-guid-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.OrganizationName == "" {
		instance.OrganizationName = fmt.Sprintf("fake-org-name-for-fake-instance-%d", fakeInstanceCounter)
	}
	if instance.MaintenanceInfoVersion == "" {
		instance.MaintenanceInfoVersion = "1.2.3"
	}
	if instance.ServicePlanMaintenanceInfoVersion == "" {
		instance.ServicePlanMaintenanceInfoVersion = "1.2.3"
	}
	if instance.LastOperationType == "" {
		instance.LastOperationType = "create"
	}
	if instance.LastOperationState == "" {
		instance.LastOperationState = "succeeded"
	}
	if instance.LastOperationDescription == "" {
		instance.LastOperationDescription = "create successful"
	}

	return instance
}
