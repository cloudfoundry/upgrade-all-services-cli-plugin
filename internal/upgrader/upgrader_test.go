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
		notUpToDateInstance1          ccapi.ServiceInstance
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
		notUpToDateInstance1 = ccapi.ServiceInstance{
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
		fakeServiceInstances = []ccapi.ServiceInstance{notUpToDateInstance1, fakeInstance2, fakeInstanceNoUpgrade, fakeInstanceCreateFailed, fakeInstanceDestroyFailed}
		fakeServiceInstancesNoUpgrade = []ccapi.ServiceInstance{fakeInstanceNoUpgrade, fakeInstanceCreateFailed}
		fakeCFClient = &upgraderfakes.FakeCFClient{}
		fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{fakePlan}, nil)
		fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstances, nil)

		fakeLog = &upgraderfakes.FakeLogger{}
		fakeLog.HasUpgradeSucceededReturns(true)
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
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, notUpToDateInstance1.GUID)))
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, fakeInstance2.GUID)))
		})
	})

	When("running with --check-up-to-date", func() {
		It("should print out service GUIDs and not attempt to upgrade", func() {
			result := captureStdout(func() {
				l := logger.New(100 * time.Millisecond)
				defer l.Cleanup()
				err := upgrader.Upgrade(fakeCFClient, l, upgrader.UpgradeConfig{
					BrokerName:       fakeBrokerName,
					ParallelUpgrades: 5,
					CheckUpToDate:    true,
				})
				Expect(err).To(MatchError("found 3 instances which are not up-to-date"))
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
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, notUpToDateInstance1.GUID)))
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, fakeInstance2.GUID)))
		})

		When("there are deactivated plans", func() {
			Context("because we force the deactivated plans check in the check-up-to-date operation", func() {
				It("returns error stating there are deactivated plans", func() {
					deactivatedPlan := ccapi.ServicePlan{
						GUID:                   "fake-deactivated-plan-1-guid",
						Available:              false,
						Name:                   "fake-deactivated-plan-1-name",
						MaintenanceInfoVersion: "1.5.7",
						ServiceOfferingGUID:    "fake-service-offering-for-plan-1-guid",
						ServiceOfferingName:    "fake-service-offering-for-plan-1-name",
					}
					activePlan := ccapi.ServicePlan{
						GUID:                   "fake-plan-2-guid",
						Available:              true,
						Name:                   "fake-plan-2-name",
						MaintenanceInfoVersion: "1.5.7",
						ServiceOfferingGUID:    "fake-service-offering-for-plan-2-guid",
						ServiceOfferingName:    "fake-service-offering-for-plan-2-name",
					}
					plans := []ccapi.ServicePlan{deactivatedPlan, activePlan}
					fakeCFClient.GetServicePlansReturns(plans, nil)

					instanceWithDeactivatedPlan := ccapi.ServiceInstance{
						GUID:                              "fake-instance-guid-1",
						Name:                              "fake-instance-name-1",
						ServicePlanGUID:                   "fake-deactivated-plan-1-guid",
						ServicePlanName:                   "fake-deactivated-plan-1-name",
						ServiceOfferingGUID:               "fake-service-offering-for-plan-1-guid",
						ServiceOfferingName:               "fake-service-offering-for-plan-1-name",
						ServicePlanDeactivated:            true,  // ccapi.ServicePlan.Available = false
						UpgradeAvailable:                  false, // ServicePlanMaintenanceInfoVersion == MaintenanceInfoVersion
						ServicePlanMaintenanceInfoVersion: "1.5.7",
						MaintenanceInfoVersion:            "1.5.7",
					}

					fakeInstance2 = ccapi.ServiceInstance{
						GUID:                              "fake-instance-guid-2",
						Name:                              "fake-instance-name-2",
						UpgradeAvailable:                  false,
						ServicePlanGUID:                   "fake-plan-2-guid",
						ServicePlanName:                   "fake-plan-2-name",
						ServiceOfferingGUID:               "fake-service-offering-for-plan-2-guid",
						ServiceOfferingName:               "fake-service-offering-for-plan-2-name",
						ServicePlanMaintenanceInfoVersion: "1.5.7",
						ServicePlanDeactivated:            false,
						MaintenanceInfoVersion:            "1.5.7",
					}

					notUpToDateInstance := ccapi.ServiceInstance{
						GUID:                              "fake-instance-guid-3",
						Name:                              "fake-instance-name-3",
						ServicePlanGUID:                   "fake-plan-2-guid",
						ServicePlanName:                   "fake-plan-2-name",
						ServiceOfferingGUID:               "fake-service-offering-for-plan-2-guid",
						ServiceOfferingName:               "fake-service-offering-for-plan-2-name",
						ServicePlanDeactivated:            false,
						UpgradeAvailable:                  true, // MaintenanceInfoVersion != ServicePlanMaintenanceInfoVersion
						ServicePlanMaintenanceInfoVersion: "1.5.7",
						MaintenanceInfoVersion:            "1.4.0",
					}

					fakeServiceInstances = []ccapi.ServiceInstance{instanceWithDeactivatedPlan, fakeInstance2, notUpToDateInstance}
					fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstances, nil)

					err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
						BrokerName:       fakeBrokerName,
						ParallelUpgrades: 1,
						CheckUpToDate:    true,
					})
					Expect(err).To(MatchError(ContainSubstring("found 1 instances which are not up-to-date")))
					Expect(err).To(MatchError(ContainSubstring("discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services")))
				})
			})
		})
	})

	When("running with --check-version-less-than-min-required", func() {

		BeforeEach(func() {
			plan := ccapi.ServicePlan{GUID: "fake-plan-1-guid"}
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{plan}, nil)
		})

		It("should return an error with the instances that are not up-to-date", func() {
			fakeInstance1 := ccapi.ServiceInstance{
				GUID:                   "fake-instance-guid-1",
				MaintenanceInfoVersion: "1.4.0",
			}

			fakeInstance2 = ccapi.ServiceInstance{
				GUID:                   "fake-instance-guid-2",
				MaintenanceInfoVersion: "1.3.0",
			}

			fakeInstance3 := ccapi.ServiceInstance{
				GUID:                   "fake-instance-guid-3",
				MaintenanceInfoVersion: "1.5.7",
			}
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{{GUID: "fake-plan-1-guid"}}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{fakeInstance1, fakeInstance2, fakeInstance3}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:         fakeBrokerName,
				ParallelUpgrades:   5,
				MinVersionRequired: "1.5.7",
			})

			Expect(err).To(MatchError("found 2 service instances with a version less than the minimum required"))
			Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(2))
			_, _, dryrunErr := fakeLog.UpgradeFailedArgsForCall(0)
			Expect(dryrunErr.Error()).To(Equal(fmt.Sprintf("dry-run prevented upgrade instance guid %s", fakeInstance1.GUID)))
			_, _, dryrunErr = fakeLog.UpgradeFailedArgsForCall(1)
			Expect(dryrunErr.Error()).To(Equal(fmt.Sprintf("dry-run prevented upgrade instance guid %s", fakeInstance2.GUID)))
		})

		It("should return an error because the instance has a malformed version", func() {
			fakeInstance1 := ccapi.ServiceInstance{
				GUID:                   "fake-instance-guid-1",
				MaintenanceInfoVersion: "invalid version",
			}
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{{GUID: "fake-plan-1-guid"}}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{fakeInstance1}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:         fakeBrokerName,
				ParallelUpgrades:   5,
				MinVersionRequired: "1.5.7",
			})

			Expect(err).To(MatchError("incorrect instance version: Malformed version: invalid version"))
			Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(0))
		})

		It("should return an error because the flag version is a malformed version", func() {
			fakeInstance1 := ccapi.ServiceInstance{
				GUID:                   "fake-instance-guid-1",
				MaintenanceInfoVersion: "1.4.7",
			}
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{{GUID: "fake-plan-1-guid"}}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{fakeInstance1}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:         fakeBrokerName,
				ParallelUpgrades:   5,
				MinVersionRequired: "invalid version",
			})

			Expect(err).To(MatchError("incorrect minimum required version: Malformed version: invalid version"))
			Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(0))
		})

		It("should not return an error because any instance has a version less than specified", func() {
			fakeInstance1 := ccapi.ServiceInstance{MaintenanceInfoVersion: "1.4.0"}
			fakeInstance2 = ccapi.ServiceInstance{MaintenanceInfoVersion: "1.3.0"}
			fakeInstance3 := ccapi.ServiceInstance{MaintenanceInfoVersion: "1.5.7"}
			fakeCFClient.GetServicePlansReturns([]ccapi.ServicePlan{{GUID: "fake-plan-1-guid"}}, nil)
			fakeCFClient.GetServiceInstancesForServicePlansReturns([]ccapi.ServiceInstance{fakeInstance1, fakeInstance2, fakeInstance3}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:         fakeBrokerName,
				ParallelUpgrades:   5,
				MinVersionRequired: "1.3.0",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(0))
			Expect(fakeLog.PrintfArgsForCall(1)).To(Equal("no instances found with version less than required"))
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

				notUpToDateInstance1 = ccapi.ServiceInstance{
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
				fakeServiceInstances = []ccapi.ServiceInstance{notUpToDateInstance1, fakeInstance2}
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
						"discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services",
					),
				)

				Expect(fakeLog.DeactivatedPlanCallCount()).To(Equal(1))
				Expect(fakeLog.DeactivatedPlanArgsForCall(0)).To(Equal(notUpToDateInstance1))
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

				notUpToDateInstance1 = ccapi.ServiceInstance{
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
				fakeServiceInstances = []ccapi.ServiceInstance{notUpToDateInstance1, fakeInstance2}
				fakeCFClient.GetServiceInstancesForServicePlansReturns(fakeServiceInstances, nil)
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
			fakeLog.HasUpgradeSucceededReturns(false)
		})

		It("should pass the correct information to the logger", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:       fakeBrokerName,
				ParallelUpgrades: 1,
			})
			Expect(err).To(HaveOccurred())

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

		It("should return an error", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeLog, upgrader.UpgradeConfig{
				BrokerName:            fakeBrokerName,
				ParallelUpgrades:      1,
				DryRun:                false,
				CheckDeactivatedPlans: false,
			})

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("there were failures upgrading one or more instances. Review the logs for more information"))

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
