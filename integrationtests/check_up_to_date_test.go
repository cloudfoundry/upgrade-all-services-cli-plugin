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

var _ = Describe("-check-up-to-date", func() {
	const brokerName = "check-up-to-date-broker"

	BeforeEach(func() {
		capi.AddBroker(
			fakecapi.ServiceBroker{Name: brokerName},
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-1"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan1", Available: true, Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-1", UpgradeAvailable: false, Version: "1.2.3"},
						fakecapi.ServiceInstance{Name: "service-instance-2", UpgradeAvailable: true, Version: "1.2.2"},
					),
				),
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-2", Available: true, Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-3", UpgradeAvailable: true, Version: "1.2.0"},
					),
				),
			),
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-2"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-3", Available: false, Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-4", UpgradeAvailable: false, Version: "1.2.3", LastOperationType: "create", LastOperationState: "succeeded"},
					),
				),
			),
		)
	})

	It("does a dry run and checks for deactivated plans", func() {
		session := cf("upgrade-all-services", brokerName, "-check-up-to-date")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: check-up-to-date-broker
\S+: ---
\S+: total instances: 4
\S+: upgradable instances: 2
\S+: ---
\S+: starting upgrade...
\S+: upgrade of instance: "service-instance-2" guid: "\S+" failed after 0s: dry-run prevented upgrade instance guid \S+
\S+: upgrade of instance: "service-instance-3" guid: "\S+" failed after 0s: dry-run prevented upgrade instance guid \S+
\S+: skipping instance: "service-instance-4" guid: "\S+" Deactivated Plan: "service-plan-3" Offering: "service-offering-2" Offering guid: "\S+" Upgrade Available: false Last Operation Type: "create" State: "succeeded"
\S+: upgraded 2 of 2
\S+: ---
\S+: skipped 1 instances
\S+: successfully upgraded 0 instances
\S+: failed to upgrade 2 instances
\S+: 

\s+Service Instance Name: "service-instance-2"
\s+Service Instance GUID: "\S+"
\s+Service Version: "1.2.2"
\s+Details: "dry-run prevented upgrade instance guid \S+"
\s+Org Name: "fake-org"
\s+Org GUID: "\S+"
\s+Space Name: "fake-space"
\s+Space GUID: "\S+"
\s+Plan Name: "service-plan1"
\s+Plan GUID: "\S+"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-1"
\s+Service Offering GUID: "\S+"


\s+Service Instance Name: "service-instance-3"
\s+Service Instance GUID: "\S+"
\s+Service Version: "1.2.0"
\s+Details: "dry-run prevented upgrade instance guid \S+"
\s+Org Name: "fake-org"
\s+Org GUID: "\S+"
\s+Space Name: "fake-space"
\s+Space GUID: "\S+"
\s+Plan Name: "service-plan-2"
\s+Plan GUID: "\S+"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-1"
\s+Service Offering GUID: "\S+"
`)))
		Expect(session.Err).To(Say(strings.TrimSpace(`
upgrade-all-services plugin failed: found 2 instances which are not up-to-date
discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services
`)))
	})
})
