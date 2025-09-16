package integrationtests_test

import (
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
	"upgrade-all-services-cli-plugin/internal/fakecapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	cf   func(...string) *Session
	capi *fakecapi.FakeCAPI
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Tests Suite")
}

var _ = SynchronizedBeforeSuite(
	// Test concurrency in Ginkgo works by using OS processes. This first function is run on process #1 and the result
	// is passed to the second function. Note that any global variable set in this first function will only be available
	// in process #1, which can result in confusing behavior. So to avoid confusion we don't set any variables.
	// We just pass data via the function return value.
	func() []byte {
		cfPath, err := exec.LookPath("cf")
		Expect(err).NotTo(HaveOccurred(), "The 'cf' executable is a pre-requisite for running this test")

		pluginPath, err := Build("upgrade-all-services-cli-plugin")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(CleanupBuildArtifacts)

		return encode(cfPath, pluginPath)
	},
	// This second function is run on all test processes, so any global variables set here are available in all processes.
	// The input to this function is the output from the first function.
	func(data []byte) {
		cfPath, pluginPath := decode(data)

		// The CF CLI does not concurrently handle installation of plugins and login/auth because under the hood the
		// different instances write to the same file. Unless we use different home directories for each instance,
		// in which case it all works.
		homePath := GinkgoT().TempDir()

		cf = func(args ...string) *Session {
			cmd := exec.Command(cfPath, args...)
			cmd.Env = append(
				os.Environ(),
				fmt.Sprintf("HOME=%s", homePath),
			)
			session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			return session
		}

		session := cf("install-plugin", "-f", pluginPath)
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))

		capi = fakecapi.New()
		DeferCleanup(func() {
			capi.Stop()
		})

		Eventually(cf("api", "--skip-ssl-validation", capi.URL)).WithTimeout(time.Minute).Should(Exit(0))
		Eventually(cf("auth", "foo", "bar")).WithTimeout(time.Minute).Should(Exit(0))
	},
)

var _ = BeforeEach(func() {
	capi.Reset()
})

// encode safely encodes two strings as a []byte
func encode(cfPath, pluginPath string) []byte {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, uint32(len(cfPath)))
	data = fmt.Append(data, cfPath)
	data = fmt.Append(data, pluginPath)

	return data
}

// decode safely decodes two strings from a []byte
func decode(data []byte) (cfPath, pluginPath string) {
	length := binary.LittleEndian.Uint32(data[:4])
	cfPath = string(data[4 : 4+length])
	pluginPath = string(data[4+length:])
	return
}
