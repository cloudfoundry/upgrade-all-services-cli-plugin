package integrationtests_test

import (
	"strings"
	"time"
	"upgrade-all-services-cli-plugin/internal/fakecapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
						fakecapi.ServiceInstance{Name: "service-instance-4", UpgradeAvailable: true, Version: "1.2.2", LastOperationType: "create", LastOperationState: "failed"},
					),
				),
			),
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-2"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-3", Available: false, Version: "1.2.3"},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-5", UpgradeAvailable: false, Version: "1.2.3", LastOperationType: "create"},
					),
				),
			),
		)
	})

	It("prints deactivated plans, pending upgrades and failed creates in text format", func() {
		session := cf("upgrade-all-services", brokerName, "-check-up-to-date")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(1))
		Expect(strings.TrimSpace(string(session.Out.Contents()))).To(Equal(strings.TrimSpace(`
Discovering service instances for broker: check-up-to-date-broker
Total number of service instances: 5
Number of service instances associated with deactivated plans: 1

  Service Instance Name: "service-instance-5"
  Service Instance GUID: "ab5b7eb3-4d38-eb33-2aec-dbf7416d1db3"
  Service Instance Version: "1.2.3"
  Service Plan Name: "service-plan-3"
  Service Plan GUID: "51f29f1b-d343-6bdd-0192-deb80d4c6d9f"
  Service Plan Version: "1.2.3"
  Service Offering Name: "service-offering-2"
  Service Offering GUID: "dda79e55-6ef6-5f90-4cd7-174fb300b1ea"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"

Number of service instances with an upgrade available: 2

  Service Instance Name: "service-instance-2"
  Service Instance GUID: "0ec2261c-5d50-c12e-4e8b-ca9273c6150f"
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

  Service Instance Name: "service-instance-3"
  Service Instance GUID: "ef7fa19f-0d66-55d0-0519-f198164d358c"
  Service Instance Version: "1.2.0"
  Service Plan Name: "service-plan-2"
  Service Plan GUID: "3ccc0ed1-1c06-036b-7bfe-f4d9dff25d02"
  Service Plan Version: "1.2.3"
  Service Offering Name: "service-offering-1"
  Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"

Number of service instances which failed to create: 1

  Service Instance Name: "service-instance-4"
  Service Instance GUID: "c53ccd0e-b88e-0d93-712d-609588651af0"
  Service Instance Version: "1.2.2"
  Service Plan Name: "service-plan-2"
  Service Plan GUID: "3ccc0ed1-1c06-036b-7bfe-f4d9dff25d02"
  Service Plan Version: "1.2.3"
  Service Offering Name: "service-offering-1"
  Service Offering GUID: "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa"
  Space Name: "fake-space"
  Space GUID: "5f870ea3-fa54-4174-ab3f-15f2d9516e07"
  Organization Name: "fake-org"
  Organization GUID: "1a2f43b5-1594-4247-a888-e8843ebd1b03"
`)))
		Expect(string(session.Err.Contents())).To(Equal("upgrade-all-services plugin failed: discovered service instances associated with deactivated plans or with an upgrade available"))
	})

	It("prints deactivated plans, pending upgrades and failed creates in JSON", func() {
		session := cf("upgrade-all-services", brokerName, "-check-up-to-date", "--json")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
		Expect(strings.TrimSpace(string(session.Out.Contents()))).To(MatchJSON(`
{
  "plan_deactivated": [
    {
      "guid": "ab5b7eb3-4d38-eb33-2aec-dbf7416d1db3",
      "maintenance_info": {
        "version": "1.2.3"
      },
      "name": "service-instance-5",
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
  ],
  "upgrade_pending": [
    {
      "guid": "0ec2261c-5d50-c12e-4e8b-ca9273c6150f",
      "maintenance_info": {
        "version": "1.2.2"
      },
      "name": "service-instance-2",
      "organization": {
        "guid": "1a2f43b5-1594-4247-a888-e8843ebd1b03",
        "name": "fake-org"
      },
      "service_offering": {
        "guid": "7fb1c0fc-45b4-fb4d-5aa5-2d2011573daa",
        "name": "service-offering-1"
      },
      "service_plan": {
        "guid": "173a3f22-e23f-27f2-9b32-8efdb64d5c14",
        "name": "service-plan1"
      },
      "space": {
        "guid": "5f870ea3-fa54-4174-ab3f-15f2d9516e07",
        "name": "fake-space"
      }
    },
    {
      "guid": "ef7fa19f-0d66-55d0-0519-f198164d358c",
      "maintenance_info": {
        "version": "1.2.0"
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
    }
  ],
  "create_failed": [
    {
      "guid": "c53ccd0e-b88e-0d93-712d-609588651af0",
      "maintenance_info": {
        "version": "1.2.2"
      },
      "name": "service-instance-4",
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
    }
  ]
}
`))
	})
})
