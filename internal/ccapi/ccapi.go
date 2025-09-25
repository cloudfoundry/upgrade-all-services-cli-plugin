package ccapi

import (
	"time"
	"upgrade-all-services-cli-plugin/internal/requester"
)

type CCAPI struct {
	requester       requester.Requester
	pollingInterval time.Duration
}

func NewCCAPI(req requester.Requester, pollingInterval time.Duration) CCAPI {
	return CCAPI{
		requester:       req,
		pollingInterval: pollingInterval,
	}
}
