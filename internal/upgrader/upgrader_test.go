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
		fakeCFClient             *upgraderfakes.FakeCFClient
		fakePlan                 ccapi.Plan
		fakeInstance1            ccapi.ServiceInstance
		fakeInstance2            ccapi.ServiceInstance
		fakeInstanceNoUpgrade    ccapi.ServiceInstance
		fakeInstanceCreateFailed ccapi.ServiceInstance
		fakeServiceInstances     []ccapi.ServiceInstance
		fakeLog                  *upgraderfakes.FakeLogger
	)

	BeforeEach(func() {
		fakePlan = ccapi.Plan{
			GUID:                   fakePlanGUID,
			MaintenanceInfoVersion: "test-maintenance-info",
		}
		fakeInstance1 = ccapi.ServiceInstance{
			Name:             "fake-instance-name-1",
			GUID:             "fake-instance-guid-1",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: true,
		}
		fakeInstance2 = ccapi.ServiceInstance{
			Name:             "fake-instance-name-2",
			GUID:             "fake-instance-guid-2",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: true,
		}
		fakeInstanceNoUpgrade = ccapi.ServiceInstance{
			GUID:             "fake-instance-no-upgrade-GUID",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: false,
		}
		fakeInstanceCreateFailed = ccapi.ServiceInstance{
			Name:             "fake-instance-create-failed",
			GUID:             "fake-instance-create-failed-GUID",
			PlanGUID:         fakePlanGUID,
			UpgradeAvailable: true,
		}

		fakeInstanceCreateFailed.LastOperation.Type = "create"
		fakeInstanceCreateFailed.LastOperation.State = "failed"
		fakeServiceInstances = []ccapi.ServiceInstance{fakeInstance1, fakeInstance2, fakeInstanceNoUpgrade, fakeInstanceCreateFailed}

		fakeCFClient = &upgraderfakes.FakeCFClient{}
		fakeCFClient.GetServicePlansReturns([]ccapi.Plan{fakePlan}, nil)
		fakeCFClient.GetServiceInstancesReturns(fakeServiceInstances, nil)

		fakeLog = &upgraderfakes.FakeLogger{}
	})

	It("upgrades a service instance", func() {
		err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, false, fakeLog)
		Expect(err).NotTo(HaveOccurred())

		By("getting the service plans")
		Expect(fakeCFClient.GetServicePlansCallCount()).To(Equal(1))
		Expect(fakeCFClient.GetServicePlansArgsForCall(0)).To(Equal(fakeBrokerName))

		By("getting the service instances")
		Expect(fakeCFClient.GetServiceInstancesCallCount()).To(Equal(1))
		Expect(fakeCFClient.GetServiceInstancesArgsForCall(0)).To(Equal([]string{fakePlanGUID}))

		By("calling upgrade on each upgradeable instance")
		Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(2))
		instanceGUID1, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(0)
		instanceGUID2, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(1)
		guids := []string{instanceGUID1, instanceGUID2}
		Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2"))
	})

	It("should pass the correct information to the logger", func() {
		err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, false, fakeLog)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeLog.InitialTotalsCallCount()).To(Equal(1))
		actualTotal, actualUpgradable := fakeLog.InitialTotalsArgsForCall(0)
		Expect(actualTotal).To(Equal(4))
		Expect(actualUpgradable).To(Equal(2))

		Expect(fakeLog.SkippingInstanceCallCount()).To(Equal(1))
		instanceSkipped := fakeLog.SkippingInstanceArgsForCall(0)
		Expect(instanceSkipped.Name).To(Equal("fake-instance-create-failed"))
		Expect(instanceSkipped.GUID).To(Equal("fake-instance-create-failed-GUID"))
		Expect(instanceSkipped.UpgradeAvailable).To(BeTrue())
		Expect(instanceSkipped.LastOperation.Type).To(Equal("create"))
		Expect(instanceSkipped.LastOperation.State).To(Equal("failed"))

		Expect(fakeLog.UpgradeStartingCallCount()).To(Equal(2))
		instance1 := fakeLog.UpgradeStartingArgsForCall(0)
		Expect(instance1.Name).To(Equal("fake-instance-name-1"))
		Expect(instance1.GUID).To(Equal("fake-instance-guid-1"))
		instance2 := fakeLog.UpgradeStartingArgsForCall(1)
		Expect(instance2.Name).To(Equal("fake-instance-name-2"))
		Expect(instance2.GUID).To(Equal("fake-instance-guid-2"))

		Expect(fakeLog.UpgradeSucceededCallCount()).To(Equal(2))
		Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(0))
		Expect(fakeLog.FinalTotalsCallCount()).To(Equal(1))
	})

	When("performing a dry run", func() {
		It("should print out service GUIDs and not attempt to upgrade", func() {
			result := captureStdout(func() {
				l := logger.New(100 * time.Millisecond)
				defer l.Cleanup()
				err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, true, l)
				Expect(err).NotTo(HaveOccurred())
			})

			By("getting the service plans")
			Expect(fakeCFClient.GetServicePlansCallCount()).To(Equal(1))
			Expect(fakeCFClient.GetServicePlansArgsForCall(0)).To(Equal(fakeBrokerName))

			By("getting the service instances")
			Expect(fakeCFClient.GetServiceInstancesCallCount()).To(Equal(1))
			Expect(fakeCFClient.GetServiceInstancesArgsForCall(0)).To(Equal([]string{fakePlanGUID}))

			By("not calling upgrade")
			Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(0))

			By("printing the GUIDs")
			Expect(result).To(ContainSubstring(fmt.Sprintf("discovering service instances for broker: %s", fakeBrokerName)))
			Expect(result).To(ContainSubstring("the following service instances would be upgraded:"))
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, fakeInstance1.GUID)))
			Expect(result).To(ContainSubstring(fmt.Sprintf(`Service Instance GUID: "%s"`, fakeInstance2.GUID)))
		})
	})

	When("no service plans are available", func() {
		It("returns error stating no plans available", func() {
			fakeCFClient.GetServicePlansReturns([]ccapi.Plan{}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, false, fakeLog)
			Expect(err).To(MatchError(fmt.Sprintf("no service plans available for broker: %s", fakeBrokerName)))
		})
	})

	When("number of parallel upgrades is less that number of upgradable instances", func() {
		It("upgrades all instances", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, false, fakeLog)
			Expect(err).NotTo(HaveOccurred())

			By("calling upgrade on each upgradeable instance")
			Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(2))
			instanceGUID1, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(0)
			instanceGUID2, _ := fakeCFClient.UpgradeServiceInstanceArgsForCall(1)
			guids := []string{instanceGUID1, instanceGUID2}
			Expect(guids).To(ConsistOf("fake-instance-guid-1", "fake-instance-guid-2"))
		})
	})

	When("there are no upgradable instances", func() {
		It("should succeed and pass the correct information to the logger", func() {
			fakeCFClient.GetServiceInstancesReturns([]ccapi.ServiceInstance{}, nil)

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, false, fakeLog)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLog.PrintfCallCount()).To(Equal(2))
			Expect(fakeLog.PrintfArgsForCall(1)).To(Equal(`no instances available to upgrade`))
		})
	})

	When("getting service plans fails", func() {
		It("returns the error", func() {
			fakeCFClient.GetServicePlansReturns(nil, fmt.Errorf("plan-error"))

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, false, fakeLog)
			Expect(err).To(MatchError("plan-error"))
		})
	})

	When("getting service instances fails", func() {
		It("returns the error", func() {
			fakeCFClient.GetServiceInstancesReturns(nil, fmt.Errorf("instance-error"))

			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 5, false, fakeLog)
			Expect(err).To(MatchError("instance-error"))
		})
	})

	When("an instance fails to upgrade", func() {
		BeforeEach(func() {
			fakeCFClient.UpgradeServiceInstanceReturnsOnCall(0, nil)
			fakeCFClient.UpgradeServiceInstanceReturnsOnCall(1, fmt.Errorf("failed to upgrade instance"))
		})

		It("should pass the correct information to the logger", func() {
			err := upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, false, fakeLog)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeLog.InitialTotalsCallCount()).To(Equal(1))
			actualTotal, actualUpgradable := fakeLog.InitialTotalsArgsForCall(0)
			Expect(actualTotal).To(Equal(4))
			Expect(actualUpgradable).To(Equal(2))

			Expect(fakeLog.UpgradeStartingCallCount()).To(Equal(2))
			instance1 := fakeLog.UpgradeStartingArgsForCall(0)
			Expect(instance1.Name).To(Equal("fake-instance-name-1"))
			Expect(instance1.GUID).To(Equal("fake-instance-guid-1"))
			instance2 := fakeLog.UpgradeStartingArgsForCall(1)
			Expect(instance2.Name).To(Equal("fake-instance-name-2"))
			Expect(instance2.GUID).To(Equal("fake-instance-guid-2"))

			Expect(fakeLog.UpgradeSucceededCallCount()).To(Equal(1))
			Expect(fakeLog.UpgradeFailedCallCount()).To(Equal(1))
			Expect(fakeLog.FinalTotalsCallCount()).To(Equal(1))
		})
	})

	Context("checkUpToDate: true", func() {
		DescribeTable("expected behaviour of Upgrade when CheckUpToDate is enabled",
			func(expectedErr error, expectedLog string, fakePlans []ccapi.Plan, fakeInstances []ccapi.ServiceInstance) {
				passedDryRun := true
				fakeCFClient.GetServicePlansReturns(fakePlans, nil)
				fakeCFClient.GetServiceInstancesReturns(fakeInstances, nil)

				var upgradeErr error
				upgradeLog := captureStdout(func() {
					l := logger.New(100 * time.Millisecond)
					defer l.Cleanup()
					upgradeErr = upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, passedDryRun, true, l)
					// Under no circumstances we want an actual upgrade to be scheduled
					Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(0))
				})

				// When `checkUpToDate: true` it takes precedence and the actual value of dryRun should be` irrelevant and lead to the exact same results
				var upgradeErr2 error
				upgradeLog2 := captureStdout(func() {
					l := logger.New(100 * time.Millisecond)
					defer l.Cleanup()
					upgradeErr2 = upgrader.Upgrade(fakeCFClient, fakeBrokerName, 1, !passedDryRun, true, l)
					// Under no circumstances we want an actual upgrade to be scheduled
					Expect(fakeCFClient.UpgradeServiceInstanceCallCount()).Should(Equal(0))
				})

				if expectedErr == nil {
					Expect(upgradeErr).NotTo(HaveOccurred())
					Expect(upgradeErr2).NotTo(HaveOccurred())
				} else {
					Expect(upgradeErr).To(Equal(expectedErr))
					Expect(upgradeErr2).To(Equal(expectedErr))
				}
				if expectedLog == "" {
					Expect(upgradeLog).To(Equal(""))
					Expect(upgradeLog2).To(Equal(""))
				} else {
					Expect(upgradeLog).To(MatchRegexp(expectedLog))
					Expect(upgradeLog2).To(MatchRegexp(expectedLog))
				}
			},
			Entry("no plans defined",
				fmt.Errorf("no service plans available for broker: fake-broker-name"), "",
				[]ccapi.Plan{},
				[]ccapi.ServiceInstance{},
			),
			Entry("no instances defined",
				nil, "no instances available to upgrade",
				[]ccapi.Plan{fakePlan},
				[]ccapi.ServiceInstance{},
			),
			Entry("all instances up to date",
				nil, "no instances available to upgrade",
				[]ccapi.Plan{fakePlan},
				[]ccapi.ServiceInstance{{UpgradeAvailable: false}},
			),
			Entry("upgradable instances",
				fmt.Errorf("check up-to-date failed: found 1 instances which are not up-to-date"),
				`(\s|.)*`, // we don't care about the logs in this test
				[]ccapi.Plan{fakePlan},
				[]ccapi.ServiceInstance{{UpgradeAvailable: true}},
			),
			Entry("upgradable instances detailed logs",
				fmt.Errorf("check up-to-date failed: found 1 instances which are not up-to-date"),
				getExpectedLogForInstance(getFakeInstanceDetailed()),
				[]ccapi.Plan{{GUID: getFakeInstanceDetailed().PlanGUID, MaintenanceInfoVersion: getFakeInstanceDetailed().PlanMaintenanceInfoVersion}},
				[]ccapi.ServiceInstance{getFakeInstanceDetailed()},
			),
			Entry("upgradeable instances whose creation failed",
				nil, "no instances available to upgrade",
				[]ccapi.Plan{{GUID: fakeInstance1.PlanGUID}},
				[]ccapi.ServiceInstance{{UpgradeAvailable: true, LastOperation: ccapi.LastOperation{Type: "create", State: "failed"}}},
			),
			/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			// WARNING:
			// We may want to check whether the following scenarios are possible and if they do, we may want to prevent them from happening
			/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
			Entry("upgradeable instances whose creation failed and are in a possible invalid state?",
				fmt.Errorf("check up-to-date failed: found 1 instances which are not up-to-date"),
				`(\s|.)*`, // we don't care about the logs in this test
				[]ccapi.Plan{{GUID: fakeInstance1.PlanGUID}},
				[]ccapi.ServiceInstance{{UpgradeAvailable: true, LastOperation: ccapi.LastOperation{Type: "destroy", State: "failed"}}},
			),
			Entry("unlikely scenario in which the api returns instances which doesn't correspond to the queried plan",
				fmt.Errorf("check up-to-date failed: found 1 instances which are not up-to-date"),
				`(\s|.)*`, // we don't care about the logs in this test
				[]ccapi.Plan{{GUID: "guid-for-unlikely-scenario"}},
				[]ccapi.ServiceInstance{getFakeInstanceDetailed()},
			),
		)
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

func getFakeInstanceDetailed() ccapi.ServiceInstance {
	return ccapi.ServiceInstance{
		UpgradeAvailable:           true,
		Name:                       "fakeName",
		GUID:                       "fakeGUID",
		MaintenanceInfoVersion:     "fakeMaintenanceInfoVersion",
		PlanGUID:                   "fakePlanGUID",
		PlanMaintenanceInfoVersion: "fakePlanMaintenanceInfoVersion",
		LastOperation: ccapi.LastOperation{
			Type:  "fakeType",
			State: "fakeState",
		},

		Included: ccapi.EmbeddedInclude{
			Organization: ccapi.Organization{
				Name: "fakeOrgName",
				GUID: "fakeOrgGUID",
			},
			Space: ccapi.Space{
				Name: "fakseSpaceName",
				GUID: "fakeSpaceGUID",
			},
			Plan: ccapi.IncludedPlan{
				Name: "fakePlanName",
				GUID: "fakePlanGUID",
			},
			ServiceOffering: ccapi.ServiceOffering{
				Name: "fakeServiceOfferingName",
				GUID: "fakeServiceOfferingGUID",
			},
		},
	}
}

func getExpectedLogForInstance(i ccapi.ServiceInstance) string {
	return `(\s|.)*total instances: 1` +
		`(\s|.)*upgradable instances: 1` +
		`(\s|.)*Service Instance Name: "` + i.Name +
		`(\s|.)*Service Instance GUID: "` + i.GUID +
		`(\s|.)*Service Version: "` + i.MaintenanceInfoVersion +
		`(\s|.)*Details: "dry-run prevented upgrade"` +
		`(\s|.)*Org Name: "` + i.Included.Organization.Name +
		`(\s|.)*Org GUID: "` + i.Included.Organization.GUID +
		`(\s|.)*Space Name: "` + i.Included.Space.Name +
		`(\s|.)*Space GUID: "` + i.Included.Space.GUID +
		`(\s|.)*Plan Name: "` + i.Included.Plan.Name +
		`(\s|.)*Plan GUID: "` + i.Included.Plan.GUID +
		`(\s|.)*Plan Version: "` + i.PlanMaintenanceInfoVersion +
		`(\s|.)*Service Offering Name: "` + i.Included.ServiceOffering.Name +
		`(\s|.)*Service Offering GUID: "` + i.Included.ServiceOffering.GUID
}
