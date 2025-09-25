package integrationtests_test

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"upgrade-all-services-cli-plugin/internal/fakecapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("upgrade", func() {
	const brokerName = "upgrade-broker"

	// cfFast runs the CF CLI with a small instance polling interval to ensure that tests run fast.
	// With the default polling interval, the tests take minutes rather than seconds.
	// There's no specific test for the instance polling interval.
	cfFast := func(args ...string) *Session {
		return cf(append(args, "--instance-polling-interval", "1ms")...)
	}

	Context("successful upgrades", func() {
		BeforeEach(func() {
			capi.AddBroker(
				fakecapi.ServiceBroker{Name: brokerName},
				fakecapi.WithServiceOffering(
					fakecapi.ServiceOffering{Name: "service-offering-1"},
					fakecapi.WithServicePlan(
						fakecapi.ServicePlan{Name: "service-plan1", Version: "1.2.3"},
						fakecapi.WithServiceInstances(repeat(1000, fakecapi.ServiceInstance{UpgradeAvailable: true, Version: "1.2.2", UpdateTime: 10 * time.Millisecond})...),
						fakecapi.WithServiceInstances(repeat(100, fakecapi.ServiceInstance{UpgradeAvailable: false, Version: "1.2.3"})...),
					),
				),
			)
		})

		It("upgrades many service instances", func() {
			session := cfFast("upgrade-all-services", brokerName)
			Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
			Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: upgrade-broker
\S+: ---
\S+: total instances: 1100
\S+: upgradable instances: 1000
\S+: ---
\S+: starting upgrade...
`)))

			Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: upgraded 1000 of 1000
\S+: ---
\S+: skipped 0 instances
\S+: successfully upgraded 1000 instances
`)))

			Expect(capi.MaxConcurrentOperations).To(Equal(10))
			Expect(capi.UpdateCount()).To(Equal(1000))
		})

		It("respects the specified -parallel flag", func() {
			session := cfFast("upgrade-all-services", brokerName, "--parallel", "25")
			Eventually(session).WithTimeout(time.Minute).Should(Exit(0))

			Expect(capi.MaxConcurrentOperations).To(Equal(25))
		})

		It("respects the specified -limit flag", func() {
			session := cfFast("upgrade-all-services", brokerName, "--limit", "538")
			Eventually(session).WithTimeout(time.Minute).Should(Exit(0))

			Expect(capi.UpdateCount()).To(Equal(538))
		})
	})

	Context("failed upgrades", func() {
		BeforeEach(func() {
			capi.AddBroker(
				fakecapi.ServiceBroker{Name: brokerName},
				fakecapi.WithServiceOffering(
					fakecapi.ServiceOffering{Name: "service-offering-1"},
					fakecapi.WithServicePlan(
						fakecapi.ServicePlan{Name: "service-plan1", Version: "1.2.3"},
						fakecapi.WithServiceInstances(repeat(5, fakecapi.ServiceInstance{UpgradeAvailable: true, Version: "1.2.2", UpdateTime: 10 * time.Millisecond})...),
						fakecapi.WithServiceInstances(repeat(10, fakecapi.ServiceInstance{UpgradeAvailable: true, Version: "1.2.2", UpdateTime: 10 * time.Millisecond, FailTimes: 1})...),
						fakecapi.WithServiceInstances(repeat(10, fakecapi.ServiceInstance{UpgradeAvailable: false, Version: "1.2.3"})...),
					),
				),
			)
		})

		It("reports the successes and failures", func() {
			session := cfFast("upgrade-all-services", brokerName)
			Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
			Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: upgrade-broker
\S+: ---
\S+: total instances: 25
\S+: upgradable instances: 15
\S+: ---
\S+: starting upgrade...
`)))

			Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: upgraded 5 of 15
\S+: ---
\S+: skipped 0 instances
\S+: successfully upgraded 5 instances
\S+: failed to upgrade 10 instances
`)))

			Expect(capi.UpdateCount()).To(Equal(15))

			expectedFailures := map[string]string{
				"fake-instance-6":  "07b9b83d-419e-b968-7a61-d55895af3466",
				"fake-instance-7":  "8df19a1a-63d4-5789-3427-228f485e03fd",
				"fake-instance-8":  "73bd62c1-94ab-e6ec-6ffc-24c994b124d5",
				"fake-instance-9":  "036b82e1-6bea-1db0-81ec-b0b6628a67ae",
				"fake-instance-10": "5e1f4213-272c-0d56-1fdb-d6f85a0d71cf",
				"fake-instance-11": "43645af7-94b7-7880-6a66-ea39704bc730",
				"fake-instance-12": "abca2199-4ddb-7537-ba3f-84b0ed472244",
				"fake-instance-13": "228ffe48-1f51-2b7b-5da5-ca114987697c",
				"fake-instance-14": "ae92ddf7-ca98-8f3f-048f-2d7ead0ff3e1",
				"fake-instance-15": "c23c4e8b-1fbb-5aef-ef5b-fc7bcfd120ce",
			}

			for name, guid := range expectedFailures {
				Expect(string(session.Out.Contents())).To(MatchRegexp(fmt.Sprintf(`upgrade of instance: "%s" guid: "%s" failed after \S+: failed as requested by test setup`, name, guid)))

				Expect(string(session.Out.Contents())).To(MatchRegexp(strings.TrimSpace(fmt.Sprintf(`
  Details: "failed as requested by test setup"
  Service Instance Name: "%s"
  Service Instance GUID: "%s"
  Service Instance Version: "1.2.2"
  Service Plan Name: "service-plan1"
  Service Plan GUID: "173a3f22-e23f-27f2-9b32-8efdb64d5c14"
  Service Plan Version: "1.2.3"
  Service Offering Name: "service-offering-1"
  Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
`, name, guid))))
			}

			expectedSuccesses := map[string]string{
				"fake-instance-1": "23525d72-bd78-5a0e-283b-cca10567e5c0",
				"fake-instance-2": "31dc65f7-0792-446a-024b-6b6fe613f99f",
				"fake-instance-3": "285c7aac-ff1f-54ea-94ae-10e28f191b5a",
				"fake-instance-4": "e2b68b0d-3f8c-7a7f-facc-493f8dd1353f",
				"fake-instance-5": "065401a6-c9a3-2b2e-2894-9ec4a562066d",
			}

			for name, guid := range expectedSuccesses {
				Expect(string(session.Out.Contents())).To(MatchRegexp(fmt.Sprintf(`finished upgrade of instance: "%s" guid: "%s" successfully after \S+`, name, guid)))
			}

			// No upgrade expected for any of these, so they should not be mentioned
			Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
				ContainSubstring("fake-instance-16"),
				ContainSubstring("fake-instance-17"),
				ContainSubstring("fake-instance-18"),
				ContainSubstring("fake-instance-19"),
				ContainSubstring("fake-instance-20"),
				ContainSubstring("fake-instance-21"),
				ContainSubstring("fake-instance-22"),
				ContainSubstring("fake-instance-23"),
				ContainSubstring("fake-instance-24"),
				ContainSubstring("fake-instance-25"),
			))
		})
	})

	Context("retrying after a failure", func() {
		const (
			numSucceedOnFirstAttempt = 89
			numFailOnFirstAttempts   = 52
			numOfAttemptsToFail      = 2
		)

		BeforeEach(func() {
			capi.AddBroker(
				fakecapi.ServiceBroker{Name: brokerName},
				fakecapi.WithServiceOffering(
					fakecapi.ServiceOffering{Name: "service-offering-1"},
					fakecapi.WithServicePlan(
						fakecapi.ServicePlan{Name: "service-plan1", Version: "1.2.3"},
						fakecapi.WithServiceInstances(repeat(numSucceedOnFirstAttempt, fakecapi.ServiceInstance{UpgradeAvailable: true, Version: "1.2.2", UpdateTime: 10 * time.Millisecond})...),
						fakecapi.WithServiceInstances(repeat(numFailOnFirstAttempts, fakecapi.ServiceInstance{UpgradeAvailable: true, Version: "1.2.2", UpdateTime: 10 * time.Millisecond, FailTimes: numOfAttemptsToFail})...),
						fakecapi.WithServiceInstances(repeat(100, fakecapi.ServiceInstance{UpgradeAvailable: false, Version: "1.2.3"})...),
					),
				),
			)
		})

		It("respects the specified -attempts flag", func() {
			session := cfFast("upgrade-all-services", brokerName, "-attempts", "3")
			Eventually(session).WithTimeout(time.Minute).Should(Exit(1))

			By("making the correct number of upgrade requests")
			Expect(capi.UpdateCount()).To(Equal(numSucceedOnFirstAttempt + (numOfAttemptsToFail+1)*numFailOnFirstAttempts))

			By("logging the correct output for an example service instance")
			Expect(session.Out).To(Say(`\S+: starting to upgrade instance: "fake-instance-96" guid: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae" \(attempt 1 of 3\)`))
			Expect(session.Out).To(Say(`\S+: upgrade of instance: "fake-instance-96" guid: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae" failed after \S+ \(attempt 1 of 3\): failed as requested by test setup`))
			Expect(session.Out).To(Say(`\S+: starting to upgrade instance: "fake-instance-96" guid: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae" \(attempt 2 of 3\)`))
			Expect(session.Out).To(Say(`\S+: upgrade of instance: "fake-instance-96" guid: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae" failed after \S+ \(attempt 2 of 3\): failed as requested by test setup`))
			Expect(session.Out).To(Say(`\S+: starting to upgrade instance: "fake-instance-96" guid: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae" \(attempt 3 of 3\)`))
			Expect(session.Out).To(Say(`\S+: finished upgrade of instance: "fake-instance-96" guid: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae" successfully after \S+ \(attempt 3 of 3\)`))
			Expect(session.Out).To(Say(strings.TrimSpace(`
  Details: "failed as requested by test setup"
  Attempt 1 of 3
  Service Instance Name: "fake-instance-96"
  Service Instance GUID: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae"
  Service Instance Version: "1.2.2"
  Service Plan Name: "service-plan1"
  Service Plan GUID: "173a3f22-e23f-27f2-9b32-8efdb64d5c14"
  Service Plan Version: "1.2.3"
  Service Offering Name: "service-offering-1"
  Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
`)))
			Expect(session.Out).To(Say(strings.TrimSpace(`
  Details: "failed as requested by test setup"
  Attempt 2 of 3
  Service Instance Name: "fake-instance-96"
  Service Instance GUID: "51a015ab-2a9c-63d4-a2f1-06bd2931ebae"
  Service Instance Version: "1.2.2"
  Service Plan Name: "service-plan1"
  Service Plan GUID: "173a3f22-e23f-27f2-9b32-8efdb64d5c14"
  Service Plan Version: "1.2.3"
  Service Offering Name: "service-offering-1"
  Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
`)))
		})
	})

	Context("waiting between retries", func() {
		var times []time.Time

		BeforeEach(func() {
			var lock sync.Mutex
			cb := func() {
				lock.Lock()
				defer lock.Unlock()
				times = append(times, time.Now())
			}

			capi.AddBroker(
				fakecapi.ServiceBroker{Name: brokerName},
				fakecapi.WithServiceOffering(
					fakecapi.ServiceOffering{Name: "service-offering-1"},
					fakecapi.WithServicePlan(
						fakecapi.ServicePlan{Name: "service-plan1", Version: "1.2.3"},
						fakecapi.WithServiceInstances(fakecapi.ServiceInstance{UpgradeAvailable: true, Version: "1.2.2", UpdateTime: time.Microsecond, FailTimes: 1, Callback: cb}),
					),
				),
			)
		})

		It("respects the -retry-interval flag", func() {
			session := cfFast("upgrade-all-services", brokerName, "-attempts", "2", "-retry-interval", "100ms")
			Eventually(session).WithTimeout(time.Minute).Should(Exit(1))

			// Ensure that the second attempt is 100ms after the first with an accuracy of 10ms
			Expect(times).To(HaveLen(2))
			first := times[0]
			second := times[1]
			Expect(second.Sub(first)).To(BeNumerically("~", 100*time.Millisecond, 10*time.Millisecond))
		})
	})
})
