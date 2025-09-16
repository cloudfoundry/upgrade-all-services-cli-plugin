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
						fakecapi.ServiceInstance{Name: "service-instance-4", UpgradeAvailable: false, Version: "1.2.3", LastOperationType: "create"},
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
\S+: upgrade of instance: "service-instance-2" guid: "0ec2261c-5d50-c12e-4e8b-ca9273c6150f" failed after 0s: dry-run prevented upgrade instance guid 0ec2261c-5d50-c12e-4e8b-ca9273c6150f
\S+: upgrade of instance: "service-instance-3" guid: "ef7fa19f-0d66-55d0-0519-f198164d358c" failed after 0s: dry-run prevented upgrade instance guid ef7fa19f-0d66-55d0-0519-f198164d358c
\S+: skipping instance: "service-instance-4" guid: "c53ccd0e-b88e-0d93-712d-609588651af0" Deactivated Plan: "service-plan-3" Offering: "service-offering-2" Offering guid: "dda79e55-6ef6-5f90-4cd7-174fb300b1ea" Upgrade Available: false Last Operation Type: "create" State: "succeeded"
\S+: upgraded 2 of 2
\S+: ---
\S+: skipped 1 instances
\S+: successfully upgraded 0 instances
\S+: failed to upgrade 2 instances
\S+: 

\s+Service Instance Name: "service-instance-2"
\s+Service Instance GUID: "0ec2261c-5d50-c12e-4e8b-ca9273c6150f"
\s+Service Version: "1.2.2"
\s+Details: "dry-run prevented upgrade instance guid 0ec2261c-5d50-c12e-4e8b-ca9273c6150f"
\s+Org Name: "fake-org"
\s+Org GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
\s+Space Name: "fake-space"
\s+Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
\s+Plan Name: "service-plan1"
\s+Plan GUID: "173a3f22-e23f-27f2-9b32-8efdb64d5c14"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-1"
\s+Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"


\s+Service Instance Name: "service-instance-3"
\s+Service Instance GUID: "ef7fa19f-0d66-55d0-0519-f198164d358c"
\s+Service Version: "1.2.0"
\s+Details: "dry-run prevented upgrade instance guid ef7fa19f-0d66-55d0-0519-f198164d358c"
\s+Org Name: "fake-org"
\s+Org GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
\s+Space Name: "fake-space"
\s+Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
\s+Plan Name: "service-plan-2"
\s+Plan GUID: "3ccc0ed1-1c06-036b-7bfe-f4d9dff25d02"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-1"
\s+Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"
`)))
		Expect(session.Err).To(Say(strings.TrimSpace(`
upgrade-all-services plugin failed: found 2 instances which are not up-to-date
discovered deactivated plans associated with instances. Review the log to collect information and restore the deactivated plans or create user provided services
`)))
	})
})
