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
					fakecapi.ServicePlan{Name: "service-plan1", Version: "1.2.3", Available: true},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-1", UpgradeAvailable: true, Version: "1.2.2"},
						fakecapi.ServiceInstance{Name: "service-instance-2", UpgradeAvailable: false, Version: "1.2.3"},
					),
				),
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-2", Version: "1.2.0", Available: true},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-3", UpgradeAvailable: true, Version: "1.1.0"},
						fakecapi.ServiceInstance{Name: "service-instance-4", UpgradeAvailable: true, Version: "1.1.0", LastOperationType: "create", LastOperationState: "failed"},
					),
				),
			),
			fakecapi.WithServiceOffering(
				fakecapi.ServiceOffering{Name: "service-offering-2"},
				fakecapi.WithServicePlan(
					fakecapi.ServicePlan{Name: "service-plan-3", Version: "1.3.0", Available: false},
					fakecapi.WithServiceInstances(
						fakecapi.ServiceInstance{Name: "service-instance-5", UpgradeAvailable: true, Version: "1.2.9"},
					),
				),
			),
		)
	})

	It("shows which service instances would be upgraded as text output", func() {
		session := cf("upgrade-all-services", brokerName, "-dry-run")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
		Expect(session.Out).To(Say(strings.TrimSpace(`
\S+: discovering service instances for broker: dry-run-broker
\S+: skipping instance: "service-instance-4" guid: "c53ccd0e-b88e-0d93-712d-609588651af0" Upgrade Available: true Last Operation Type: "create" State: "failed"
\S+: ---
\S+: total instances: 5
\S+: upgradable instances: 3
\S+: ---
\S+: starting upgrade...
\S+: upgrade of instance: "service-instance-1" guid: "5cc87b43-f885-3b94-328f-8a5f953590d3" failed after 0s: dry-run prevented upgrade instance guid 5cc87b43-f885-3b94-328f-8a5f953590d3
\S+: upgrade of instance: "service-instance-3" guid: "ef7fa19f-0d66-55d0-0519-f198164d358c" failed after 0s: dry-run prevented upgrade instance guid ef7fa19f-0d66-55d0-0519-f198164d358c
\S+: upgrade of instance: "service-instance-5" guid: "ab5b7eb3-4d38-eb33-2aec-dbf7416d1db3" failed after 0s: dry-run prevented upgrade instance guid ab5b7eb3-4d38-eb33-2aec-dbf7416d1db3
\S+: upgraded 0 of 3
\S+: ---
\S+: skipped 1 instances
\S+: successfully upgraded 0 instances
\S+: failed to upgrade 3 instances
\S+: 

  Details: "dry-run prevented upgrade instance guid 5cc87b43-f885-3b94-328f-8a5f953590d3"
  Service Instance Name: "service-instance-1"
  Service Instance GUID: "5cc87b43-f885-3b94-328f-8a5f953590d3"
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

  Details: "dry-run prevented upgrade instance guid ef7fa19f-0d66-55d0-0519-f198164d358c"
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

  Details: "dry-run prevented upgrade instance guid ab5b7eb3-4d38-eb33-2aec-dbf7416d1db3"
  Service Instance Name: "service-instance-5"
  Service Instance GUID: "ab5b7eb3-4d38-eb33-2aec-dbf7416d1db3"
  Service Instance Version: "1.2.9"
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
	})

	It("shows which service instances would be upgraded as JSON", func() {
		session := cf("upgrade-all-services", brokerName, "-dry-run", "-json")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`
  {
    "upgrade": [
      {
        "guid": "5cc87b43-f885-3b94-328f-8a5f953590d3",
        "maintenance_info": {
          "version": "1.2.2"
        },
        "name": "service-instance-1",
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
        "guid": "ab5b7eb3-4d38-eb33-2aec-dbf7416d1db3",
        "maintenance_info": {
          "version": "1.2.9"
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
    "skip": [
      {
        "guid": "c53ccd0e-b88e-0d93-712d-609588651af0",
        "maintenance_info": {
          "version": "1.1.0"
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
  }`))
	})

	It("accepts a limit", func() {
		session := cf("upgrade-all-services", brokerName, "-dry-run", "-json", "-limit", "2")
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))
		Expect(string(session.Out.Contents())).To(MatchJSON(`
  {
    "upgrade": [
      {
        "guid": "5cc87b43-f885-3b94-328f-8a5f953590d3",
        "maintenance_info": {
          "version": "1.2.2"
        },
        "name": "service-instance-1",
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
      }
    ],
    "skip": [
      {
        "guid": "c53ccd0e-b88e-0d93-712d-609588651af0",
        "maintenance_info": {
          "version": "1.1.0"
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
  }`))
	})

})
