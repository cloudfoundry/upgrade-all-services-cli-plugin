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
		testRequester requester.Requester
		fakeServer    *ghttp.Server
		testReceiver  struct {
			TestValue string `json:"test_value"`
		}
	)

	BeforeEach(func() {
		testReceiver.TestValue = ""

		fakeServer = ghttp.NewServer()
		DeferCleanup(fakeServer.Close)

		testRequester = requester.NewRequester(fakeServer.URL(), "fake-token", false)
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
				err := testRequester.Get("test-endpoint", &testReceiver)

				Expect(err).NotTo(HaveOccurred())
				Expect(testReceiver.TestValue).To(Equal("foo"))
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
				err := testRequester.Get("not-a-real-url", &testReceiver)
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
				err := testRequester.Get("test-endpoint", &testReceiver)
				Expect(err).To(MatchError("failed to unmarshal response into receiver error: error parsing JSON: EOF"))
			})
		})

		When("passed a receiver that is not a pointer", func() {
			It("returns an error", func() {
				err := testRequester.Get("test-endpoint", testReceiver)
				Expect(err).To(MatchError("receiver must be a pointer to a struct, got non-pointer"))
			})
		})

		When("passed a receiver that is a pointer to a non-struct", func() {
			It("returns an error", func() {
				var s string
				err := testRequester.Get("test-endpoint", &s)
				Expect(err).To(MatchError("receiver must be a pointer to a struct, got non-struct"))
			})
		})
	})

	Describe("Patch", func() {
		var testBody struct {
			Data string `json:"data"`
		}

		BeforeEach(func() {
			testBody.Data = "bar"
		})

		When("request is valid", func() {

			BeforeEach(func() {
				fakeServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyHeaderKV("Authorization", "fake-token"),
						ghttp.VerifyRequest("PATCH", "/test-endpoint", ""),
						ghttp.VerifyBody([]byte(`{"data":"bar"}`)),
						ghttp.RespondWith(http.StatusAccepted, ``, nil),
					),
				)
			})

			It("succeeds", func() {
				err := testRequester.Patch("test-endpoint", testBody)
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
					err := testRequester.Patch("test-endpoint", testBody)
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
					err := testRequester.Patch("test-endpoint", testBody)
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
					err := testRequester.Patch("test-endpoint", testBody)
					Expect(err.Error()).To(ContainSubstring("http_error: 500 Internal Server Error"))
					Expect(err.Error()).To(ContainSubstring("capi_error_code: 10008 capi_error_title: error title capi_error_detail: error detail"))
					Expect(err.Error()).To(ContainSubstring("capi_error_code: 10009 capi_error_title: other error title capi_error_detail: other error detail"))
				})
			})
		})

		It("errors if body is not a struct", func() {
			err := testRequester.Patch("test-endpoint", math.Inf(1))
			Expect(err).To(MatchError("input body must be a struct"))
		})

		It("errors if body can not be marshalled", func() {
			input := struct {
				Data func()
			}{}
			err := testRequester.Patch("test-endpoint", input)
			Expect(err).To(MatchError(`error marshaling data: unsupported type "func()" at field "Data" (type "func()")`))
		})
	})
})
