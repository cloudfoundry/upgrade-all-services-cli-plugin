package upgrader_test

import (
	"fmt"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/config"
	"upgrade-all-services-cli-plugin/internal/upgrader"
	"upgrade-all-services-cli-plugin/internal/upgrader/upgraderfakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check Deactivated Plans", func() {
	var (
		fakeCFClient *upgraderfakes.FakeCFClient
		fakeLog      *upgraderfakes.FakeLogger
		actionError  error
	)

	BeforeEach(func() {
		fakeCFClient = &upgraderfakes.FakeCFClient{}
		fakeLog = &upgraderfakes.FakeLogger{}
	})

	JustBeforeEach(func() {
		actionError = upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
			Action:     config.CheckDeactivatedPlansAction,
			BrokerName: fakeBrokerName,
		})
	})

	When("there are a deactivated plans with service instances", func() {
		var (
			deactivatedPlan1            ccapi.ServicePlan
			deactivatedPlan2            ccapi.ServicePlan
			deactivatedServiceInstance1 ccapi.ServiceInstance
			deactivatedServiceInstance2 ccapi.ServiceInstance
			deactivatedServiceInstance3 ccapi.ServiceInstance
		)

		BeforeEach(func() {
			deactivatedPlan1 = fakePlan(ccapi.ServicePlan{Available: false})
			deactivatedPlan2 = fakePlan(ccapi.ServicePlan{Available: false})
			deactivatedServiceInstance1 = fakeServiceInstance(ccapi.ServiceInstance{
				ServicePlanGUID:        deactivatedPlan1.GUID,
				ServicePlanName:        deactivatedPlan1.Name,
				ServicePlanDeactivated: true,
			})
			deactivatedServiceInstance2 = fakeServiceInstance(ccapi.ServiceInstance{
				ServicePlanGUID:        deactivatedPlan2.GUID,
				ServicePlanName:        deactivatedPlan2.Name,
				ServicePlanDeactivated: true,
			})
			deactivatedServiceInstance3 = fakeServiceInstance(ccapi.ServiceInstance{
				ServicePlanGUID:        deactivatedPlan2.GUID,
				ServicePlanName:        deactivatedPlan2.Name,
				ServicePlanDeactivated: true,
			})

			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
				fakePlan(ccapi.ServicePlan{Available: true}),
				deactivatedPlan1,
				fakePlan(ccapi.ServicePlan{Available: true}),
				deactivatedPlan2,
			}, nil)

			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{
				fakeServiceInstance(ccapi.ServiceInstance{}),
				deactivatedServiceInstance1,
				fakeServiceInstance(ccapi.ServiceInstance{}),
				deactivatedServiceInstance2,
				deactivatedServiceInstance3,
			}, nil)
		})

		It("detects the service instances associated with deactivated plans", func() {
			Expect(actionError).To(MatchError("discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services"))

			Expect(fakeLog.DeactivatedPlanCallCount()).To(Equal(3))
			Expect(fakeLog.DeactivatedPlanArgsForCall(0)).To(Equal(deactivatedServiceInstance1))
			Expect(fakeLog.DeactivatedPlanArgsForCall(1)).To(Equal(deactivatedServiceInstance2))
			Expect(fakeLog.DeactivatedPlanArgsForCall(2)).To(Equal(deactivatedServiceInstance3))
		})
	})

	When("there are deactivated plans with no service instances", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
				fakePlan(ccapi.ServicePlan{Available: true}),
				fakePlan(ccapi.ServicePlan{Available: false}),
				fakePlan(ccapi.ServicePlan{Available: true}),
			}, nil)

			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{
				fakeServiceInstance(ccapi.ServiceInstance{}),
				fakeServiceInstance(ccapi.ServiceInstance{}),
				fakeServiceInstance(ccapi.ServiceInstance{}),
			}, nil)
		})

		It("does not report an issue", func() {
			Expect(actionError).NotTo(HaveOccurred())
			Expect(fakeLog.DeactivatedPlanCallCount()).To(Equal(0))
		})
	})

	When("there are no deactivated plans", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
				fakePlan(ccapi.ServicePlan{Available: true}),
				fakePlan(ccapi.ServicePlan{Available: true}),
				fakePlan(ccapi.ServicePlan{Available: true}),
			}, nil)

			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{
				fakeServiceInstance(ccapi.ServiceInstance{}),
				fakeServiceInstance(ccapi.ServiceInstance{}),
				fakeServiceInstance(ccapi.ServiceInstance{}),
			}, nil)
		})

		It("does not report an issue", func() {
			Expect(actionError).NotTo(HaveOccurred())
			Expect(fakeLog.DeactivatedPlanCallCount()).To(Equal(0))
		})
	})

	When("no service plans are available", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{}, nil)
		})

		It("returns error stating no plans available", func() {
			Expect(actionError).To(MatchError(fmt.Sprintf("no service plans available for broker: %s", fakeBrokerName)))
		})
	})

	When("getting service plans fails", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns(nil, fmt.Errorf("plan-error"))
		})

		It("returns the error", func() {
			Expect(actionError).To(MatchError("plan-error"))
		})
	})

	When("getting service instances fails", func() {
		BeforeEach(func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{fakePlan(ccapi.ServicePlan{})}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns(nil, fmt.Errorf("instance-error"))
		})

		It("returns the error", func() {
			Expect(actionError).To(MatchError("instance-error"))
		})
	})
})
