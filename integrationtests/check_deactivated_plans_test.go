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
						fakecapi.ServiceInstance{Name: "service-instance-3", UpgradeAvailable: true, Version: "1.1.0", LastOperationType: "create", LastOperationState: "succeeded"},
					),
				),
			),
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-2"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-3", Available: false, Version: "1.3.0"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-4", UpgradeAvailable: false, Version: "1.3.0", LastOperationType: "upgrade", LastOperationState: "failed"},
					),
				),
			),
		)
	})

	It("shows which service instances belong to deactivated plans", func() {
		session := cf("upgrade-all-services", brokerName, "-check-deactivated-plans")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: check-deactivated-plans-broker
\S+: skipping instance: "service-instance-3" guid: "\S+" Deactivated Plan: "service-plan-2" Offering: "service-offering-1" Offering guid: "\S+" Upgrade Available: true Last Operation Type: "create" State: "succeeded"
\S+: skipping instance: "service-instance-4" guid: "\S+" Deactivated Plan: "service-plan-3" Offering: "service-offering-2" Offering guid: "\S+" Upgrade Available: false Last Operation Type: "upgrade" State: "failed"
`)))
		Expect(session.Err).To(Say(`upgrade-all-services plugin failed: discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services`))
	})
})
