package integrationtests_test

import (
	"os/exec"
	"testing"
	"time"
	"upgrade-all-services-cli-plugin/internal/fakecapi"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var (
	cfPath string
	capi   *fakecapi.FakeCAPI
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Tests Suite")
}

var _ = SynchronizedBeforeSuite(
	func() []byte {
		cf, err := exec.LookPath("cf")
		Expect(err).NotTo(HaveOccurred(), "The 'cf' executable is a pre-requisite for running this test")

		plugin, err := Build("upgrade-all-services-cli-plugin")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(CleanupBuildArtifacts)

		session, err := Start(exec.Command(cf, "install-plugin", "-f", plugin), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).WithTimeout(time.Minute).Should(Exit(0))

		return []byte(cf)
	},
	func(input []byte) {
		cfPath = string(input)
	},
)

var _ = BeforeEach(func() {
	capi = fakecapi.New()
	DeferCleanup(func() {
		capi.Stop()
	})

	Eventually(cf("api", "--skip-ssl-validation", capi.URL)).WithTimeout(time.Minute).Should(Exit(0))
	Eventually(cf("auth", "foo", "bar")).WithTimeout(time.Minute).Should(Exit(0))
})

func cf(args ...string) *Session {
	cmd := exec.Command(cfPath, args...)
	session, err := Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
