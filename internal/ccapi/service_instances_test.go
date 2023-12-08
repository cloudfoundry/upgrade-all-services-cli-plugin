package ccapi_test

import (
	"net/http"

	"upgrade-all-services-cli-plugin/internal/requester"

	"upgrade-all-services-cli-plugin/internal/ccapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("GetServiceInstances", func() {
	var (
		fakeServer *ghttp.Server
		req        requester.Requester
		fakeCCAPI  ccapi.CCAPI
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		DeferCleanup(fakeServer.Close)
		req = requester.NewRequester(fakeServer.URL(), "fake-token", false)
		fakeCCAPI = ccapi.NewCCAPI(req)
	})

	When("service instances exist in the given plans", func() {
		BeforeEach(func() {
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances", ccapi.BuildQueryParams([]string{
						"72abfc2f-5473-4fda-b895-a59d47b8f001",
						"e55b84e8-b953-4a14-98b2-67bec998a632",
						"510da794-1e71-4192-bd39-d974de20b7a4",
					})),
					ghttp.RespondWith(http.StatusOK, fakeResponse()),
				),
			)
		})

		It("returns instances from the given plans", func() {
			servicePlanOne := ccapi.ServicePlan{
				GUID:                   "72abfc2f-5473-4fda-b895-a59d47b8f001",
				Available:              true,
				Name:                   "db-small",
				MaintenanceInfoVersion: "1.5.1",
				ServiceOffering: ccapi.ServiceOffering{
					GUID: "707cff6a-fc54-471a-9594-442c306fb1d0",
					Name: "fake-service-offering-name-1",
				},
			}

			servicePlanTwo := ccapi.ServicePlan{
				GUID:                   "e55b84e8-b953-4a14-98b2-67bec998a632",
				Available:              true,
				Name:                   "postgres-db-f1-micro",
				MaintenanceInfoVersion: "",
				ServiceOffering: ccapi.ServiceOffering{
					GUID: "8551df49-1fb2-4d12-a009-5307176db52c",
					Name: "fake-service-offering-name-2",
				},
			}

			servicePlanThree := ccapi.ServicePlan{
				GUID:                   "510da794-1e71-4192-bd39-d974de20b7a4",
				Available:              true,
				Name:                   "small",
				MaintenanceInfoVersion: "",
				ServiceOffering: ccapi.ServiceOffering{
					GUID: "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd",
					Name: "fake-service-offering-name-3",
				},
			}
			actualInstances, err := fakeCCAPI.GetServiceInstancesForServicePlans([]ccapi.ServicePlan{servicePlanOne, servicePlanTwo, servicePlanThree})

			By("checking the valid service instance is returned")
			Expect(err).NotTo(HaveOccurred())

			Expect(actualInstances).To(ConsistOf(
				ccapi.ServiceInstance{
					GUID:                              "c5518540-7353-4d66-bae7-e07dfed8dd70",
					Name:                              "fake-service-instance-name-1",
					UpgradeAvailable:                  false,
					LastOperationType:                 "create",
					LastOperationState:                "succeeded",
					LastOperationDescription:          "Instance provisioning completed",
					MaintenanceInfoVersion:            "2.10.14-build.3",
					ServicePlanGUID:                   "72abfc2f-5473-4fda-b895-a59d47b8f001",
					ServicePlanName:                   "db-small",
					ServiceOfferingGUID:               "707cff6a-fc54-471a-9594-442c306fb1d0",
					ServiceOfferingName:               "fake-service-offering-name-1",
					SpaceGUID:                         "dcf1a90a-47ee-4ba2-a369-255aa00c2de0",
					SpaceName:                         "broker-cf-test",
					OrganizationGUID:                  "69086541-1b9d-449d-b8a4-79029b25e74f",
					OrganizationName:                  "pivotal",
					ServicePlanMaintenanceInfoVersion: "1.5.1",
				},
				ccapi.ServiceInstance{
					GUID:                              "3358305d-7402-48b3-80a7-e0148a38675b",
					Name:                              "fake-service-instance-name-2",
					UpgradeAvailable:                  false,
					LastOperationType:                 "create",
					LastOperationState:                "succeeded",
					LastOperationDescription:          "",
					MaintenanceInfoVersion:            "",
					ServicePlanGUID:                   "e55b84e8-b953-4a14-98b2-67bec998a632",
					ServicePlanName:                   "postgres-db-f1-micro",
					ServiceOfferingGUID:               "8551df49-1fb2-4d12-a009-5307176db52c",
					ServiceOfferingName:               "fake-service-offering-name-2",
					SpaceGUID:                         "dcf1a90a-47ee-4ba2-a369-255aa00c2de0",
					SpaceName:                         "broker-cf-test",
					OrganizationGUID:                  "69086541-1b9d-449d-b8a4-79029b25e74f",
					OrganizationName:                  "pivotal",
					ServicePlanMaintenanceInfoVersion: "", // Not in API response
				},
				ccapi.ServiceInstance{
					GUID:                              "5b528bf8-ac0f-4fed-85d0-0fb5f8588968",
					Name:                              "fake-service-instance-name-3",
					UpgradeAvailable:                  true,
					LastOperationType:                 "update",
					LastOperationState:                "succeeded",
					LastOperationDescription:          "update succeeded",
					MaintenanceInfoVersion:            "1.3.9",
					ServicePlanGUID:                   "510da794-1e71-4192-bd39-d974de20b7a4",
					ServicePlanName:                   "small",
					ServiceOfferingGUID:               "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd",
					ServiceOfferingName:               "fake-service-offering-name-3",
					SpaceGUID:                         "bbd00d42-8577-11ee-9b75-6feb4799d316",
					SpaceName:                         "broker-csb-test",
					OrganizationGUID:                  "529d3532-87a9-11ee-8a24-d354d25d7923",
					OrganizationName:                  "vmware",
					ServicePlanMaintenanceInfoVersion: "", // Not in API response
				},
			))

			requests := fakeServer.ReceivedRequests()
			Expect(requests).To(HaveLen(1))

			By("making the appending the plan guids")
			Expect(requests[0].Method).To(Equal("GET"))
			Expect(requests[0].URL.Path).To(Equal("/v3/service_instances"))
			Expect(requests[0].URL.RawQuery).To(Equal("per_page=5000&fields[space]=name,guid,relationships.organization&fields[space.organization]=name,guid&service_plan_guids=72abfc2f-5473-4fda-b895-a59d47b8f001,e55b84e8-b953-4a14-98b2-67bec998a632,510da794-1e71-4192-bd39-d974de20b7a4"))
		})
	})

	When("the request fails", func() {
		BeforeEach(func() {

			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.RespondWith(http.StatusInternalServerError, nil),
				),
			)
		})

		It("returns an error", func() {
			_, err := fakeCCAPI.GetServiceInstancesForServicePlans([]ccapi.ServicePlan{{GUID: "test-guid"}})

			Expect(err).To(MatchError("error getting service instances: http response: 500"))
		})
	})
})

func fakeResponse() string {
	return `
{
  "pagination": {
    "total_results": 3,
    "total_pages": 1,
    "first": {
      "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances?fields%5Bservice_plan%5D=name%2Cguid%2Crelationships.service_offering&fields%5Bservice_plan.service_offering%5D=guid%2Cname&fields%5Bspace%5D=name%2Cguid%2Crelationships.organization&fields%5Bspace.organization%5D=name%2Cguid&page=1&per_page=5000&service_plan_guids=72abfc2f-5473-4fda-b895-a59d47b8f001%2Ce55b84e8-b953-4a14-98b2-67bec998a632%2C510da794-1e71-4192-bd39-d974de20b7a4"
    },
    "last": {
      "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances?fields%5Bservice_plan%5D=name%2Cguid%2Crelationships.service_offering&fields%5Bservice_plan.service_offering%5D=guid%2Cname&fields%5Bspace%5D=name%2Cguid%2Crelationships.organization&fields%5Bspace.organization%5D=name%2Cguid&page=1&per_page=5000&service_plan_guids=72abfc2f-5473-4fda-b895-a59d47b8f001%2Ce55b84e8-b953-4a14-98b2-67bec998a632%2C510da794-1e71-4192-bd39-d974de20b7a4"
    },
    "next": null,
    "previous": null
  },
  "resources": [
    {
      "guid": "c5518540-7353-4d66-bae7-e07dfed8dd70",
      "created_at": "2023-11-16T23:23:09Z",
      "updated_at": "2023-11-16T23:23:13Z",
      "name": "fake-service-instance-name-1",
      "tags": [],
      "last_operation": {
        "type": "create",
        "state": "succeeded",
        "description": "Instance provisioning completed",
        "updated_at": "2023-11-16T23:27:41Z",
        "created_at": "2023-11-16T23:27:41Z"
      },
      "type": "managed",
      "maintenance_info": {
        "version": "2.10.14-build.3",
        "description": "MySQL(\"2.10.14-build.3\") for VMware Tanzu"
      },
      "upgrade_available": false,
      "dashboard_url": null,
      "relationships": {
        "space": {
          "data": {
            "guid": "dcf1a90a-47ee-4ba2-a369-255aa00c2de0"
          }
        },
        "service_plan": {
          "data": {
            "guid": "72abfc2f-5473-4fda-b895-a59d47b8f001"
          }
        }
      },
      "metadata": {
        "labels": {},
        "annotations": {}
      },
      "links": {
        "self": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/c5518540-7353-4d66-bae7-e07dfed8dd70"
        },
        "space": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/spaces/dcf1a90a-47ee-4ba2-a369-255aa00c2de0"
        },
        "service_credential_bindings": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_credential_bindings?service_instance_guids=c5518540-7353-4d66-bae7-e07dfed8dd70"
        },
        "service_route_bindings": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_route_bindings?service_instance_guids=c5518540-7353-4d66-bae7-e07dfed8dd70"
        },
        "service_plan": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_plans/72abfc2f-5473-4fda-b895-a59d47b8f001"
        },
        "parameters": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/c5518540-7353-4d66-bae7-e07dfed8dd70/parameters"
        },
        "shared_spaces": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/c5518540-7353-4d66-bae7-e07dfed8dd70/relationships/shared_spaces"
        }
      }
    },
    {
      "guid": "3358305d-7402-48b3-80a7-e0148a38675b",
      "created_at": "2023-11-17T11:12:37Z",
      "updated_at": "2023-11-17T11:12:41Z",
      "name": "fake-service-instance-name-2",
      "tags": [],
      "last_operation": {
        "type": "create",
        "state": "succeeded",
        "description": null,
        "updated_at": "2023-11-17T11:20:24Z",
        "created_at": "2023-11-17T11:20:24Z"
      },
      "type": "managed",
      "maintenance_info": {},
      "upgrade_available": false,
      "dashboard_url": null,
      "relationships": {
        "space": {
          "data": {
            "guid": "dcf1a90a-47ee-4ba2-a369-255aa00c2de0"
          }
        },
        "service_plan": {
          "data": {
            "guid": "e55b84e8-b953-4a14-98b2-67bec998a632"
          }
        }
      },
      "metadata": {
        "labels": {},
        "annotations": {}
      },
      "links": {
        "self": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/3358305d-7402-48b3-80a7-e0148a38675b"
        },
        "space": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/spaces/dcf1a90a-47ee-4ba2-a369-255aa00c2de0"
        },
        "service_credential_bindings": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_credential_bindings?service_instance_guids=3358305d-7402-48b3-80a7-e0148a38675b"
        },
        "service_route_bindings": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_route_bindings?service_instance_guids=3358305d-7402-48b3-80a7-e0148a38675b"
        },
        "service_plan": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_plans/e55b84e8-b953-4a14-98b2-67bec998a632"
        },
        "parameters": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/3358305d-7402-48b3-80a7-e0148a38675b/parameters"
        },
        "shared_spaces": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/3358305d-7402-48b3-80a7-e0148a38675b/relationships/shared_spaces"
        }
      }
    },
    {
      "guid": "5b528bf8-ac0f-4fed-85d0-0fb5f8588968",
      "created_at": "2023-11-17T11:46:11Z",
      "updated_at": "2023-11-17T11:46:16Z",
      "name": "fake-service-instance-name-3",
      "tags": [],
      "last_operation": {
        "type": "update",
        "state": "succeeded",
        "description": "update succeeded",
        "updated_at": "2023-11-17T13:56:14Z",
        "created_at": "2023-11-17T13:56:14Z"
      },
      "type": "managed",
      "maintenance_info": {
        "version": "1.3.9",
        "description": "This upgrade provides support for Terraform version: 1.3.9. The upgrade operation will take a while. The instance and all associated bindings will be upgraded."
      },
      "upgrade_available": true,
      "dashboard_url": null,
      "relationships": {
        "space": {
          "data": {
            "guid": "bbd00d42-8577-11ee-9b75-6feb4799d316"
          }
        },
        "service_plan": {
          "data": {
            "guid": "510da794-1e71-4192-bd39-d974de20b7a4"
          }
        }
      },
      "metadata": {
        "labels": {},
        "annotations": {}
      },
      "links": {
        "self": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/5b528bf8-ac0f-4fed-85d0-0fb5f8588968"
        },
        "space": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/spaces/dcf1a90a-47ee-4ba2-a369-255aa00c2de0"
        },
        "service_credential_bindings": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_credential_bindings?service_instance_guids=5b528bf8-ac0f-4fed-85d0-0fb5f8588968"
        },
        "service_route_bindings": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_route_bindings?service_instance_guids=5b528bf8-ac0f-4fed-85d0-0fb5f8588968"
        },
        "service_plan": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_plans/510da794-1e71-4192-bd39-d974de20b7a4"
        },
        "parameters": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/5b528bf8-ac0f-4fed-85d0-0fb5f8588968/parameters"
        },
        "shared_spaces": {
          "href": "https://api.sys.mycfname.cf-app.com/v3/service_instances/5b528bf8-ac0f-4fed-85d0-0fb5f8588968/relationships/shared_spaces"
        }
      }
    }
  ],
  "included": {
    "spaces": [
      {
        "guid": "dcf1a90a-47ee-4ba2-a369-255aa00c2de0",
        "name": "broker-cf-test",
        "relationships": {
          "organization": {
            "data": {
              "guid": "69086541-1b9d-449d-b8a4-79029b25e74f"
            }
          }
        }
      },
      {
        "guid": "bbd00d42-8577-11ee-9b75-6feb4799d316",
        "name": "broker-csb-test",
        "relationships": {
          "organization": {
            "data": {
              "guid": "529d3532-87a9-11ee-8a24-d354d25d7923"
            }
          }
        }
      }
    ],
    "organizations": [
      {
        "name": "pivotal",
        "guid": "69086541-1b9d-449d-b8a4-79029b25e74f"
      },
      {
        "name": "vmware",
        "guid": "529d3532-87a9-11ee-8a24-d354d25d7923"
      }
    ],
    "service_plans": [
      {
        "guid": "72abfc2f-5473-4fda-b895-a59d47b8f001",
        "name": "db-small",
        "relationships": {
          "service_offering": {
            "data": {
              "guid": "707cff6a-fc54-471a-9594-442c306fb1d0"
            }
          }
        }
      },
      {
        "guid": "e55b84e8-b953-4a14-98b2-67bec998a632",
        "name": "postgres-db-f1-micro",
        "relationships": {
          "service_offering": {
            "data": {
              "guid": "8551df49-1fb2-4d12-a009-5307176db52c"
            }
          }
        }
      },
      {
        "guid": "510da794-1e71-4192-bd39-d974de20b7a4",
        "name": "small",
        "relationships": {
          "service_offering": {
            "data": {
              "guid": "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd"
            }
          }
        }
      }
    ],
    "service_offerings": [
      {
        "name": "fake-service-offering-name-1",
        "guid": "707cff6a-fc54-471a-9594-442c306fb1d0"
      },
      {
        "name": "fake-service-offering-name-2",
        "guid": "8551df49-1fb2-4d12-a009-5307176db52c"
      },
      {
        "name": "fake-service-offering-name-3",
        "guid": "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd"
      }
    ]
  }
}
`
}

// [FAILED] Expected
// <[]ccapi.ServiceInstance | len:3, cap:3>: [

// {
// GUID: "5b528bf8-ac0f-4fed-85d0-0fb5f8588968",
// Name: "fake-service-instance-name-3",
// UpgradeAvailable: true,
// ServicePlanGUID: "510da794-1e71-4192-bd39-d974de20b7a4",
// SpaceGUID: "bbd00d42-8577-11ee-9b75-6feb4799d316",
// LastOperationType: "update",
// LastOperationState: "succeeded",
// LastOperationDescription: "update succeeded",
// MaintenanceInfoVersion: "1.3.9",
// ServicePlanName: "small",
// ServiceOfferingGUID: "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd",
// ServiceOfferingName: "fake-service-offering-name-3",
// SpaceName: "broker-csb-test",
// OrganizationGUID: "529d3532-87a9-11ee-8a24-d354d25d7923",
// OrganizationName: "vmware",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: true,
// },
// ]

// {
// GUID: "3358305d-7402-48b3-80a7-e0148a38675b",
// Name: "fake-service-instance-name-2",
// UpgradeAvailable: false,
// ServicePlanGUID: "e55b84e8-b953-4a14-98b2-67bec998a632",
// SpaceGUID: "dcf1a90a-47ee-4ba2-a369-255aa00c2de0",
// LastOperationType: "create",
// LastOperationState: "succeeded",
// LastOperationDescription: "",
// MaintenanceInfoVersion: "",
// ServicePlanName: "postgres-db-f1-micro",
// ServiceOfferingGUID: "8551df49-1fb2-4d12-a009-5307176db52c",
// ServiceOfferingName: "fake-service-offering-name-2",
// SpaceName: "broker-cf-test",
// OrganizationGUID: "69086541-1b9d-449d-b8a4-79029b25e74f",
// OrganizationName: "pivotal",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: true,
// },

// {
// GUID: "3358305d-7402-48b3-80a7-e0148a38675b",
// Name: "fake-service-instance-name-2",
// UpgradeAvailable: false,
// ServicePlanGUID: "e55b84e8-b953-4a14-98b2-67bec998a632",
// SpaceGUID: "dcf1a90a-47ee-4ba2-a369-255aa00c2de0",
// LastOperationType: "create",
// LastOperationState: "succeeded",
// LastOperationDescription: "",
// MaintenanceInfoVersion: "",
// ServicePlanName: "postgres-db-f1-micro",
// ServiceOfferingGUID: "8551df49-1fb2-4d12-a009-5307176db52c",
// ServiceOfferingName: "fake-service-offering-name-2",
// SpaceName: "broker-cf-test",
// OrganizationGUID: "69086541-1b9d-449d-b8a4-79029b25e74f",
// OrganizationName: "pivotal",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: false,
// },

// to consist of
// <[]ccapi.ServiceInstance | len:3, cap:3>: [

// {
// GUID: "5b528bf8-ac0f-4fed-85d0-0fb5f8588968",
// Name: "fake-service-instance-name-3",
// UpgradeAvailable: true,
// ServicePlanGUID: "510da794-1e71-4192-bd39-d974de20b7a4",
// SpaceGUID: "bbd00d42-8577-11ee-9b75-6feb4799d316",
// LastOperationType: "update",
// LastOperationState: "succeeded",
// LastOperationDescription: "update succeeded",
// MaintenanceInfoVersion: "1.3.9",
// ServicePlanName: "small",
// ServiceOfferingGUID: "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd",
// ServiceOfferingName: "fake-service-offering-name-3",
// SpaceName: "broker-csb-test",
// OrganizationGUID: "529d3532-87a9-11ee-8a24-d354d25d7923",
// OrganizationName: "vmware",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: false,
// },
// ]
// the missing elements were
// <[]ccapi.ServiceInstance | len:2, cap:2>: [
// {
// GUID: "3358305d-7402-48b3-80a7-e0148a38675b",
// Name: "fake-service-instance-name-2",
// UpgradeAvailable: false,
// ServicePlanGUID: "e55b84e8-b953-4a14-98b2-67bec998a632",
// SpaceGUID: "dcf1a90a-47ee-4ba2-a369-255aa00c2de0",
// LastOperationType: "create",
// LastOperationState: "succeeded",
// LastOperationDescription: "",
// MaintenanceInfoVersion: "",
// ServicePlanName: "postgres-db-f1-micro",
// ServiceOfferingGUID: "8551df49-1fb2-4d12-a009-5307176db52c",
// ServiceOfferingName: "fake-service-offering-name-2",
// SpaceName: "broker-cf-test",
// OrganizationGUID: "69086541-1b9d-449d-b8a4-79029b25e74f",
// OrganizationName: "pivotal",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: false,
// },
// {
// GUID: "5b528bf8-ac0f-4fed-85d0-0fb5f8588968",
// Name: "fake-service-instance-name-3",
// UpgradeAvailable: true,
// ServicePlanGUID: "510da794-1e71-4192-bd39-d974de20b7a4",
// SpaceGUID: "bbd00d42-8577-11ee-9b75-6feb4799d316",
// LastOperationType: "update",
// LastOperationState: "succeeded",
// LastOperationDescription: "update succeeded",
// MaintenanceInfoVersion: "1.3.9",
// ServicePlanName: "small",
// ServiceOfferingGUID: "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd",
// ServiceOfferingName: "fake-service-offering-name-3",
// SpaceName: "broker-csb-test",
// OrganizationGUID: "529d3532-87a9-11ee-8a24-d354d25d7923",
// OrganizationName: "vmware",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: false,
// },
// ]
// the extra elements were
// <[]ccapi.ServiceInstance | len:2, cap:2>: [
// {
// GUID: "3358305d-7402-48b3-80a7-e0148a38675b",
// Name: "fake-service-instance-name-2",
// UpgradeAvailable: false,
// ServicePlanGUID: "e55b84e8-b953-4a14-98b2-67bec998a632",
// SpaceGUID: "dcf1a90a-47ee-4ba2-a369-255aa00c2de0",
// LastOperationType: "create",
// LastOperationState: "succeeded",
// LastOperationDescription: "",
// MaintenanceInfoVersion: "",
// ServicePlanName: "postgres-db-f1-micro",
// ServiceOfferingGUID: "8551df49-1fb2-4d12-a009-5307176db52c",
// ServiceOfferingName: "fake-service-offering-name-2",
// SpaceName: "broker-cf-test",
// OrganizationGUID: "69086541-1b9d-449d-b8a4-79029b25e74f",
// OrganizationName: "pivotal",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: true,
// },
// {
// GUID: "5b528bf8-ac0f-4fed-85d0-0fb5f8588968",
// Name: "fake-service-instance-name-3",
// UpgradeAvailable: true,
// ServicePlanGUID: "510da794-1e71-4192-bd39-d974de20b7a4",
// SpaceGUID: "bbd00d42-8577-11ee-9b75-6feb4799d316",
// LastOperationType: "update",
// LastOperationState: "succeeded",
// LastOperationDescription: "update succeeded",
// MaintenanceInfoVersion: "1.3.9",
// ServicePlanName: "small",
// ServiceOfferingGUID: "ebdddfd4-c95a-4e1a-bdd1-4697ffb57fcd",
// ServiceOfferingName: "fake-service-offering-name-3",
// SpaceName: "broker-csb-test",
// OrganizationGUID: "529d3532-87a9-11ee-8a24-d354d25d7923",
// OrganizationName: "vmware",
// ServicePlanMaintenanceInfoVersion: "",
// ServicePlanDeactivated: true,
// },
// ]
