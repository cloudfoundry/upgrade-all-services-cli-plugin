// Code generated by counterfeiter. DO NOT EDIT.
package upgraderfakes

import (
	"sync"
	"time"
	"upgrade-all-services-cli-plugin/internal/upgrader"
)

type FakeLogger struct {
	FinalTotalsStub        func()
	finalTotalsMutex       sync.RWMutex
	finalTotalsArgsForCall []struct {
	}
	InitialTotalsStub        func(int, int)
	initialTotalsMutex       sync.RWMutex
	initialTotalsArgsForCall []struct {
		arg1 int
		arg2 int
	}
	PrintfStub        func(string, ...any)
	printfMutex       sync.RWMutex
	printfArgsForCall []struct {
		arg1 string
		arg2 []any
	}
	SkippingInstanceStub        func(string, string, bool, string, string)
	skippingInstanceMutex       sync.RWMutex
	skippingInstanceArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 bool
		arg4 string
		arg5 string
	}
	UpgradeFailedStub        func(string, string, time.Duration, error)
	upgradeFailedMutex       sync.RWMutex
	upgradeFailedArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 time.Duration
		arg4 error
	}
	UpgradeStartingStub        func(string, string)
	upgradeStartingMutex       sync.RWMutex
	upgradeStartingArgsForCall []struct {
		arg1 string
		arg2 string
	}
	UpgradeSucceededStub        func(string, string, time.Duration)
	upgradeSucceededMutex       sync.RWMutex
	upgradeSucceededArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 time.Duration
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeLogger) FinalTotals() {
	fake.finalTotalsMutex.Lock()
	fake.finalTotalsArgsForCall = append(fake.finalTotalsArgsForCall, struct {
	}{})
	stub := fake.FinalTotalsStub
	fake.recordInvocation("FinalTotals", []interface{}{})
	fake.finalTotalsMutex.Unlock()
	if stub != nil {
		fake.FinalTotalsStub()
	}
}

func (fake *FakeLogger) FinalTotalsCallCount() int {
	fake.finalTotalsMutex.RLock()
	defer fake.finalTotalsMutex.RUnlock()
	return len(fake.finalTotalsArgsForCall)
}

func (fake *FakeLogger) FinalTotalsCalls(stub func()) {
	fake.finalTotalsMutex.Lock()
	defer fake.finalTotalsMutex.Unlock()
	fake.FinalTotalsStub = stub
}

func (fake *FakeLogger) InitialTotals(arg1 int, arg2 int) {
	fake.initialTotalsMutex.Lock()
	fake.initialTotalsArgsForCall = append(fake.initialTotalsArgsForCall, struct {
		arg1 int
		arg2 int
	}{arg1, arg2})
	stub := fake.InitialTotalsStub
	fake.recordInvocation("InitialTotals", []interface{}{arg1, arg2})
	fake.initialTotalsMutex.Unlock()
	if stub != nil {
		fake.InitialTotalsStub(arg1, arg2)
	}
}

func (fake *FakeLogger) InitialTotalsCallCount() int {
	fake.initialTotalsMutex.RLock()
	defer fake.initialTotalsMutex.RUnlock()
	return len(fake.initialTotalsArgsForCall)
}

func (fake *FakeLogger) InitialTotalsCalls(stub func(int, int)) {
	fake.initialTotalsMutex.Lock()
	defer fake.initialTotalsMutex.Unlock()
	fake.InitialTotalsStub = stub
}

func (fake *FakeLogger) InitialTotalsArgsForCall(i int) (int, int) {
	fake.initialTotalsMutex.RLock()
	defer fake.initialTotalsMutex.RUnlock()
	argsForCall := fake.initialTotalsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeLogger) Printf(arg1 string, arg2 ...any) {
	fake.printfMutex.Lock()
	fake.printfArgsForCall = append(fake.printfArgsForCall, struct {
		arg1 string
		arg2 []any
	}{arg1, arg2})
	stub := fake.PrintfStub
	fake.recordInvocation("Printf", []interface{}{arg1, arg2})
	fake.printfMutex.Unlock()
	if stub != nil {
		fake.PrintfStub(arg1, arg2...)
	}
}

func (fake *FakeLogger) PrintfCallCount() int {
	fake.printfMutex.RLock()
	defer fake.printfMutex.RUnlock()
	return len(fake.printfArgsForCall)
}

func (fake *FakeLogger) PrintfCalls(stub func(string, ...any)) {
	fake.printfMutex.Lock()
	defer fake.printfMutex.Unlock()
	fake.PrintfStub = stub
}

func (fake *FakeLogger) PrintfArgsForCall(i int) (string, []any) {
	fake.printfMutex.RLock()
	defer fake.printfMutex.RUnlock()
	argsForCall := fake.printfArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeLogger) SkippingInstance(arg1 string, arg2 string, arg3 bool, arg4 string, arg5 string) {
	fake.skippingInstanceMutex.Lock()
	fake.skippingInstanceArgsForCall = append(fake.skippingInstanceArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 bool
		arg4 string
		arg5 string
	}{arg1, arg2, arg3, arg4, arg5})
	stub := fake.SkippingInstanceStub
	fake.recordInvocation("SkippingInstance", []interface{}{arg1, arg2, arg3, arg4, arg5})
	fake.skippingInstanceMutex.Unlock()
	if stub != nil {
		fake.SkippingInstanceStub(arg1, arg2, arg3, arg4, arg5)
	}
}

func (fake *FakeLogger) SkippingInstanceCallCount() int {
	fake.skippingInstanceMutex.RLock()
	defer fake.skippingInstanceMutex.RUnlock()
	return len(fake.skippingInstanceArgsForCall)
}

func (fake *FakeLogger) SkippingInstanceCalls(stub func(string, string, bool, string, string)) {
	fake.skippingInstanceMutex.Lock()
	defer fake.skippingInstanceMutex.Unlock()
	fake.SkippingInstanceStub = stub
}

func (fake *FakeLogger) SkippingInstanceArgsForCall(i int) (string, string, bool, string, string) {
	fake.skippingInstanceMutex.RLock()
	defer fake.skippingInstanceMutex.RUnlock()
	argsForCall := fake.skippingInstanceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5
}

func (fake *FakeLogger) UpgradeFailed(arg1 string, arg2 string, arg3 time.Duration, arg4 error) {
	fake.upgradeFailedMutex.Lock()
	fake.upgradeFailedArgsForCall = append(fake.upgradeFailedArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 time.Duration
		arg4 error
	}{arg1, arg2, arg3, arg4})
	stub := fake.UpgradeFailedStub
	fake.recordInvocation("UpgradeFailed", []interface{}{arg1, arg2, arg3, arg4})
	fake.upgradeFailedMutex.Unlock()
	if stub != nil {
		fake.UpgradeFailedStub(arg1, arg2, arg3, arg4)
	}
}

func (fake *FakeLogger) UpgradeFailedCallCount() int {
	fake.upgradeFailedMutex.RLock()
	defer fake.upgradeFailedMutex.RUnlock()
	return len(fake.upgradeFailedArgsForCall)
}

func (fake *FakeLogger) UpgradeFailedCalls(stub func(string, string, time.Duration, error)) {
	fake.upgradeFailedMutex.Lock()
	defer fake.upgradeFailedMutex.Unlock()
	fake.UpgradeFailedStub = stub
}

func (fake *FakeLogger) UpgradeFailedArgsForCall(i int) (string, string, time.Duration, error) {
	fake.upgradeFailedMutex.RLock()
	defer fake.upgradeFailedMutex.RUnlock()
	argsForCall := fake.upgradeFailedArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeLogger) UpgradeStarting(arg1 string, arg2 string) {
	fake.upgradeStartingMutex.Lock()
	fake.upgradeStartingArgsForCall = append(fake.upgradeStartingArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.UpgradeStartingStub
	fake.recordInvocation("UpgradeStarting", []interface{}{arg1, arg2})
	fake.upgradeStartingMutex.Unlock()
	if stub != nil {
		fake.UpgradeStartingStub(arg1, arg2)
	}
}

func (fake *FakeLogger) UpgradeStartingCallCount() int {
	fake.upgradeStartingMutex.RLock()
	defer fake.upgradeStartingMutex.RUnlock()
	return len(fake.upgradeStartingArgsForCall)
}

func (fake *FakeLogger) UpgradeStartingCalls(stub func(string, string)) {
	fake.upgradeStartingMutex.Lock()
	defer fake.upgradeStartingMutex.Unlock()
	fake.UpgradeStartingStub = stub
}

func (fake *FakeLogger) UpgradeStartingArgsForCall(i int) (string, string) {
	fake.upgradeStartingMutex.RLock()
	defer fake.upgradeStartingMutex.RUnlock()
	argsForCall := fake.upgradeStartingArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeLogger) UpgradeSucceeded(arg1 string, arg2 string, arg3 time.Duration) {
	fake.upgradeSucceededMutex.Lock()
	fake.upgradeSucceededArgsForCall = append(fake.upgradeSucceededArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 time.Duration
	}{arg1, arg2, arg3})
	stub := fake.UpgradeSucceededStub
	fake.recordInvocation("UpgradeSucceeded", []interface{}{arg1, arg2, arg3})
	fake.upgradeSucceededMutex.Unlock()
	if stub != nil {
		fake.UpgradeSucceededStub(arg1, arg2, arg3)
	}
}

func (fake *FakeLogger) UpgradeSucceededCallCount() int {
	fake.upgradeSucceededMutex.RLock()
	defer fake.upgradeSucceededMutex.RUnlock()
	return len(fake.upgradeSucceededArgsForCall)
}

func (fake *FakeLogger) UpgradeSucceededCalls(stub func(string, string, time.Duration)) {
	fake.upgradeSucceededMutex.Lock()
	defer fake.upgradeSucceededMutex.Unlock()
	fake.UpgradeSucceededStub = stub
}

func (fake *FakeLogger) UpgradeSucceededArgsForCall(i int) (string, string, time.Duration) {
	fake.upgradeSucceededMutex.RLock()
	defer fake.upgradeSucceededMutex.RUnlock()
	argsForCall := fake.upgradeSucceededArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeLogger) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.finalTotalsMutex.RLock()
	defer fake.finalTotalsMutex.RUnlock()
	fake.initialTotalsMutex.RLock()
	defer fake.initialTotalsMutex.RUnlock()
	fake.printfMutex.RLock()
	defer fake.printfMutex.RUnlock()
	fake.skippingInstanceMutex.RLock()
	defer fake.skippingInstanceMutex.RUnlock()
	fake.upgradeFailedMutex.RLock()
	defer fake.upgradeFailedMutex.RUnlock()
	fake.upgradeStartingMutex.RLock()
	defer fake.upgradeStartingMutex.RUnlock()
	fake.upgradeSucceededMutex.RLock()
	defer fake.upgradeSucceededMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeLogger) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ upgrader.Logger = new(FakeLogger)
