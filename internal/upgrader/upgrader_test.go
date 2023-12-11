package upgrader_test

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/logger"
	"upgrade-all-services-cli-plugin/internal/upgrader"
	"upgrade-all-services-cli-plugin/internal/upgrader/upgraderfakes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Upgrade", func() {
	const (
		fakePlanGUID   = "test-plan-guid"
		fakeBrokerName = "fake-broker-name"
	)

	var (
		fakeCFClient                  *upgraderfakes.FakeCFClient
		fakePlan                      ccapi.ServicePlan
		fakeInstance1                 ccapi.ServiceInstance
		fakeInstance2                 ccapi.ServiceInstance
		fakeInstanceNoUpgrade         ccapi.ServiceInstance
		fakeInstanceCreateFailed      ccapi.ServiceInstance
		fakeInstanceDestroyFailed     ccapi.ServiceInstance
		fakeServiceInstances          []ccapi.ServiceInstance
		fakeServiceInstancesNoUpgrade []ccapi.ServiceInstance
		fakeLog                       *upgraderfakes.FakeLogger
	)

	BeforeEach(func() {
		fakePlan = ccapi.ServicePlan{
			GUID:                   fakePlanGUID,
			MaintenanceInfoVersion: "test-maintenance-info",
		}
		fakeInstance1 = ccapi.ServiceInstance{
			Name:             "fake-instance-name-1",
			GUID:             "fake-instance-guid-1",
			ServicePlanGUID:  fakePlanGUID,
			UpgradeAvailable: true,
		}
		fakeInstance2 = ccapi.ServiceInstance{
			Name:             "fake-instance-name-2",
			GUID:             "fake-instance-guid-2",
			ServicePlanGUID:  fakePlanGUID,
			UpgradeAvailable: true,
		}
		fakeInstanceNoUpgrade = ccapi.ServiceInstance{
			GUID:             "fake-instance-no-upgrade-GUID",
			ServicePlanGUID:  fakePlanGUID,
			UpgradeAvailable: false,
		}
		fakeInstanceCreateFailed = ccapi.ServiceInstance{
			Name:             "fake-instance-create-failed",
			GUID:             "fake-instance-create-failed-GUID",
			ServicePlanGUID:  fakePlanGUID,
			UpgradeAvailable: true,
		}
		fakeInstanceDestroyFailed = ccapi.ServiceInstance{
			Name:             "fake-instance-destroy-failed",
			GUID:             "fake-instance-destroy-failed-GUID",
			ServicePlanGUID:  fakePlanGUID,
			UpgradeAvailable: true,
		}

		fakeInstanceCreateFailed.LastOperationType = "create"
		fakeInstanceCreateFailed.LastOperationState = "failed"
		fakeInstanceDestroyFailed.LastOperationType = "destroy"
		fakeInstanceDestroyFailed.LastOperationState = "failed"
		fakeServiceInstances = []ccapi.ServiceInstance{fakeInstance1, fakeInstance2, fakeInstanceNoUpgrade, fakeInstanceCreateFailed, fakeInstanceDestroyFailed}
		fakeServiceInstancesNoUpgrade = []ccapi.ServiceInstance{fakeInstanceNoUpgrade, fakeInstanceCreateFailed}
		fakeCFClient = &upgraderfakes.FakeCFClient{}
		fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{fakePlan}, nil)
		fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstances, nil)

		fakeLog = &upgraderfakes.FakeLogger{}
	})

	It("upgrades a service instance", func() {
		err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
			BrokerName:       fakeBrokerName,
			ParallelUpgrades: 5,
		})
		Expect(err).NotTo(HaveOccurred())

		By("getting the service plans")
		Expect(fakeCFClient.GetServicePlansCallCount()).To(Equal(1))
		Expect(fakeCFClient.GetServicePlansArgsForCall(0)).To(Equal(fakeBrokerName))

		By("getting the service instances")
		Expect(fakeCFClient.GetServiceInstancesForServicePlansCallCount()).To(Equal(1))
		Expect(fakeCFClient.GetServiceInstancesForServicePlansArgsForCall(0)).To(Equal([]ccapi.ServicePlan{
			{
				GUID:                   "test-plan-guid",
				MaintenanceInfoVersion: "test-maintenance-info",
			},
		}))

		By("calling upgrade on each upgradeable instance")
		Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(3))
		instanceGUID1, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(0)
		instanceGUID2, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(1)
		instanceGUID3, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(2)
		guids := []string{instanceGUID1, instanceGUID2, instanceGUID3}
		Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2", "fake-instance-destroy-failed-GUID"))
	})

	It("should pass the correct information to the logger", func() {
		err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
			BrokerName:       fakeBrokerName,
			ParallelUpgrades: 1,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeLog.InitialTotalsCallCount()).To(Equal(1))
		actualTotal, actualUpgradable := fakeLog.InitialTotalsArgsForCall(0)
		Expect(actualTotal).To(Equal(5))
		Expect(actualUpgradable).To(Equal(3))

		Expect(fakeLog.SkippingInstanceCallCount()).To(Equal(1))
		instanceSkipped := fakeLog.SkippingInstanceArgsForCall(0)
		Expect(instanceSkipped.Name).To(Equal("fake-instance-create-failed"))
		Expect(instanceSkipped.GUID).To(Equal("fake-instance-create-failed-GUID"))
		Expect(instanceSkipped.UpgradeAvailable).To(BeTrue())
		Expect(instanceSkipped.LastOperationType).To(Equal("create"))
		Expect(instanceSkipped.LastOperationState).To(Equal("failed"))

		Expect(fakeLog.UpgradeStartingCallCount()).To(Equal(3))
		instance1 := fakeLog.UpgradeStartingArgsForCall(0)
		Expect(instance1.Name).To(Equal("fake-instance-name-1"))
		Expect(instance1.GUID).To(Equal("fake-instance-guid-1"))
		instance2 := fakeLog.UpgradeStartingArgsForCall(1)
		Expect(instance2.Name).To(Equal("fake-instance-name-2"))
		Expect(instance2.GUID).To(Equal("fake-instance-guid-2"))
		instance3 := fakeLog.UpgradeStartingArgsForCall(2)
		Expect(instance3.Name).To(Equal("fake-instance-destroy-failed"))
		Expect(instance3.GUID).To(Equal("fake-instance-destroy-failed-GUID"))

		Expect(fakeLog.UpgradeSucceededCallCount()).To(Equal(3))
		Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(0))
		Expect(fakeLog.FinalTotalsCallCount()).To(Equal(1))
	})

	When("running with --dry-run", func() {
		It("should print out service GUIDs and not attempt to upgrade", func() {
			result := captureStdout(func() {
				l := logger.New(100 * time.Millisecond)
				defer l.Cleanup()
				err := upgrader.Upgrade(fakeCFClient, l, upgrader.UpgradeConfig{
					BrokerName:       fakeBrokerName,
					ParallelUpgrades: 5,
					DryRun:           true,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			By("getting the service plans")
			Expect(fakeCFClient.GetServicePlansCallCount()).To(Equal(1))
			Expect(fakeCFClient.GetServicePlansArgsForCall(0)).To(Equal(fakeBrokerName))

			By("getting the service instances")
			Expect(fakeCFClient.GetServiceInstancesForServicePlansCallCount()).To(Equal(1))
			Expect(fakeCFClient.GetServiceInstancesForServicePlansArgsForCall(0)).To(Equal([]ccapi.ServicePlan{fakePlan}))

			By("not calling upgrade")
			Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(0))

			By("printing the GUIDs")
			Expect(result).To(ContainSubstring(fmt.Sprintf("discovering service instances for broker: %s", fakeBrokerName)))
			Expect(result).To(ContainSubstring("the following service instances would be upgraded:"))
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, fakeInstance1.GUID)))
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, fakeInstance2.GUID)))
		})
	})

	When("running with --check-up-to-date", func() {
		Context("there are instances not up to date", func() {
			var notUpToDateInstance, withPlanDeactivated ccapi.ServiceInstance
			BeforeEach(func() {
				fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
					{
						GUID:                   "fake-plan-1-guid",
						Available:              true,
						Name:                   "fake-plan-1-name",
						MaintenanceInfoVersion: "1.5.7",
					},
					{
						GUID:                   "fake-plan-2-guid",
						Available:              true,
						Name:                   "fake-plan-2-name",
						MaintenanceInfoVersion: "1.5.7",
					},
				}, nil)

				fakeInstance1 = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-1",
					Name:                              "fake-instance-name-1",
					UpgradeAvailable:                  false, // no upgradable, skipped instance
					MaintenanceInfoVersion:            "1.3.0",
					ServicePlanGUID:                   "fake-plan-1-guid",
					ServicePlanMaintenanceInfoVersion: "1.5.7",
					ServicePlanDeactivated:            false,
				}

				fakeInstance2 = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-2",
					Name:                              "fake-instance-name-2",
					UpgradeAvailable:                  true,
					MaintenanceInfoVersion:            "1.5.7",
					ServicePlanGUID:                   "fake-plan-1-guid",
					ServicePlanMaintenanceInfoVersion: "1.5.7",
					ServicePlanDeactivated:            false,
				}

				// Plan and instance with different version
				notUpToDateInstance = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-3",
					UpgradeAvailable:                  true,
					MaintenanceInfoVersion:            "1.5.7",
					ServicePlanGUID:                   "fake-plan-2-guid",
					ServicePlanMaintenanceInfoVersion: "1.5.0",
					ServicePlanDeactivated:            false,
				}

				withPlanDeactivated = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-4",
					Name:                              "fake-instance-name-4",
					UpgradeAvailable:                  true,
					MaintenanceInfoVersion:            "1.3.0",
					ServicePlanGUID:                   "fake-plan-1-guid",
					ServicePlanMaintenanceInfoVersion: "1.5.7",
					ServicePlanDeactivated:            true, // plan deactivated, skipped instance
				}

				fakeServiceInstances = []ccapi.ServiceInstance{fakeInstance1, fakeInstance2, notUpToDateInstance, withPlanDeactivated}
				fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstances, nil)
			})

			It("should fail and not attempt to upgrade", func() {
				err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
					BrokerName:       fakeBrokerName,
					ParallelUpgrades: 5,
					CheckUpToDate:    true,
				})

				Expect(err).To(
					MatchError(
						"discovered upgradable instances that are not up to date. Review the log to collect information and update them",
					),
				)

				Expect(fakeLog.InstanceIsNotUpToDateCallCount()).To(Equal(1))
				Expect(fakeLog.InstanceIsNotUpToDateArgsForCall(0)).To(Equal(notUpToDateInstance))
				Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(0))
			})
		})

		When("no service instances have pending upgrades", func() {
			It("does not return an error", func() {
				fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{}, nil)

				err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
					BrokerName:       fakeBrokerName,
					ParallelUpgrades: 5,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("all instances are up to date", func() {
			It("does not return an error", func() {
				fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstancesNoUpgrade, nil)

				err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
					BrokerName:       fakeBrokerName,
					ParallelUpgrades: 5,
				})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	When("running with --check-deactivated-plans", func() {
		When("there is a deactivated plan", func() {
			BeforeEach(func() {
				fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
					fakePlan,
					{
						GUID:                   "fake-deactivated-plan-1-guid",
						Available:              false,
						Name:                   "fake-deactivated-plan-1-name",
						MaintenanceInfoVersion: "1.5.7",
						ServiceOfferingGUID:    "fake-service-offering-for-plan-1-guid",
						ServiceOfferingName:    "fake-service-offering-for-plan-1-name",
					},
					{
						GUID:                   "fake-deactivated-plan-2-guid",
						Available:              true,
						Name:                   "fake-deactivated-plan-2-name",
						MaintenanceInfoVersion: "1.5.7",
						ServiceOfferingGUID:    "fake-service-offering-for-plan-2-guid",
						ServiceOfferingName:    "fake-service-offering-for-plan-2-name",
					},
				}, nil)

				fakeInstance1 = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-1",
					Name:                              "fake-instance-name-1",
					UpgradeAvailable:                  true,
					ServicePlanGUID:                   "fake-deactivated-plan-1-guid",
					ServicePlanName:                   "fake-deactivated-plan-1-name",
					ServiceOfferingGUID:               "fake-service-offering-for-plan-1-guid",
					ServiceOfferingName:               "fake-service-offering-for-plan-1-name",
					ServicePlanMaintenanceInfoVersion: "1.5.7",
					ServicePlanDeactivated:            true,
				}
				fakeInstance2 = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-2",
					Name:                              "fake-instance-name-2",
					UpgradeAvailable:                  true,
					ServicePlanGUID:                   "fake-deactivated-plan-2-guid",
					ServicePlanName:                   "fake-deactivated-plan-2-name",
					ServiceOfferingGUID:               "fake-service-offering-for-plan-2-guid",
					ServiceOfferingName:               "fake-service-offering-for-plan-2-name",
					ServicePlanMaintenanceInfoVersion: "1.5.7",
					ServicePlanDeactivated:            false,
				}
				fakeServiceInstances = []ccapi.ServiceInstance{fakeInstance1, fakeInstance2}
				fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstances, nil)
			})

			It("returns an error", func() {

				err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
					BrokerName:            fakeBrokerName,
					ParallelUpgrades:      5,
					CheckDeactivatedPlans: true,
				})

				Expect(err).To(
					MatchError(
						"discovered deactivated plans associated with upgradable instances. Review the log to collect information and restore the deactivated plans or create user provided services",
					),
				)

				Expect(fakeLog.DeactivatedPlanCallCount()).To(Equal(1))
				Expect(fakeLog.DeactivatedPlanArgsForCall(0)).To(Equal(fakeInstance1))
			})
		})

		When("there are no deactivated plans", func() {
			BeforeEach(func() {
				fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{
					fakePlan,
					{
						GUID:                   "fake-deactivated-plan-1-guid",
						Available:              true,
						Name:                   "fake-deactivated-plan-1-name",
						MaintenanceInfoVersion: "1.5.7",
						ServiceOfferingGUID:    "fake-service-offering-for-plan-1-guid",
						ServiceOfferingName:    "fake-service-offering-for-plan-1-name",
					},
					{
						GUID:                   "fake-deactivated-plan-2-guid",
						Available:              true,
						Name:                   "fake-deactivated-plan-2-name",
						MaintenanceInfoVersion: "1.5.7",
						ServiceOfferingGUID:    "fake-service-offering-for-plan-2-guid",
						ServiceOfferingName:    "fake-service-offering-for-plan-2-name",
					},
				}, nil)

				fakeInstance1 = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-1",
					Name:                              "fake-instance-name-1",
					UpgradeAvailable:                  false,
					ServicePlanGUID:                   "fake-deactivated-plan-1-guid",
					ServicePlanName:                   "fake-deactivated-plan-1-name",
					ServiceOfferingGUID:               "fake-service-offering-for-plan-1-guid",
					ServiceOfferingName:               "fake-service-offering-for-plan-1-name",
					ServicePlanMaintenanceInfoVersion: "1.5.7",
					ServicePlanDeactivated:            false,
				}
				fakeInstance2 = ccapi.ServiceInstance{
					GUID:                              "fake-instance-guid-2",
					Name:                              "fake-instance-name-2",
					UpgradeAvailable:                  true,
					ServicePlanGUID:                   "fake-deactivated-plan-2-guid",
					ServicePlanName:                   "fake-deactivated-plan-2-name",
					ServiceOfferingGUID:               "fake-service-offering-for-plan-2-guid",
					ServiceOfferingName:               "fake-service-offering-for-plan-2-name",
					ServicePlanMaintenanceInfoVersion: "1.5.7",
					ServicePlanDeactivated:            false,
				}
				fakeServiceInstances = []ccapi.ServiceInstance{fakeInstance1, fakeInstance2}
				fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstances, nil)
			})

			Context("all instances are up to date", func() {
				It("does not return an error", func() {
					err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
						BrokerName:            fakeBrokerName,
						ParallelUpgrades:      5,
						CheckDeactivatedPlans: true,
					})

					Expect(err).To(BeNil())
					Expect(fakeLog.DeactivatedPlanCallCount()).To(Equal(0))
				})
			})
		})
	})

	When("no service plans are available", func() {
		It("returns error stating no plans available", func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:       fakeBrokerName,
				ParallelUpgrades: 5,
			})
			Expect(err).To(MatchError(fmt.Sprintf("no service plans available for broker: %s", fakeBrokerName)))
		})
	})

	When("number of parallel upgrades is less that number of upgradable instances", func() {
		It("upgrades all instances", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:       fakeBrokerName,
				ParallelUpgrades: 1,
			})
			Expect(err).NotTo(HaveOccurred())

			By("calling upgrade on each upgradeable instance")
			Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(3))
			instanceGUID1, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(0)
			instanceGUID2, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(1)
			instanceGUID3, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(2)
			guids := []string{instanceGUID1, instanceGUID2, instanceGUID3}
			Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2", "fake-instance-destroy-failed-GUID"))
		})
	})

	When("there are no upgradable instances", func() {
		It("should succeed and pass the correct information to the logger", func() {
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:       fakeBrokerName,
				ParallelUpgrades: 1,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLog.PrintfCallCount()).To(Equal(2))
			Expect(fakeLog.PrintfArgsForCall(1)).To(Equal(`no instances available to upgrade`))
		})
	})

	When("getting service plans fails", func() {
		It("returns the error", func() {
			fakeCFClient.GetServicePlansReturns(nil, fmt.Errorf("plan-error"))

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:       fakeBrokerName,
				ParallelUpgrades: 1,
			})
			Expect(err).To(MatchError("plan-error"))
		})
	})

	When("getting service instances fails", func() {
		It("returns the error", func() {
			fakeCFClient.GetServiceInstancesForServicePlansReturns(nil, fmt.Errorf("instance-error"))

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:       fakeBrokerName,
				ParallelUpgrades: 1,
			})
			Expect(err).To(MatchError("instance-error"))
		})
	})

	When("an instance fails to upgrade", func() {
		BeforeEach(func() {
			fakeCFClient.UpgradeServiceInstanceReturnsOnCall(0, nil)
			fakeCFClient.UpgradeServiceInstanceReturnsOnCall(1, fmt.Errorf("failed to upgrade instance"))
			fakeCFClient.UpgradeServiceInstanceReturnsOnCall(2, nil)
		})

		It("should pass the correct information to the logger", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:       fakeBrokerName,
				ParallelUpgrades: 1,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLog.InitialTotalsCallCount()).To(Equal(1))
			actualTotal, actualUpgradable := fakeLog.InitialTotalsArgsForCall(0)
			Expect(actualTotal).To(Equal(5))
			Expect(actualUpgradable).To(Equal(3))

			Expect(fakeLog.UpgradeStartingCallCount()).To(Equal(3))
			instance1 := fakeLog.UpgradeStartingArgsForCall(0)
			Expect(instance1.Name).To(Equal("fake-instance-name-1"))
			Expect(instance1.GUID).To(Equal("fake-instance-guid-1"))
			instance2 := fakeLog.UpgradeStartingArgsForCall(1)
			Expect(instance2.Name).To(Equal("fake-instance-name-2"))
			Expect(instance2.GUID).To(Equal("fake-instance-guid-2"))
			instance3 := fakeLog.UpgradeStartingArgsForCall(2)
			Expect(instance3.Name).To(Equal("fake-instance-destroy-failed"))
			Expect(instance3.GUID).To(Equal("fake-instance-destroy-failed-GUID"))

			Expect(fakeLog.UpgradeSucceededCallCount()).To(Equal(2))
			Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(1))
			Expect(fakeLog.FinalTotalsCallCount()).To(Equal(1))
		})
	})
})

var captureStdoutLock sync.Mutex

func captureStdout(callback func()) (result string) {
	captureStdoutLock.Lock()

	reader, writer, err := os.Pipe()
	Expect(err).NotTo(HaveOccurred())

	originalStdout := os.Stdout
	os.Stdout = writer

	defer func() {
		writer.Close()
		os.Stdout = originalStdout
		captureStdoutLock.Unlock()

		data, err := io.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		result = string(data)
	}()

	callback()
	return
}
