package upgrader_test

import (
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/upgrader"
	"upgrade-all-services-cli-plugin/internal/upgrader/upgraderfakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("--check-up-to-date", func() {
	const (
		fakeBrokerName   = "fake-broker-name"
		fakePlanGUID     = "fake-plan-guid"
		fakeInstanceGUID = "fake-instance-guid"
	)

	var (
		fakeCFClient *upgraderfakes.FakeCFClient
		fakeLogger   *upgraderfakes.FakeLogger
	)

	BeforeEach(func() {
		fakeCFClient = &upgraderfakes.FakeCFClient{}
		fakeLogger = &upgraderfakes.FakeLogger{}
	})

	When("all service instances are up to date", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
				{GUID: fakePlanGUID, Available: true, MaintenanceInfoVersion: "1.2.3"},
			}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{
				{
					GUID:                              fakeInstanceGUID,
					UpgradeAvailable:                  false,
					ServicePlanGUID:                   fakePlanGUID,
					LastOperationType:                 "create",
					LastOperationState:                "succeeded",
					MaintenanceInfoVersion:            "1.2.3",
					ServicePlanMaintenanceInfoVersion: "1.2.3",
					ServicePlanDeactivated:            false,
				},
			}, nil)
		})

		It("succeeds", func() {
			output := captureStdout(func() {
				err := upgrader.Upgrade(fakeCFClient, fakeLogger, upgrader.UpgradeConfig{
					BrokerName: fakeBrokerName,
					Action:     config.CheckUpToDateAction,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).NotTo(ContainSubstring(fakeInstanceGUID))
		})
	})

	When("there are deactivated plans", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
				{GUID: fakePlanGUID, Available: false, MaintenanceInfoVersion: "1.2.3"},
			}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{
				{
					GUID:                              fakeInstanceGUID,
					UpgradeAvailable:                  false,
					ServicePlanGUID:                   fakePlanGUID,
					LastOperationType:                 "create",
					LastOperationState:                "succeeded",
					MaintenanceInfoVersion:            "1.2.3",
					ServicePlanMaintenanceInfoVersion: "1.2.3",
					ServicePlanDeactivated:            true,
				},
			}, nil)
		})

		It("returns an error", func() {
			output := captureStdout(func() {
				err := upgrader.Upgrade(fakeCFClient, fakeLogger, upgrader.UpgradeConfig{
					BrokerName: fakeBrokerName,
					Action:     config.CheckUpToDateAction,
				})
				Expect(err).To(MatchError("discovered service instances associated with deactivated plans or with an upgrade available"))
				Expect(err).To(BeAssignableToTypeOf(upgrader.InstanceError{}))
			})

			Expect(output).To(ContainSubstring(fakeInstanceGUID))
		})
	})

	When("there are outdated service instances", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
				{GUID: fakePlanGUID, Available: true, MaintenanceInfoVersion: "1.2.3"},
			}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{
				{
					GUID:                              fakeInstanceGUID,
					UpgradeAvailable:                  true,
					ServicePlanGUID:                   fakePlanGUID,
					LastOperationType:                 "create",
					LastOperationState:                "succeeded",
					MaintenanceInfoVersion:            "1.2.2",
					ServicePlanMaintenanceInfoVersion: "1.2.3",
					ServicePlanDeactivated:            false,
				},
			}, nil)
		})

		It("returns an error", func() {
			output := captureStdout(func() {
				err := upgrader.Upgrade(fakeCFClient, fakeLogger, upgrader.UpgradeConfig{
					BrokerName: fakeBrokerName,
					Action:     config.CheckUpToDateAction,
				})
				Expect(err).To(MatchError("discovered service instances associated with deactivated plans or with an upgrade available"))
				Expect(err).To(BeAssignableToTypeOf(upgrader.InstanceError{}))
			})

			Expect(output).To(ContainSubstring(fakeInstanceGUID))
		})
	})

	When("service instances failed to create", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
				{GUID: fakePlanGUID, Available: true, MaintenanceInfoVersion: "1.2.3"},
			}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{
				{
					GUID:                              fakeInstanceGUID,
					UpgradeAvailable:                  true,
					ServicePlanGUID:                   fakePlanGUID,
					LastOperationType:                 "create",
					LastOperationState:                "failed",
					MaintenanceInfoVersion:            "1.2.3",
					ServicePlanMaintenanceInfoVersion: "1.2.2",
					ServicePlanDeactivated:            false,
				},
			}, nil)
		})

		It("succeeds but logs the instance", func() {
			// Although the instance looks upgradeable, because the create failed, there isn't really an instance to upgrade.
			output := captureStdout(func() {
				err := upgrader.Upgrade(fakeCFClient, fakeLogger, upgrader.UpgradeConfig{
					BrokerName: fakeBrokerName,
					Action:     config.CheckUpToDateAction,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).To(ContainSubstring(fakeInstanceGUID))
		})
	})
})
