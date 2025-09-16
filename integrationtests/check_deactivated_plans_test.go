package integrationtests_test

import (
	"strings"
	"time"
	"upgrade-all-services-cli-plugin/internal/fakecapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("-check-deactivated-plans", func() {
	const brokerName = "check-deactivated-plans-broker"

	BeforeEach(func() {
		capi.AddBroker(
			fakecapi.ServiceBroker{Name: brokerName},
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-1"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan1", Available: true, Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-1", UpgradeAvailable: true, Version: "1.2.2"},
						fakecapi.ServiceInstance{Name: "service-instance-2", UpgradeAvailable: false, Version: "1.2.3"},
					),
				),
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-2", Available: false, Version: "1.2.0"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-3", UpgradeAvailable: true, Version: "1.1.0", LastOperationType: "create"},
					),
				),
			),
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-2"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-3", Available: false, Version: "1.3.0"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-4", UpgradeAvailable: false, Version: "1.3.0", LastOperationState: "failed"},
					),
				),
			),
		)
	})

	It("shows which service instances belong to deactivated plans", func() {
		By("firstly checking that we have instances in need of upgrade")
		session := cf("upgrade-all-services", brokerName, "-min-version-required", "1.4.0")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(string(session.Out.Contents())).To(SatisfyAll(
			ContainSubstring(`service-instance-1`),
			ContainSubstring(`service-instance-2`),
			ContainSubstring(`service-instance-3`),
			ContainSubstring(`service-instance-4`),
		))

		By("observing that some belong to deactivated plans")
		session = cf("upgrade-all-services", brokerName, "-check-deactivated-plans")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: check-deactivated-plans-broker
\S+: skipping instance: "service-instance-3" guid: "ef7fa19f-0d66-55d0-0519-f198164d358ce662614b25499cd4ebf411f5e6ea55ae" Deactivated Plan: "service-plan-2" Offering: "service-offering-1" Offering guid: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa2e7736a44b2e4ea39df28a5c1e96c760" Upgrade Available: true Last Operation Type: "create" State: "succeeded"
\S+: skipping instance: "service-instance-4" guid: "c53ccd0e-b88e-0d93-712d-609588651af020db5207350e8f031a12585cef7accd9" Deactivated Plan: "service-plan-3" Offering: "service-offering-2" Offering guid: "dda79e55-6ef6-5f90-4cd7-174fb300b1ea4412b95dbebf58dda02ee194c7c4598b" Upgrade Available: false Last Operation Type: "update" State: "failed"
`)))

		Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
			ContainSubstring(`service-instance-1`),
			ContainSubstring(`service-instance-2`),
		))
		Expect(session.Err).To(Say(`upgrade-all-services plugin failed: discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services`))
	})
})
