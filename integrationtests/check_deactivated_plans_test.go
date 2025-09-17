package integrationtests_test

import (
	"strings"
	"time"
	"upgrade-all-services-cli-plugin/internal/fakecapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	It("shows instances belonging to deactivated plans in text", func() {
		session := cf("upgrade-all-services", brokerName, "-check-deactivated-plans")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(strings.TrimSpace(string(session.Out.Contents()))).To(Equal(strings.TrimSpace(`
Discovering service instances for broker: check-deactivated-plans-broker
Total number of service instances: 4
Number of service instances associated with deactivated plans: 2

  Service Instance Name: "service-instance-3"
  Service Instance GUID: "ef7fa19f-0d66-55d0-0519-f198164d358c"
  Service Instance Version: "1.1.0"
  Service Plan Name: "service-plan-2"
  Service Plan GUID: "3ccc0ed1-1c06-036b-7bfe-f4d9dff25d02"
  Service Plan Version: "1.2.0"
  Service Offering Name: "service-offering-1"
  Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"

  Service Instance Name: "service-instance-4"
  Service Instance GUID: "c53ccd0e-b88e-0d93-712d-609588651af0"
  Service Instance Version: "1.3.0"
  Service Plan Name: "service-plan-3"
  Service Plan GUID: "51f29f1b-d343-6bdd-0192-deb80d4c6d9f"
  Service Plan Version: "1.3.0"
  Service Offering Name: "service-offering-2"
  Service Offering GUID: "dda79e55-6ef6-5f90-4cd7-174fb300b1ea"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
`)))
		Expect(string(session.Err.Contents())).To(Equal("upgrade-all-services plugin failed: discovered deactivated plans associated with instances"))
	})

	It("shows instances belonging to deactivated plans in JSON", func() {
		session := cf("upgrade-all-services", brokerName, "-check-deactivated-plans", "-json")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`
  [
    {
      "guid": "ef7fa19f-0d66-55d0-0519-f198164d358c",
      "maintenance_info": {
        "version": "1.1.0"
      },
      "name": "service-instance-3",
      "organization": {
        "guid": "1a2f43b5-1594-4247-a888-e8843ebd1b03",
        "name": "fake-org"
      },
      "service_offering": {
        "guid": "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa",
        "name": "service-offering-1"
      },
      "service_plan": {
        "guid": "3ccc0ed1-1c06-036b-7bfe-f4d9dff25d02",
        "name": "service-plan-2"
      },
      "space": {
        "guid": "5f870ea3-fa54-4174-ab3f-15f2d9516e07",
        "name": "fake-space"
      }
    },
    {
      "guid": "c53ccd0e-b88e-0d93-712d-609588651af0",
      "maintenance_info": {
        "version": "1.3.0"
      },
      "name": "service-instance-4",
      "organization": {
        "guid": "1a2f43b5-1594-4247-a888-e8843ebd1b03",
        "name": "fake-org"
      },
      "service_offering": {
        "guid": "dda79e55-6ef6-5f90-4cd7-174fb300b1ea",
        "name": "service-offering-2"
      },
      "service_plan": {
        "guid": "51f29f1b-d343-6bdd-0192-deb80d4c6d9f",
        "name": "service-plan-3"
      },
      "space": {
        "guid": "5f870ea3-fa54-4174-ab3f-15f2d9516e07",
        "name": "fake-space"
      }
    }
  ]
`))
	})
})
