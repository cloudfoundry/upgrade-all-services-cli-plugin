// Code generated by counterfeiter. DO NOT EDIT.
package upgraderfakes

import (
	"sync"
	"time"
	"upgrade-all-services-cli-plugin/internal/ccapi"
	"upgrade-all-services-cli-plugin/internal/upgrader"
)

type FakeLogger struct {
	DeactivatedPlanStub        func(ccapi.ServiceInstance)
	deactivatedPlanMutex       sync.RWMutex
	deactivatedPlanArgsForCall []struct {
		arg1 ccapi.ServiceInstance
	}
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
	InstanceIsNotUpToDateStub        func(ccapi.ServiceInstance)
	instanceIsNotUpToDateMutex       sync.RWMutex
	instanceIsNotUpToDateArgsForCall []struct {
		arg1 ccapi.ServiceInstance
	}
	PrintfStub        func(string, ...any)
	printfMutex       sync.RWMutex
	printfArgsForCall []struct {
		arg1 string
		arg2 []any
	}
	SkippingInstanceStub        func(ccapi.ServiceInstance)
	skippingInstanceMutex       sync.RWMutex
	skippingInstanceArgsForCall []struct {
		arg1 ccapi.ServiceInstance
	}
	UpgradeFailedStub        func(ccapi.ServiceInstance, time.Duration, error)
	upgradeFailedMutex       sync.RWMutex
	upgradeFailedArgsForCall []struct {
		arg1 ccapi.ServiceInstance
		arg2 time.Duration
		arg3 error
	}
	UpgradeStartingStub        func(ccapi.ServiceInstance)
	upgradeStartingMutex       sync.RWMutex
	upgradeStartingArgsForCall []struct {
		arg1 ccapi.ServiceInstance
	}
	UpgradeSucceededStub        func(ccapi.ServiceInstance, time.Duration)
	upgradeSucceededMutex       sync.RWMutex
	upgradeSucceededArgsForCall []struct {
		arg1 ccapi.ServiceInstance
		arg2 time.Duration
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeLogger) DeactivatedPlan(arg1 ccapi.ServiceInstance) {
	fake.deactivatedPlanMutex.Lock()
	fake.deactivatedPlanArgsForCall = append(fake.deactivatedPlanArgsForCall, struct {
		arg1 ccapi.ServiceInstance
	}{arg1})
	stub := fake.DeactivatedPlanStub
	fake.recordInvocation("DeactivatedPlan", []interface{}{arg1})
	fake.deactivatedPlanMutex.Unlock()
	if stub != nil {
		fake.DeactivatedPlanStub(arg1)
	}
}

func (fake *FakeLogger) DeactivatedPlanCallCount() int {
	fake.deactivatedPlanMutex.RLock()
	defer fake.deactivatedPlanMutex.RUnlock()
	return len(fake.deactivatedPlanArgsForCall)
}

func (fake *FakeLogger) DeactivatedPlanCalls(stub func(ccapi.ServiceInstance)) {
	fake.deactivatedPlanMutex.Lock()
	defer fake.deactivatedPlanMutex.Unlock()
	fake.DeactivatedPlanStub = stub
}

func (fake *FakeLogger) DeactivatedPlanArgsForCall(i int) ccapi.ServiceInstance {
	fake.deactivatedPlanMutex.RLock()
	defer fake.deactivatedPlanMutex.RUnlock()
	argsForCall := fake.deactivatedPlanArgsForCall[i]
	return argsForCall.arg1
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

func (fake *FakeLogger) InstanceIsNotUpToDate(arg1 ccapi.ServiceInstance) {
	fake.instanceIsNotUpToDateMutex.Lock()
	fake.instanceIsNotUpToDateArgsForCall = append(fake.instanceIsNotUpToDateArgsForCall, struct {
		arg1 ccapi.ServiceInstance
	}{arg1})
	stub := fake.InstanceIsNotUpToDateStub
	fake.recordInvocation("InstanceIsNotUpToDate", []interface{}{arg1})
	fake.instanceIsNotUpToDateMutex.Unlock()
	if stub != nil {
		fake.InstanceIsNotUpToDateStub(arg1)
	}
}

func (fake *FakeLogger) InstanceIsNotUpToDateCallCount() int {
	fake.instanceIsNotUpToDateMutex.RLock()
	defer fake.instanceIsNotUpToDateMutex.RUnlock()
	return len(fake.instanceIsNotUpToDateArgsForCall)
}

func (fake *FakeLogger) InstanceIsNotUpToDateCalls(stub func(ccapi.ServiceInstance)) {
	fake.instanceIsNotUpToDateMutex.Lock()
	defer fake.instanceIsNotUpToDateMutex.Unlock()
	fake.InstanceIsNotUpToDateStub = stub
}

func (fake *FakeLogger) InstanceIsNotUpToDateArgsForCall(i int) ccapi.ServiceInstance {
	fake.instanceIsNotUpToDateMutex.RLock()
	defer fake.instanceIsNotUpToDateMutex.RUnlock()
	argsForCall := fake.instanceIsNotUpToDateArgsForCall[i]
	return argsForCall.arg1
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

func (fake *FakeLogger) SkippingInstance(arg1 ccapi.ServiceInstance) {
	fake.skippingInstanceMutex.Lock()
	fake.skippingInstanceArgsForCall = append(fake.skippingInstanceArgsForCall, struct {
		arg1 ccapi.ServiceInstance
	}{arg1})
	stub := fake.SkippingInstanceStub
	fake.recordInvocation("SkippingInstance", []interface{}{arg1})
	fake.skippingInstanceMutex.Unlock()
	if stub != nil {
		fake.SkippingInstanceStub(arg1)
	}
}

func (fake *FakeLogger) SkippingInstanceCallCount() int {
	fake.skippingInstanceMutex.RLock()
	defer fake.skippingInstanceMutex.RUnlock()
	return len(fake.skippingInstanceArgsForCall)
}

func (fake *FakeLogger) SkippingInstanceCalls(stub func(ccapi.ServiceInstance)) {
	fake.skippingInstanceMutex.Lock()
	defer fake.skippingInstanceMutex.Unlock()
	fake.SkippingInstanceStub = stub
}

func (fake *FakeLogger) SkippingInstanceArgsForCall(i int) ccapi.ServiceInstance {
	fake.skippingInstanceMutex.RLock()
	defer fake.skippingInstanceMutex.RUnlock()
	argsForCall := fake.skippingInstanceArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeLogger) UpgradeFailed(arg1 ccapi.ServiceInstance, arg2 time.Duration, arg3 error) {
	fake.upgradeFailedMutex.Lock()
	fake.upgradeFailedArgsForCall = append(fake.upgradeFailedArgsForCall, struct {
		arg1 ccapi.ServiceInstance
		arg2 time.Duration
		arg3 error
	}{arg1, arg2, arg3})
	stub := fake.UpgradeFailedStub
	fake.recordInvocation("UpgradeFailed", []interface{}{arg1, arg2, arg3})
	fake.upgradeFailedMutex.Unlock()
	if stub != nil {
		fake.UpgradeFailedStub(arg1, arg2, arg3)
	}
}

func (fake *FakeLogger) UpgradeFailedCallCount() int {
	fake.upgradeFailedMutex.RLock()
	defer fake.upgradeFailedMutex.RUnlock()
	return len(fake.upgradeFailedArgsForCall)
}

func (fake *FakeLogger) UpgradeFailedCalls(stub func(ccapi.ServiceInstance, time.Duration, error)) {
	fake.upgradeFailedMutex.Lock()
	defer fake.upgradeFailedMutex.Unlock()
	fake.UpgradeFailedStub = stub
}

func (fake *FakeLogger) UpgradeFailedArgsForCall(i int) (ccapi.ServiceInstance, time.Duration, error) {
	fake.upgradeFailedMutex.RLock()
	defer fake.upgradeFailedMutex.RUnlock()
	argsForCall := fake.upgradeFailedArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeLogger) UpgradeStarting(arg1 ccapi.ServiceInstance) {
	fake.upgradeStartingMutex.Lock()
	fake.upgradeStartingArgsForCall = append(fake.upgradeStartingArgsForCall, struct {
		arg1 ccapi.ServiceInstance
	}{arg1})
	stub := fake.UpgradeStartingStub
	fake.recordInvocation("UpgradeStarting", []interface{}{arg1})
	fake.upgradeStartingMutex.Unlock()
	if stub != nil {
		fake.UpgradeStartingStub(arg1)
	}
}

func (fake *FakeLogger) UpgradeStartingCallCount() int {
	fake.upgradeStartingMutex.RLock()
	defer fake.upgradeStartingMutex.RUnlock()
	return len(fake.upgradeStartingArgsForCall)
}

func (fake *FakeLogger) UpgradeStartingCalls(stub func(ccapi.ServiceInstance)) {
	fake.upgradeStartingMutex.Lock()
	defer fake.upgradeStartingMutex.Unlock()
	fake.UpgradeStartingStub = stub
}

func (fake *FakeLogger) UpgradeStartingArgsForCall(i int) ccapi.ServiceInstance {
	fake.upgradeStartingMutex.RLock()
	defer fake.upgradeStartingMutex.RUnlock()
	argsForCall := fake.upgradeStartingArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeLogger) UpgradeSucceeded(arg1 ccapi.ServiceInstance, arg2 time.Duration) {
	fake.upgradeSucceededMutex.Lock()
	fake.upgradeSucceededArgsForCall = append(fake.upgradeSucceededArgsForCall, struct {
		arg1 ccapi.ServiceInstance
		arg2 time.Duration
	}{arg1, arg2})
	stub := fake.UpgradeSucceededStub
	fake.recordInvocation("UpgradeSucceeded", []interface{}{arg1, arg2})
	fake.upgradeSucceededMutex.Unlock()
	if stub != nil {
		fake.UpgradeSucceededStub(arg1, arg2)
	}
}

func (fake *FakeLogger) UpgradeSucceededCallCount() int {
	fake.upgradeSucceededMutex.RLock()
	defer fake.upgradeSucceededMutex.RUnlock()
	return len(fake.upgradeSucceededArgsForCall)
}

func (fake *FakeLogger) UpgradeSucceededCalls(stub func(ccapi.ServiceInstance, time.Duration)) {
	fake.upgradeSucceededMutex.Lock()
	defer fake.upgradeSucceededMutex.Unlock()
	fake.UpgradeSucceededStub = stub
}

func (fake *FakeLogger) UpgradeSucceededArgsForCall(i int) (ccapi.ServiceInstance, time.Duration) {
	fake.upgradeSucceededMutex.RLock()
	defer fake.upgradeSucceededMutex.RUnlock()
	argsForCall := fake.upgradeSucceededArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeLogger) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.deactivatedPlanMutex.RLock()
	defer fake.deactivatedPlanMutex.RUnlock()
	fake.finalTotalsMutex.RLock()
	defer fake.finalTotalsMutex.RUnlock()
	fake.initialTotalsMutex.RLock()
	defer fake.initialTotalsMutex.RUnlock()
	fake.instanceIsNotUpToDateMutex.RLock()
	defer fake.instanceIsNotUpToDateMutex.RUnlock()
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
