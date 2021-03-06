package logger

import (
	"fmt"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

func New(period time.Duration) *Logger {
	l := Logger{
		ticker:   time.NewTicker(period),
		failures: []failure{},
	}

	go func() {
		for range l.ticker.C {
			l.Printf(l.tickerMessage())
		}
	}()

	return &l
}

type failure struct {
	name string
	guid string
	err  error
}

type Logger struct {
	lock      sync.Mutex
	ticker    *time.Ticker
	target    int
	complete  int
	successes int
	failures  []failure
}

func (l *Logger) Printf(format string, a ...any) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.printf(format, a...)
}

func (l *Logger) UpgradeStarting(name, guid string) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.printf("starting to upgrade instance: %q guid: %q", name, guid)
}

func (l *Logger) UpgradeSucceeded(name, guid string, duration time.Duration) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.successes++
	l.complete++
	l.printf("finished upgrade of instance: %q guid: %q successfully after %s", name, guid, duration)
}

func (l *Logger) UpgradeFailed(name, guid string, duration time.Duration, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.failures = append(l.failures, failure{
		name: name,
		guid: guid,
		err:  err,
	})
	l.complete++
	l.printf("upgrade of instance: %q guid: %q failed after %s: %s", name, guid, duration, err)
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

	l.printf(l.tickerMessage())
	l.separator()
	l.printf("successfully upgraded %d instances", l.successes)

	if len(l.failures) > 0 {
		l.printf("failed to upgrade %d instances", len(l.failures))
		l.printf("")

		var sb strings.Builder
		tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', tabwriter.Debug)
		fmt.Fprintln(tw, "Service Instance Name\tService Instance GUID\t Details")
		fmt.Fprintln(tw, "---------------------\t---------------------\t -------")

		for _, failure := range l.failures {
			fmt.Fprintf(tw, "%s\t %s\t %s\n", failure.name, failure.guid, failure.err)
		}
		tw.Flush()

		for _, line := range strings.Split(sb.String(), "\n") {
			l.printf(line)
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
	return fmt.Sprintf("upgraded %d of %d", l.complete, l.target)
}
