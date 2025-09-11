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

var _ = Describe("MinVersionRequired", func() {
	const brokerName = "min-ver-broker"

	BeforeEach(func() {
		capi.AddBroker(
			fakecapi.ServiceBroker{Name: brokerName},
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-1"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan1", Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-1", Version: "1.2.3"},
						fakecapi.ServiceInstance{Name: "service-instance-2", Version: "1.2.2"},
					),
				),
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-2", Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-3", Version: "1.2.0"},
					),
				),
			),
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-2"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-3", Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-4", Version: "1.2.1"},
					),
				),
			),
		)
	})

	It("detects versions below the minimum", func() {
		session := cf("upgrade-all-services", brokerName, "-min-version-required", "1.2.3")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: min-ver-broker
\S+: ---
\S+: total instances: 4
\S+: upgradable instances: 3
\S+: ---
\S+: starting upgrade...
\S+: upgrade of instance: "service-instance-2" guid: "\S+" failed after 0s: dry-run prevented upgrade instance guid \S+
\S+: upgrade of instance: "service-instance-3" guid: "\S+" failed after 0s: dry-run prevented upgrade instance guid \S+
\S+: upgrade of instance: "service-instance-4" guid: "\S+" failed after 0s: dry-run prevented upgrade instance guid \S+
\S+: upgraded 3 of 3
\S+: ---
\S+: skipped 0 instances
\S+: successfully upgraded 0 instances
\S+: failed to upgrade 3 instances
\S+:\s

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


\s+Service Instance Name: "service-instance-4"
\s+Service Instance GUID: "\S+"
\s+Service Version: "1.2.1"
\s+Details: "dry-run prevented upgrade instance guid \S+"
\s+Org Name: "fake-org"
\s+Org GUID: "\S+"
\s+Space Name: "fake-space"
\s+Space GUID: "\S+"
\s+Plan Name: "service-plan-3"
\s+Plan GUID: "\S+"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-2"
\s+Service Offering GUID: "\S+"
`)))
		Expect(session.Err).To(Say(`upgrade-all-services plugin failed: found 3 service instances with a version less than the minimum required`))
	})

	It("ignores versions at or above the minimum", func() {
		session := cf("upgrade-all-services", brokerName, "-min-version-required", "1.2.2")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: min-ver-broker
\S+: ---
\S+: total instances: 4
\S+: upgradable instances: 2
\S+: ---
\S+: starting upgrade...
\S+: upgrade of instance: "service-instance-3" guid: "\S+" failed after 0s: dry-run prevented upgrade instance guid \S+
\S+: upgrade of instance: "service-instance-4" guid: "\S+" failed after 0s: dry-run prevented upgrade instance guid \S+
\S+: upgraded 2 of 2
\S+: ---
\S+: skipped 0 instances
\S+: successfully upgraded 0 instances
\S+: failed to upgrade 2 instances
\S+: 

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


\s+Service Instance Name: "service-instance-4"
\s+Service Instance GUID: "\S+"
\s+Service Version: "1.2.1"
\s+Details: "dry-run prevented upgrade instance guid \S+"
\s+Org Name: "fake-org"
\s+Org GUID: "\S+"
\s+Space Name: "fake-space"
\s+Space GUID: "\S+"
\s+Plan Name: "service-plan-3"
\s+Plan GUID: "\S+"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-2"
\s+Service Offering GUID: "\S+"
`)))
		Expect(session.Err).To(Say(`upgrade-all-services plugin failed: found 2 service instances with a version less than the minimum required`))
	})

	It("succeeds when all instances are at or above the minimum", func() {
		session := cf("upgrade-all-services", brokerName, "-min-version-required", "1.2.0")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: min-ver-broker
\S+: no instances found with version less than required
`)))
	})
})
