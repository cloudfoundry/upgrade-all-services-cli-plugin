package upgrader_test

import (
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/upgrader"
	"upgrade-all-services-cli-plugin/internal/upgrader/upgraderfakes"

	"github.com/hashicorp/go-version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("--min-version-required", func() {
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

	When("services are at specified version", func() {
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
					Action:     config.MinVersionCheckAction,
					MinVersion: version.Must(version.NewVersion("1.2.3")),
				})
				Expect(err).NotTo(HaveOccurred())
			})

			Expect(output).NotTo(ContainSubstring(fakeInstanceGUID))
		})
	})

	When("services are below specified version", func() {
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

		It("returns an error", func() {
			output := captureStdout(func() {
				err := upgrader.Upgrade(fakeCFClient, fakeLogger, upgrader.UpgradeConfig{
					BrokerName: fakeBrokerName,
					Action:     config.MinVersionCheckAction,
					MinVersion: version.Must(version.NewVersion("1.2.4")),
				})
				Expect(err).To(MatchError("found 1 service instances with a version less than the minimum required"))
				Expect(err).To(BeAssignableToTypeOf(upgrader.InstanceError{}))
			})

			Expect(output).To(ContainSubstring(fakeInstanceGUID))
		})
	})

	When("a service instance has a malformed version", func() {
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
					MaintenanceInfoVersion:            "malformed",
					ServicePlanMaintenanceInfoVersion: "1.2.3",
					ServicePlanDeactivated:            false,
				},
			}, nil)
		})

		It("returns an error", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeLogger, upgrader.UpgradeConfig{
				BrokerName: fakeBrokerName,
				Action:     config.MinVersionCheckAction,
				MinVersion: version.Must(version.NewVersion("1.2.3")),
			})
			Expect(err).To(MatchError("incorrect instance version: Malformed version: malformed"))
		})
	})
})
