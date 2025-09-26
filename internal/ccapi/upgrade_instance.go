package ccapi

import (
	"fmt"
	"time"
)

func (c CCAPI) UpgradeServiceInstance(guid, miVersion string) error {
	body := struct {
		MaintenanceInfoVersion string `jsonry:"maintenance_info.version"`
	}{
		MaintenanceInfoVersion: miVersion,
	}

	err := c.requester.Patch(fmt.Sprintf("v3/service_instances/%s", guid), body)
	if err != nil {
		return fmt.Errorf("upgrade request error: %s", err)
	}

	for timeout := time.After(time.Minute * 10); ; {
		select {
		case <-timeout:
			return fmt.Errorf("error upgrade request timeout")
		default:
			var si ServiceInstance
			err = c.requester.Get(fmt.Sprintf("v3/service_instances/%s", guid), &si)
			if err != nil {
				return fmt.Errorf("upgrade request error: %s", err)
			}

			if si.LastOperationState == "failed" && si.LastOperationType == "update" {
				return fmt.Errorf("%s", si.LastOperationDescription)
			}

			if si.LastOperationState != "in progress" || si.LastOperationType != "update" {
				return nil
			}
		}
		time.Sleep(c.pollingInterval)
	}
}
