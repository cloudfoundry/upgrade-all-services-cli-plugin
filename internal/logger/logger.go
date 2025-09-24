package logger

import (
	"fmt"
	"maps"
	"slices"
	"sync"
	"time"
	"upgrade-all-services-cli-plugin/internal/slicex"

	"upgrade-all-services-cli-plugin/internal/ccapi"
)

type instanceState int

const (
	stateUnstarted instanceState = iota
	stateStarted
	stateSucceeded
	stateFailed
	stateSkipped
)

func New(period time.Duration) *Logger {
	l := Logger{
		ticker: time.NewTicker(period),
		states: make(map[string]instanceState),
	}

	go func() {
		for range l.ticker.C {
			l.Printf("%s", l.tickerMessage())
		}
	}()

	return &l
}

type failure struct {
	instance    ccapi.ServiceInstance
	err         error
	attempt, of int
}

type Logger struct {
	lock     sync.Mutex
	ticker   *time.Ticker
	target   int
	states   map[string]instanceState
	failures []failure
}

func (l *Logger) Printf(format string, a ...any) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.printf(format, a...)
}

func (l *Logger) SkippingInstance(instance ccapi.ServiceInstance) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.states[instance.GUID] = stateSkipped
	l.printf("skipping instance: %q guid: %q Upgrade Available: %v Last Operation Type: %q State: %q", instance.Name, instance.GUID, instance.UpgradeAvailable, instance.LastOperationType, instance.LastOperationState)
}

func (l *Logger) UpgradeStarting(instance ccapi.ServiceInstance, attempt, of int) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.states[instance.GUID] = stateStarted
	l.printf("starting to upgrade instance: %q guid: %q%s", instance.Name, instance.GUID, attemptMessage(attempt, of))
}

func (l *Logger) UpgradeSucceeded(instance ccapi.ServiceInstance, attempt, of int, duration time.Duration) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.states[instance.GUID] = stateSucceeded
	l.printf("finished upgrade of instance: %q guid: %q successfully after %s%s", instance.Name, instance.GUID, duration, attemptMessage(attempt, of))
}

func (l *Logger) UpgradeFailed(instance ccapi.ServiceInstance, attempt, of int, duration time.Duration, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.failures = append(l.failures, failure{
		instance: instance,
		err:      err,
		attempt:  attempt,
		of:       of,
	})
	l.states[instance.GUID] = stateFailed
	l.printf("upgrade of instance: %q guid: %q failed after %s%s: %s", instance.Name, instance.GUID, duration, attemptMessage(attempt, of), err)
}

func (l *Logger) InitialTotals(totalServiceInstances, totalUpgradableServiceInstances int) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.target = totalUpgradableServiceInstances

	l.separator()
	l.printf("total instances: %d", totalServiceInstances)
	l.printf("upgradable instances: %d", totalUpgradableServiceInstances)
	l.separator()
	l.printf("starting upgrade...")
}

func (l *Logger) FinalTotals() {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.printf("%s", l.tickerMessage())
	l.separator()
	l.printf("skipped %d instances", l.numInState(stateSkipped))
	l.printf("successfully upgraded %d instances", l.numInState(stateSucceeded))

	logRowFormatTotals(l)
}

func (l *Logger) HasUpgradeSucceeded() bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	return len(l.failures) == 0
}

func logRowFormatTotals(l *Logger) {
	if len(l.failures) > 0 {
		l.printf("failed to upgrade %d instances", l.numInState(stateFailed))
		l.printf("")
		for _, failure := range l.failures {
			fmt.Println()
			fmt.Printf("  Details: %q\n", failure.err)
			if failure.of != 1 {
				fmt.Printf("  Attempt %d of %d\n", failure.attempt, failure.of)
			}
			fmt.Printf("  Service Instance Name: %q\n", failure.instance.Name)
			fmt.Printf("  Service Instance GUID: %q\n", failure.instance.GUID)
			fmt.Printf("  Service Instance Version: %q\n", failure.instance.MaintenanceInfoVersion)
			fmt.Printf("  Service Plan Name: %q\n", failure.instance.ServicePlanName)
			fmt.Printf("  Service Plan GUID: %q\n", failure.instance.ServicePlanGUID)
			fmt.Printf("  Service Plan Version: %q\n", failure.instance.ServicePlanMaintenanceInfoVersion)
			fmt.Printf("  Service Offering Name: %q\n", failure.instance.ServiceOfferingName)
			fmt.Printf("  Service Offering GUID: %q\n", failure.instance.ServiceOfferingGUID)
			fmt.Printf("  Space Name: %q\n", failure.instance.SpaceName)
			fmt.Printf("  Space GUID: %q\n", failure.instance.SpaceGUID)
			fmt.Printf("  Organization Name: %q\n", failure.instance.OrganizationName)
			fmt.Printf("  Organization GUID: %q\n", failure.instance.OrganizationGUID)
		}
	}
}

func (l *Logger) Cleanup() {
	l.ticker.Stop()
}

func (l *Logger) printf(format string, a ...any) {
	fmt.Print(time.Now().Format(time.RFC3339))
	fmt.Print(": ")
	fmt.Printf(format, a...)
	fmt.Println()
}

func (l *Logger) separator() {
	l.printf("---")
}

func (l *Logger) tickerMessage() string {
	return fmt.Sprintf("upgraded %d of %d", l.numInState(stateSucceeded), l.target)
}

func (l *Logger) numInState(s instanceState) int {
	return len(slicex.Filter(slices.Collect(maps.Values(l.states)), func(state instanceState) bool { return state == s }))
}

func attemptMessage(attempt, of int) string {
	if of == 1 {
		return ""
	}
	return fmt.Sprintf(" (attempt %d of %d)", attempt, of)
}
