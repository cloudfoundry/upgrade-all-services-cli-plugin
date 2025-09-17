package integrationtests_test

import (
	"fmt"
	"strings"
	"time"
	"upgrade-all-services-cli-plugin/internal/fakecapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("upgrade", func() {
	const brokerName = "upgrade-broker"

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
			session := cf("upgrade-all-services", brokerName)
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

			Expect(capi.MaxOperations).To(Equal(10))
			Expect(capi.UpdateCount()).To(Equal(1000))
		})

		It("obeys the specified -parallel flag", func() {
			session := cf("upgrade-all-services", brokerName, "--parallel", "25")
			Eventually(session).WithTimeout(time.Minute).Should(Exit(0))

			Expect(capi.MaxOperations).To(Equal(25))
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
			session := cf("upgrade-all-services", brokerName)
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
\S+: upgraded 15 of 15
\S+: ---
\S+: skipped 0 instances
\S+: successfully upgraded 5 instances
\S+: failed to upgrade 10 instances
`)))

			Expect(capi.UpdateCount()).To(Equal(15))

			expectedFailures := map[string]string{
				"fake-instance-6":  "07b9b83d-419e-b968-7a61-d55895af3466fcc76eabeaa4bc0735c5b605306d0738",
				"fake-instance-7":  "8df19a1a-63d4-5789-3427-228f485e03fda46ac199eb215ef943e3bccde9ffd852",
				"fake-instance-8":  "73bd62c1-94ab-e6ec-6ffc-24c994b124d535125bff49dac290158b0a25039f6ee3",
				"fake-instance-9":  "036b82e1-6bea-1db0-81ec-b0b6628a67ae8e21d596149404f6974d6848f14edd28",
				"fake-instance-10": "5e1f4213-272c-0d56-1fdb-d6f85a0d71cfe73a922572b972153d1e855fc7f406ac",
				"fake-instance-11": "43645af7-94b7-7880-6a66-ea39704bc7308c3b669104787ff4c1575a7468d3f80b",
				"fake-instance-12": "abca2199-4ddb-7537-ba3f-84b0ed4722448fa8e538f272409c9d4a3cac33478229",
				"fake-instance-13": "228ffe48-1f51-2b7b-5da5-ca114987697c6a8971d1407854c137a9d2ecd197cc33",
				"fake-instance-14": "ae92ddf7-ca98-8f3f-048f-2d7ead0ff3e1919cdea572c55dcfabac5aadc21f7b95",
				"fake-instance-15": "c23c4e8b-1fbb-5aef-ef5b-fc7bcfd120ce2af37f5658d0c154e5772d10e34f5b42",
			}

			for name, guid := range expectedFailures {
				Expect(string(session.Out.Contents())).To(MatchRegexp(fmt.Sprintf(`upgrade of instance: "%s" guid: "%s" failed after \S+: failed as requested by test setup`, name, guid)))

				Expect(string(session.Out.Contents())).To(MatchRegexp(strings.TrimSpace(fmt.Sprintf(`
\s+Service Instance Name: "%s"
\s+Service Instance GUID: "%s"
\s+Service Version: "1.2.2"
\s+Details: "failed as requested by test setup"
\s+Org Name: "fake-org"
\s+Org GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
\s+Space Name: "fake-space"
\s+Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
\s+Plan Name: "service-plan1"
\s+Plan GUID: "173a3f22-e23f-27f2-9b32-8efdb64d5c14254a50e4ad138cb08b109433e249a934"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-1"
\s+Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa2e7736a44b2e4ea39df28a5c1e96c760"
`, name, guid))))
			}

			expectedSuccesses := map[string]string{
				"fake-instance-1": "23525d72-bd78-5a0e-283b-cca10567e5c039c4791506dfe91297255359639054c8",
				"fake-instance-2": "31dc65f7-0792-446a-024b-6b6fe613f99f21244c8d14ae2fb58e88a348502fdcf7",
				"fake-instance-3": "285c7aac-ff1f-54ea-94ae-10e28f191b5a031b9a6de68ed094cc285fb62fe6893c",
				"fake-instance-4": "e2b68b0d-3f8c-7a7f-facc-493f8dd1353feb4c6cd57c08fa65d717f3bb4eb34e75",
				"fake-instance-5": "065401a6-c9a3-2b2e-2894-9ec4a562066d2c54a68a67fcc7824f6903a8b6f01366",
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
})
