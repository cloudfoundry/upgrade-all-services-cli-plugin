package requester

import (
	"crypto/tls"
	"net/http"
	"time"
)

func NewRequester(apiBaseURL, apiToken string, insecureSkipVerify bool) Requester {
	return Requester{
		baseURL: apiBaseURL,
		token:   apiToken,
		Logger:  nullLogger{},
		client: &http.Client{
			Timeout: time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
			},
		},
	}
}

type Requester struct {
	baseURL string
	token   string
	client  *http.Client
	Logger  Logger
}
