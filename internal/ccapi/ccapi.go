package ccapi

import "upgrade-all-services-cli-plugin/internal/requester"

type CCAPI struct {
	requester requester.Requester
}

func NewCCAPI(req requester.Requester) CCAPI {
	return CCAPI{
		requester: req,
	}
}
