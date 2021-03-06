package logger_test

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"upgrade-all-services-cli-plugin/internal/logger"
	"upgrade-all-services-cli-plugin/internal/upgrader"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ upgrader.Logger = &logger.Logger{}

var _ = Describe("Logger", func() {
	const timestampRegexp = `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}([+-]\d{2}:\d{2}|Z)`

	var l *logger.Logger

	BeforeEach(func() {
		l = logger.New(100 * time.Millisecond)
		DeferCleanup(l.Cleanup)
	})

	It("can log a message", func() {
		result := captureStdout(func() {
			l.Printf("a message")
		})
		Expect(result).To(MatchRegexp(timestampRegexp + ": a message\n"))
	})

	It("can log the initial totals", func() {
		result := captureStdout(func() {
			l.InitialTotals(1, 2)
		})
		Expect(result).To(MatchRegexp(`total instances: 1\n.*upgradable instances: 2\n`))
	})

	It("can log the start of an upgrade", func() {
		result := captureStdout(func() {
			l.UpgradeStarting("my-instance", "fake-guid")
		})
		Expect(result).To(MatchRegexp(timestampRegexp + `: starting to upgrade instance: "my-instance" guid: "fake-guid"\n`))
	})

	It("can log the success of an upgrade", func() {
		result := captureStdout(func() {
			l.UpgradeSucceeded("my-instance", "fake-guid", time.Minute)
		})
		Expect(result).To(MatchRegexp(timestampRegexp + `: finished upgrade of instance: "my-instance" guid: "fake-guid" successfully after 1m0s\n`))
	})

	It("can log the failure of an upgrade", func() {
		result := captureStdout(func() {
			l.UpgradeFailed("my-instance", "fake-guid", time.Minute, fmt.Errorf("boom"))
		})
		Expect(result).To(MatchRegexp(timestampRegexp + `: upgrade of instance: "my-instance" guid: "fake-guid" failed after 1m0s: boom\n`))
	})

	It("can log the final totals", func() {
		l.InitialTotals(10, 5)
		l.UpgradeFailed("my-first-instance", "fake-guid-1", time.Minute, fmt.Errorf("boom"))
		l.UpgradeFailed("my-second-instance", "fake-guid-2", time.Minute, fmt.Errorf("bang"))
		l.UpgradeSucceeded("my-third-instance", "fake-guid-3", time.Minute)

		result := captureStdout(func() {
			l.FinalTotals()
		})
		Expect(result).To(MatchRegexp(`: successfully upgraded 1 instances\n`))
		Expect(result).To(MatchRegexp(`: failed to upgrade 2 instances\n`))
		Expect(result).To(MatchRegexp(`: my-first-instance\s+| fake-guid-1\s+| boom\n'`))
		Expect(result).To(MatchRegexp(`: my-second-instance\s+| fake-guid-2\s+| bang\n'`))
	})

	It("logs on a ticker", func() {
		l.InitialTotals(10, 5)
		l.UpgradeSucceeded("fake-name", "fake-guid", time.Minute)
		l.UpgradeSucceeded("fake-name", "fake-guid", time.Minute)

		result := captureStdout(func() {
			time.Sleep(150 * time.Millisecond)
		})

		Expect(result).To(MatchRegexp(timestampRegexp + `: upgraded 2 of 5\n`))
	})
})

var captureStdoutLock sync.Mutex

func captureStdout(callback func()) (result string) {
	captureStdoutLock.Lock()

	reader, writer, err := os.Pipe()
	Expect(err).NotTo(HaveOccurred())

	originalStdout := os.Stdout
	os.Stdout = writer

	defer func() {
		writer.Close()
		os.Stdout = originalStdout
		captureStdoutLock.Unlock()

		data, err := io.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		result = string(data)
	}()

	callback()
	return
}
