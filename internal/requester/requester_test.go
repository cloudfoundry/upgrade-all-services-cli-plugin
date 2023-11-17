package requester_test

import (
	"math"
	"net/http"

	"upgrade-all-services-cli-plugin/internal/requester"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Requester", func() {
	var (
		fakeRequester requester.Requester
		fakeServer    *ghttp.Server
		testReceiver  map[string]interface{}
	)

	BeforeEach(func() {
		testReceiver = map[string]interface{}{}

		fakeServer = ghttp.NewServer()
		DeferCleanup(fakeServer.Close)

		fakeRequester = requester.NewRequester(fakeServer.URL(), "fake-token", false)
	})

	Describe("Get", func() {
		When("request is valid", func() {
			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Authorization", "fake-token"),
						ghttp.VerifyRequest("GET", "/test-endpoint", ""),
						ghttp.RespondWith(http.StatusOK, `{"test_value": "foo"}`, nil),
					),
				)
			})

			It("succeeds", func() {
				err := fakeRequester.Get("test-endpoint", &testReceiver)

				Expect(err).NotTo(HaveOccurred())
				Expect(testReceiver).To(Equal(map[string]interface{}{"test_value": "foo"}))
			})
		})

		When("request is invalid", func() {
			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Authorization", "fake-token"),
						ghttp.VerifyRequest("GET", "/not-a-real-url", ""),
						ghttp.RespondWith(http.StatusNotFound, "", nil),
					),
				)
			})

			It("returns an error", func() {
				err := fakeRequester.Get("not-a-real-url", &testReceiver)
				Expect(err).To(MatchError("http response: 404"))
			})
		})

		When("unable to parse response body as JSON", func() {
			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Authorization", "fake-token"),
						ghttp.VerifyRequest("GET", "/test-endpoint", ""),
						ghttp.RespondWith(http.StatusOK, ``, nil),
					),
				)
			})

			It("returns an error", func() {
				err := fakeRequester.Get("test-endpoint", &testReceiver)
				Expect(err).To(MatchError("failed to unmarshal response into receiver error: unexpected end of JSON input"))
			})
		})

		When("passed a receiver that is not a pointer", func() {
			It("returns an error", func() {
				err := fakeRequester.Get("test-endpoint", testReceiver)
				Expect(err).To(MatchError("receiver must be of type Pointer"))
			})
		})
	})

	Describe("Patch", func() {
		When("request is valid", func() {
			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Authorization", "fake-token"),
						ghttp.VerifyRequest("PATCH", "/test-endpoint", ""),
						ghttp.RespondWith(http.StatusAccepted, ``, nil),
					),
				)
			})

			It("succeeds", func() {
				err := fakeRequester.Patch("test-endpoint", `data`)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeServer.ReceivedRequests()).To(HaveLen(1))
			})
		})

		When("the patch request fails", func() {
			When("fails with unexpected error", func() {
				BeforeEach(func() {
					fakeServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyHeaderKV("Authorization", "fake-token"),
							ghttp.VerifyRequest("PATCH", "/test-endpoint", ""),
							ghttp.RespondWith(http.StatusInternalServerError, `Some body`, nil),
						),
					)
				})

				It("returns an error", func() {
					err := fakeRequester.Patch("test-endpoint", `data`)
					Expect(err).To(MatchError("http_error: 500 Internal Server Error response_body: Some body"))
				})
			})

			When("fails with capi error", func() {
				BeforeEach(func() {
					fakeServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyHeaderKV("Authorization", "fake-token"),
							ghttp.VerifyRequest("PATCH", "/test-endpoint", ""),
							ghttp.RespondWith(http.StatusInternalServerError, `{"errors": [{"code":10008, "title":"error title", "detail":"error detail"}]}`, nil),
						),
					)
				})

				It("returns an error", func() {
					err := fakeRequester.Patch("test-endpoint", `data`)
					Expect(err).To(MatchError("http_error: 500 Internal Server Error capi_error_code: 10008 capi_error_title: error title capi_error_detail: error detail"))
				})
			})

			When("fails with multiple capi errors", func() {
				BeforeEach(func() {
					fakeServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyHeaderKV("Authorization", "fake-token"),
							ghttp.VerifyRequest("PATCH", "/test-endpoint", ""),
							ghttp.RespondWith(http.StatusInternalServerError, `{"errors": [{"code":10008, "title":"error title", "detail":"error detail"}, {"code":10009, "title":"other error title", "detail":"other error detail"}]}`, nil),
						),
					)
				})

				It("returns an error", func() {
					err := fakeRequester.Patch("test-endpoint", `data`)
					Expect(err.Error()).To(ContainSubstring("http_error: 500 Internal Server Error"))
					Expect(err.Error()).To(ContainSubstring("capi_error_code: 10008 capi_error_title: error title capi_error_detail: error detail"))
					Expect(err.Error()).To(ContainSubstring("capi_error_code: 10009 capi_error_title: other error title capi_error_detail: other error detail"))
				})
			})
		})

		It("errors if data can not be marshalled", func() {
			err := fakeRequester.Patch("test-endpoint", math.Inf(1))
			Expect(err).To(MatchError("error marshaling data: json: unsupported value: +Inf"))
		})
	})
})
