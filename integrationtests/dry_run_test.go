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

var _ = Describe("-dry-run", func() {
	const brokerName = "dry-run-broker"

	BeforeEach(func() {
		capi.AddBroker(
			fakecapi.ServiceBroker{Name: brokerName},
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-1"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan1", Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-1", UpgradeAvailable: true, Version: "1.2.2"},
						fakecapi.ServiceInstance{Name: "service-instance-2", UpgradeAvailable: false, Version: "1.2.3"},
					),
				),
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-2", Version: "1.2.0"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-3", UpgradeAvailable: true, Version: "1.1.0"},
					),
				),
			),
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-2"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-3", Version: "1.3.0"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-4", UpgradeAvailable: true, Version: "1.2.9"},
					),
				),
			),
		)
	})

	It("shows which service instances would be upgraded", func() {
		session := cf("upgrade-all-services", brokerName, "-dry-run")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: dry-run-broker
\S+: ---
\S+: total instances: 4
\S+: upgradable instances: 3
\S+: ---
\S+: starting upgrade...
\S+: upgrade of instance: "service-instance-1" guid: "5cc87b43-f885-3b94-328f-8a5f953590d341f6730f5ba530f723a6ab0fea651bf6" failed after 0s: dry-run prevented upgrade instance guid 5cc87b43-f885-3b94-328f-8a5f953590d341f6730f5ba530f723a6ab0fea651bf6
\S+: upgrade of instance: "service-instance-3" guid: "ef7fa19f-0d66-55d0-0519-f198164d358ce662614b25499cd4ebf411f5e6ea55ae" failed after 0s: dry-run prevented upgrade instance guid ef7fa19f-0d66-55d0-0519-f198164d358ce662614b25499cd4ebf411f5e6ea55ae
\S+: upgrade of instance: "service-instance-4" guid: "c53ccd0e-b88e-0d93-712d-609588651af020db5207350e8f031a12585cef7accd9" failed after 0s: dry-run prevented upgrade instance guid c53ccd0e-b88e-0d93-712d-609588651af020db5207350e8f031a12585cef7accd9
\S+: upgraded 3 of 3
\S+: ---
\S+: skipped 0 instances
\S+: successfully upgraded 0 instances
\S+: failed to upgrade 3 instances
\S+: 

\s+Service Instance Name: "service-instance-1"
\s+Service Instance GUID: "5cc87b43-f885-3b94-328f-8a5f953590d341f6730f5ba530f723a6ab0fea651bf6"
\s+Service Version: "1.2.2"
\s+Details: "dry-run prevented upgrade instance guid 5cc87b43-f885-3b94-328f-8a5f953590d341f6730f5ba530f723a6ab0fea651bf6"
\s+Org Name: "fake-org"
\s+Org GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
\s+Space Name: "fake-space"
\s+Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
\s+Plan Name: "service-plan1"
\s+Plan GUID: "173a3f22-e23f-27f2-9b32-8efdb64d5c14254a50e4ad138cb08b109433e249a934"
\s+Plan Version: "1.2.3"
\s+Service Offering Name: "service-offering-1"
\s+Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa2e7736a44b2e4ea39df28a5c1e96c760"


\s+Service Instance Name: "service-instance-3"
\s+Service Instance GUID: "ef7fa19f-0d66-55d0-0519-f198164d358ce662614b25499cd4ebf411f5e6ea55ae"
\s+Service Version: "1.1.0"
\s+Details: "dry-run prevented upgrade instance guid ef7fa19f-0d66-55d0-0519-f198164d358ce662614b25499cd4ebf411f5e6ea55ae"
\s+Org Name: "fake-org"
\s+Org GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
\s+Space Name: "fake-space"
\s+Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
\s+Plan Name: "service-plan-2"
\s+Plan GUID: "3ccc0ed1-1c06-036b-7bfe-f4d9dff25d02a8992444e56e9743dd4de35058e8373d"
\s+Plan Version: "1.2.0"
\s+Service Offering Name: "service-offering-1"
\s+Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa2e7736a44b2e4ea39df28a5c1e96c760"


\s+Service Instance Name: "service-instance-4"
\s+Service Instance GUID: "c53ccd0e-b88e-0d93-712d-609588651af020db5207350e8f031a12585cef7accd9"
\s+Service Version: "1.2.9"
\s+Details: "dry-run prevented upgrade instance guid c53ccd0e-b88e-0d93-712d-609588651af020db5207350e8f031a12585cef7accd9"
\s+Org Name: "fake-org"
\s+Org GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
\s+Space Name: "fake-space"
\s+Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
\s+Plan Name: "service-plan-3"
\s+Plan GUID: "51f29f1b-d343-6bdd-0192-deb80d4c6d9f15383c397664b3d6c28372cd53816129"
\s+Plan Version: "1.3.0"
\s+Service Offering Name: "service-offering-2"
\s+Service Offering GUID: "dda79e55-6ef6-5f90-4cd7-174fb300b1ea4412b95dbebf58dda02ee194c7c4598b"
`)))
	})
})
