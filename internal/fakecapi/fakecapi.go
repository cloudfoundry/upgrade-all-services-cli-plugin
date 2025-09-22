// Package fakecapi is a test fake for the Cloud Controller API
package fakecapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
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
		URL:      fmt.Sprintf("http://localhost:%d", capiPort),
		loginURL: fmt.Sprintf("http://localhost:%d", loginPort),
	}

	f.Reset()
	f.stopLogin = start(loginMux(), loginPort)
	f.stopCAPI = start(f.capiMux(), capiPort)

	return &f
}

type FakeCAPI struct {
	URL                     string
	loginURL                string
	stopLogin               func()
	stopCAPI                func()
	brokers                 map[string]ServiceBroker
	plans                   map[string]ServicePlan
	offerings               map[string]ServiceOffering
	instances               map[string]*ServiceInstance
	fakeNameCount           map[string]int
	lock                    sync.Mutex
	concurrentOperations    int
	MaxConcurrentOperations int
}

func (f *FakeCAPI) Reset() {
	f.brokers = make(map[string]ServiceBroker)
	f.plans = make(map[string]ServicePlan)
	f.offerings = make(map[string]ServiceOffering)
	f.instances = make(map[string]*ServiceInstance)
	f.fakeNameCount = make(map[string]int)
	f.concurrentOperations = 0
	f.MaxConcurrentOperations = 0
}

func (f *FakeCAPI) Stop() {
	f.stopLogin()
	f.stopCAPI()
}

// UpdateCount sums the number of updates completed
func (f *FakeCAPI) UpdateCount() (result int) {
	for _, v := range f.instances {
		result += v.UpdateCount
	}
	return
}

func (f *FakeCAPI) capiMux() *http.ServeMux {
	capi := http.NewServeMux()

	capi.HandleFunc("GET /v2/info", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"api_version":"2.264.0"}`))
	})

	capi.HandleFunc("GET /v3/service_plans", f.listServicePlansHandler())
	capi.HandleFunc("GET /v3/service_instances", f.listServiceInstancesHandler())
	capi.HandleFunc("GET /v3/service_instances/{guid}", f.getServiceInstanceHandler())
	capi.HandleFunc("PATCH /v3/service_instances/{guid}", f.updateServiceInstanceHandler())

	capi.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Write(fmt.Appendf(nil, `{"links":{"login":{"href":"%s"},"cloud_controller_v3":{"href":"%s/v3","meta":{"version":"3.199.0"}},"cloud_controller_v2":{"href":"%s/v2","meta":{"version":"2.264.0"}}}}`, f.loginURL, f.URL, f.URL))
			return
		}

		http.NotFound(w, r)
	})

	return capi
}

func (f *FakeCAPI) startOperation() {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.concurrentOperations++
	if f.concurrentOperations > f.MaxConcurrentOperations {
		f.MaxConcurrentOperations = f.concurrentOperations
	}
}

func (f *FakeCAPI) stopOperation() {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.concurrentOperations--
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

	// Wait for the server to actually start
	expiry := time.Now().Add(10 * time.Second)
	for !ping(port) {
		if time.Now().After(expiry) {
			panic(fmt.Sprintf("timed out waiting for ping on port %d", port))
		}
		time.Sleep(50 * time.Millisecond)
	}

	return func() {
		_ = svr.Shutdown(context.Background())
	}
}

func ping(port int) bool {
	_, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
	return err == nil
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
