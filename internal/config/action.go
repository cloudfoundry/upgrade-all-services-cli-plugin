package config

import (
	"fmt"
	"sort"
	"strings"
)

type Action int

const (
	InvalidAction Action = iota
	UpgradeAction
	DryRunAction
	CheckUpToDateAction
	CheckDeactivatedPlansAction
	MinVersionCheckAction
)

func determineAction(checkDeactivatedPlans, checkUpToDate, dryRun bool, minVersionRequired string) (Action, error) {
	spec := map[string]bool{
		checkDeactivatedPlansFlag: checkDeactivatedPlans,
		checkUpToDateFlag:         checkUpToDate,
		dryRunFlag:                dryRun,
		minVersionRequiredFlag:    minVersionRequired != "",
	}

	var flagsSpecified []string
	for k, v := range spec {
		if v {
			flagsSpecified = append(flagsSpecified, fmt.Sprintf("--%s", k))
		}
	}

	if len(flagsSpecified) > 1 {
		sort.Strings(flagsSpecified) // Because maps order is random, ensures identical error message each time
		return InvalidAction, fmt.Errorf("invalid flag combination: %s", strings.Join(flagsSpecified, ", "))
	}

	switch {
	case checkDeactivatedPlans:
		return CheckDeactivatedPlansAction, nil
	case checkUpToDate:
		return CheckUpToDateAction, nil
	case dryRun:
		return DryRunAction, nil
	case minVersionRequired != "":
		return MinVersionCheckAction, nil
	default:
		return UpgradeAction, nil
	}
}
