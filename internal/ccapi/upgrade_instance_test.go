package ccapi_test

import (
	"net/http"
	"time"

	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/requester"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("UpgradeServiceInstance", func() {
	const instanceUpdatingResponse = `
{
  "guid": "test-guid",
  "last_operation": {
    "type": "update",
    "state": "in progress",
    "description": "Update in progress"
  },
  "maintenance_info": {
    "version": "2.10.7-build.13"
  },
  "upgrade_available": false,
  "relationships": {
    "service_plan": {
      "data": {
        "guid": "3c994d0a-1ffa-4285-a88e-1a64cbc203c9"
      }
    }
  }
}
`
	const instanceSuccessResponse = `
{
  "guid": "test-guid",
  "last_operation": {
    "type": "update",
    "state": "succeeded",
    "description": "Instance update completed"
  },
  "maintenance_info": {
    "version": "2.10.7-build.13"
  },
  "upgrade_available": false,
  "relationships": {
    "service_plan": {
      "data": {
        "guid": "3c994d0a-1ffa-4285-a88e-1a64cbc203c9"
      }
    }
  }
}
`
	const instanceFailedResponse = `
{
  "guid": "test-guid",
  "last_operation": {
    "type": "update",
    "state": "failed",
    "description": "Instance update failed"
  },
  "maintenance_info": {
    "version": "2.10.7-build.13"
  },
  "upgrade_available": false,
  "relationships": {
    "service_plan": {
      "data": {
        "guid": "3c994d0a-1ffa-4285-a88e-1a64cbc203c9"
      }
    }
  }
}
`

	var (
		fakeServer  *ghttp.Server
		req         requester.Requester
		ccapiClient ccapi.CCAPI
	)

	BeforeEach(func() {
		fakeServer = ghttp.NewServer()
		DeferCleanup(fakeServer.Close)
		req = requester.NewRequester(fakeServer.URL(), "fake-token", false)
		ccapiClient = ccapi.NewCCAPI(req, time.Millisecond)
	})

	When("given an upgradeable instance", func() {
		BeforeEach(func() {
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("PATCH", "/v3/service_instances/test-guid"),
					ghttp.VerifyBody([]byte(`{"maintenance_info":{"version":"test-mi-version"}}`)),
					ghttp.RespondWith(http.StatusAccepted, ``, nil),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances/test-guid"),
					ghttp.RespondWith(http.StatusOK, instanceUpdatingResponse, nil),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances/test-guid"),
					ghttp.RespondWith(http.StatusOK, instanceSuccessResponse, nil),
				),
			)
		})

		It("successfully upgrades", func() {
			err := ccapiClient.UpgradeServiceInstance("test-guid", "test-mi-version")
			Expect(err).NotTo(HaveOccurred())

			requests := fakeServer.ReceivedRequests()
			Expect(requests).To(HaveLen(3))

			By("making the patch request")
			Expect(requests[0].Method).To(Equal("PATCH"))
			Expect(requests[0].URL.Path).To(Equal("/v3/service_instances/test-guid"))

			By("polling the service instance until complete")
			Expect(requests[1].Method).To(Equal("GET"))
			Expect(requests[1].URL.Path).To(Equal("/v3/service_instances/test-guid"))
			Expect(requests[2].Method).To(Equal("GET"))
			Expect(requests[2].URL.Path).To(Equal("/v3/service_instances/test-guid"))
		})
	})

	When("the upgrade request fails", func() {
		BeforeEach(func() {
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("PATCH", "/v3/service_instances/test-guid"),
					ghttp.RespondWith(http.StatusInternalServerError, ``, nil),
				),
			)
		})

		It("returns the error", func() {
			err := ccapiClient.UpgradeServiceInstance("test-guid", "test-mi-version")
			Expect(err).To(MatchError("upgrade request error: http_error: 500 Internal Server Error response_body: "))

			requests := fakeServer.ReceivedRequests()
			Expect(requests).To(HaveLen(1))

			By("making the patch request")
			Expect(requests[0].Method).To(Equal("PATCH"))
			Expect(requests[0].URL.Path).To(Equal("/v3/service_instances/test-guid"))
		})
	})

	When("the upgrade fails", func() {
		BeforeEach(func() {
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("PATCH", "/v3/service_instances/test-guid"),
					ghttp.VerifyBody([]byte(`{"maintenance_info":{"version":"test-mi-version"}}`)),
					ghttp.RespondWith(http.StatusAccepted, ``, nil),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances/test-guid"),
					ghttp.RespondWith(http.StatusOK, instanceUpdatingResponse, nil),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances/test-guid"),
					ghttp.RespondWith(http.StatusOK, instanceFailedResponse, nil),
				),
			)
		})

		It("returns the error", func() {
			err := ccapiClient.UpgradeServiceInstance("test-guid", "test-mi-version")
			Expect(err).To(MatchError("Instance update failed"))

			requests := fakeServer.ReceivedRequests()
			Expect(requests).To(HaveLen(3))

			By("making the patch request")
			Expect(requests[0].Method).To(Equal("PATCH"))
			Expect(requests[0].URL.Path).To(Equal("/v3/service_instances/test-guid"))

			By("polling the service instance until complete")
			Expect(requests[1].Method).To(Equal("GET"))
			Expect(requests[1].URL.Path).To(Equal("/v3/service_instances/test-guid"))
			Expect(requests[2].Method).To(Equal("GET"))
			Expect(requests[2].URL.Path).To(Equal("/v3/service_instances/test-guid"))
		})
	})

	DescribeTable("polling interval",
		func(interval time.Duration) {
			fakeServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("PATCH", "/v3/service_instances/test-guid"),
					ghttp.VerifyBody([]byte(`{"maintenance_info":{"version":"test-mi-version"}}`)),
					ghttp.RespondWith(http.StatusAccepted, ``, nil),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances/test-guid"),
					ghttp.RespondWith(http.StatusOK, instanceUpdatingResponse, nil),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyHeaderKV("Authorization", "fake-token"),
					ghttp.VerifyRequest("GET", "/v3/service_instances/test-guid"),
					ghttp.RespondWith(http.StatusOK, instanceSuccessResponse, nil),
				),
			)

			ccapiClient = ccapi.NewCCAPI(req, interval)

			const accuracy = 25 * time.Millisecond
			start := time.Now()
			Expect(ccapiClient.UpgradeServiceInstance("test-guid", "test-mi-version")).To(Succeed())
			Expect(time.Since(start)).To(BeNumerically("~", interval, accuracy), interval.String())
		},
		// essentially zero
		Entry("tiny interval", time.Millisecond),
		// big enough that we can see a change of behavior from above, but small enough that tests are still snappy
		Entry("moderate interval", 100*time.Millisecond),
	)
})
