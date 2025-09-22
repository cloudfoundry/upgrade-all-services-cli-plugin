package logger_test

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"upgrade-all-services-cli-plugin/internal/ccapi"
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

	It("can log that it is skipping an instance", func() {
		result := captureStdout(func() {
			l.SkippingInstance(createFailedInstance())
		})
		Expect(result).To(MatchRegexp(timestampRegexp + `: skipping instance: "create-failed-instance" guid: "create-failed-instance-guid" Upgrade Available: true Last Operation Type: "create" State: "failed"\n`))
	})

	It("can log the start of an upgrade", func() {
		result := captureStdout(func() {
			l.UpgradeStarting(upgradeableInstance(1))
		})
		Expect(result).To(MatchRegexp(timestampRegexp + `: starting to upgrade instance: "my-service-instance-1" guid: "my-service-instance-guid-1"\n`))
	})

	It("can log the success of an upgrade", func() {
		result := captureStdout(func() {
			l.UpgradeSucceeded(upgradeableInstance(1), time.Minute)
		})
		Expect(result).To(MatchRegexp(timestampRegexp + `: finished upgrade of instance: "my-service-instance-1" guid: "my-service-instance-guid-1" successfully after 1m0s\n`))
	})

	It("can log the failure of an upgrade", func() {
		result := captureStdout(func() {
			l.UpgradeFailed(upgradeableInstance(1), time.Minute, fmt.Errorf("boom"))
		})
		Expect(result).To(MatchRegexp(timestampRegexp + `: upgrade of instance: "my-service-instance-1" guid: "my-service-instance-guid-1" failed after 1m0s: boom\n`))
	})

	It("can log the final totals", func() {
		l.InitialTotals(10, 5)
		l.UpgradeFailed(upgradeableInstance(1), time.Minute, fmt.Errorf("boom"))
		l.UpgradeFailed(upgradeableInstance(2), time.Minute, fmt.Errorf("bang"))
		l.UpgradeSucceeded(upToDateInstance(3), time.Minute)
		l.SkippingInstance(createFailedInstance())

		result := captureStdout(func() {
			l.FinalTotals()
		})
		Expect(result).To(MatchRegexp(`: skipped 1 instances\n`))
		Expect(result).To(MatchRegexp(`: successfully upgraded 1 instances\n`))
		Expect(result).To(MatchRegexp(`: failed to upgrade 2 instances\n`))
		Expect(result).To(MatchRegexp(`Details: "boom"\n`))
		Expect(result).To(MatchRegexp(`Service Instance Name: "my-service-instance-1"\s+`))
		Expect(result).To(MatchRegexp(`Service Instance GUID: "my-service-instance-guid-1"\s+`))
		Expect(result).To(MatchRegexp(`Service Instance Version: "fake-version-1"\s+`))
		Expect(result).To(MatchRegexp(`Service Plan Name: "fake-plan-name-1"\s+`))
		Expect(result).To(MatchRegexp(`Service Plan GUID: "fake-plan-guid-1"\s+`))
		Expect(result).To(MatchRegexp(`Service Plan Version: "fake-plan-version-1"\s+`))
		Expect(result).To(MatchRegexp(`Service Offering Name: "fake-soffer-name-1"\s+`))
		Expect(result).To(MatchRegexp(`Service Offering GUID: "fake-soffer-guid-1"\s+`))
		Expect(result).To(MatchRegexp(`Space Name: "fake-space-name-1"\s+`))
		Expect(result).To(MatchRegexp(`Space GUID: "fake-space-guid-1"\s+`))
		Expect(result).To(MatchRegexp(`Organization Name: "fake-org-name-1"\s+`))
		Expect(result).To(MatchRegexp(`Organization GUID: "fake-org-guid-1"\s+`))
		Expect(result).To(MatchRegexp(`Details: "bang"\n`))
		Expect(result).To(MatchRegexp(`Service Instance Name: "my-service-instance-2"\s+`))
		Expect(result).To(MatchRegexp(`Service Instance GUID: "my-service-instance-guid-2"\s+`))
		Expect(result).To(MatchRegexp(`Service Instance Version: "fake-version-2"\s+`))
		Expect(result).To(MatchRegexp(`Service Plan Name: "fake-plan-name-2"\s+`))
		Expect(result).To(MatchRegexp(`Service Plan GUID: "fake-plan-guid-2"\s+`))
		Expect(result).To(MatchRegexp(`Service Plan Version: "fake-plan-version-2"\s+`))
		Expect(result).To(MatchRegexp(`Service Offering Name: "fake-soffer-name-2"\s+`))
		Expect(result).To(MatchRegexp(`Service Offering GUID: "fake-soffer-guid-2"\s+`))
		Expect(result).To(MatchRegexp(`Space Name: "fake-space-name-2"\s+`))
		Expect(result).To(MatchRegexp(`Space GUID: "fake-space-guid-2"\s+`))
		Expect(result).To(MatchRegexp(`Organization Name: "fake-org-name-2"\s+`))
		Expect(result).To(MatchRegexp(`Organization GUID: "fake-org-guid-2"\s+`))
	})

	It("logs on a ticker", func() {
		l.InitialTotals(10, 5)
		l.UpgradeSucceeded(upgradeableInstance(1), time.Minute)
		l.UpgradeSucceeded(upgradeableInstance(1), time.Minute)

		result := captureStdout(func() {
			time.Sleep(150 * time.Millisecond)
		})

		Expect(result).To(MatchRegexp(timestampRegexp + `: upgraded 2 of 5\n`))
	})

	Describe("HasUpgradeSucceded", func() {
		It("can signal upgrade failures", func() {
			l.UpgradeFailed(upgradeableInstance(1), time.Minute, fmt.Errorf("boom"))
			Expect(l.HasUpgradeSucceeded()).To(BeFalse())
		})

		It("can signal upgrade success", func() {
			l.UpgradeSucceeded(upgradeableInstance(1), time.Minute)
			l.SkippingInstance(indexedInstance(1, false))
			Expect(l.HasUpgradeSucceeded()).To(BeTrue())
		})

	})
})

func createFailedInstance() ccapi.ServiceInstance {
	return ccapi.ServiceInstance{
		Name:               "create-failed-instance",
		GUID:               "create-failed-instance-guid",
		UpgradeAvailable:   true,
		LastOperationType:  "create",
		LastOperationState: "failed",
	}
}

func formatValue(stringID string, index int) string {
	return fmt.Sprintf("%s-%d", stringID, index)
}
func upgradeableInstance(index int) ccapi.ServiceInstance {
	return indexedInstance(index, true)
}

func upToDateInstance(index int) ccapi.ServiceInstance {
	return indexedInstance(index, false)
}

func indexedInstance(index int, upgradeAvailable bool) ccapi.ServiceInstance {
	return ccapi.ServiceInstance{
		Name: formatValue("my-service-instance", index),
		GUID: formatValue("my-service-instance-guid", index),

		ServicePlanGUID: formatValue("fake-plan-guid", index),
		SpaceGUID:       formatValue("fake-space-guid", index),

		MaintenanceInfoVersion:            formatValue("fake-version", index),
		ServicePlanMaintenanceInfoVersion: formatValue("fake-plan-version", index),

		UpgradeAvailable:   upgradeAvailable,
		LastOperationType:  formatValue("last-operation-type", index),
		LastOperationState: formatValue("last-operation-state", index),

		ServiceOfferingGUID: formatValue("fake-soffer-guid", index),
		ServiceOfferingName: formatValue("fake-soffer-name", index),
		ServicePlanName:     formatValue("fake-plan-name", index),
		OrganizationGUID:    formatValue("fake-org-guid", index),
		OrganizationName:    formatValue("fake-org-name", index),
		SpaceName:           formatValue("fake-space-name", index),
	}
}

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
