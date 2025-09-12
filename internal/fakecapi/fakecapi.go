// Package fakecapi is a test fake for the Cloud Controller API
package fakecapi

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// fakeJWT is a valid JWT generated using http://jwt.io with a payload "exp" of 9999999999
const fakeJWT = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMiwiZXhwIjo5OTk5OTk5OTk5fQ.t4hnLA9gcRfu_W3_KpcS1UjCEl4xyDmEHUVODobhgAk`

func New() *FakeCAPI {
	loginPort, err := freePort()
	if err != nil {
		panic(err)
	}

	capiPort, err := freePort()
	if err != nil {
		panic(err)
	}

	f := FakeCAPI{
		URL:       fmt.Sprintf("http://localhost:%d", capiPort),
		loginURL:  fmt.Sprintf("http://localhost:%d", loginPort),
		brokers:   make(map[string]ServiceBroker),
		plans:     make(map[string]ServicePlan),
		offerings: make(map[string]ServiceOffering),
		instances: make(map[string]ServiceInstance),
	}

	f.stopLogin = start(loginMux(), loginPort)
	f.stopCAPI = start(f.capiMux(), capiPort)

	return &f
}

type FakeCAPI struct {
	URL       string
	loginURL  string
	stopLogin func()
	stopCAPI  func()
	brokers   map[string]ServiceBroker
	plans     map[string]ServicePlan
	offerings map[string]ServiceOffering
	instances map[string]ServiceInstance
}

func (f *FakeCAPI) Stop() {
	f.stopLogin()
	f.stopCAPI()
}

func (f *FakeCAPI) capiMux() *http.ServeMux {
	capi := http.NewServeMux()

	capi.HandleFunc("GET /v2/info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"api_version":"2.264.0"}`))
	})

	capi.HandleFunc("GET /v3/service_plans", f.servicePlanHandler())
	capi.HandleFunc("GET /v3/service_instances", f.serviceInstanceHandler())

	capi.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Write(fmt.Appendf(nil, `{"links":{"login":{"href":"%s"},"cloud_controller_v3":{"href":"%s/v3","meta":{"version":"3.199.0"}},"cloud_controller_v2":{"href":"%s/v2","meta":{"version":"2.264.0"}}}}`, f.loginURL, f.URL, f.URL))
			return
		}

		http.NotFound(w, r)
	})

	return capi
}

func loginMux() *http.ServeMux {
	login := http.NewServeMux()
	login.HandleFunc("POST /oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Write(fmt.Appendf(nil, `{"access_token":"%s"}`, fakeJWT))
	})

	return login
}

func start(handler http.Handler, port int) (stop func()) {
	svr := http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: handler,
	}

	go func() {
		if err := svr.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return func() {
		_ = svr.Shutdown(context.Background())
	}
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func guid() string {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}

	return fmt.Sprintf("%x-%x-%x-%x-%x", data[0:4], data[4:6], data[6:8], data[8:10], data[10:])
}

func filter[A any](a []A, cb func(A) bool) (result []A) {
	for _, v := range a {
		if cb(v) {
			result = append(result, v)
		}
	}
	return
}
