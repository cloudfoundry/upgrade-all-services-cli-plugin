package requester

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

type Requester struct {
	APIBaseURL string
	APIToken   string
	client     *http.Client
	Logger     Logger
}

type CCApiErrors struct {
	Errors []CCApiError
}
type CCApiError struct {
	Code   int    `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func NewRequester(apiBaseURL, apiToken string, insecureSkipVerify bool) Requester {
	return Requester{
		APIBaseURL: apiBaseURL,
		APIToken:   apiToken,
		Logger:     nullLogger{},
		client: &http.Client{
			Timeout: time.Minute,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
			},
		},
	}
}

type Logger interface {
	Printf(string, ...any)
}

func (r Requester) Get(url string, receiver any) error {
	if reflect.TypeOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf("receiver must be of type Pointer")
	}

	url = fmt.Sprintf("%s/%s", r.APIBaseURL, url)
	r.Logger.Printf("HTTP GET: %s", url)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}
	request.Header.Set("Authorization", r.APIToken)

	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("http request error: %s", err)
	}
	r.Logger.Printf("Response status %s", response.Status)
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("http response: %d", response.StatusCode)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("unable to read http response body error: %s", err)
	}
	r.Logger.Printf("Response body: %s", data)
	err = json.Unmarshal(data, &receiver)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response into receiver error: %s", err)
	}

	return nil
}

func (r Requester) Patch(url string, data any) error {
	d, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %s", err)
	}

	url = fmt.Sprintf("%s/%s", r.APIBaseURL, url)
	r.Logger.Printf("HTTP PATCH: %s", url)
	r.Logger.Printf("Request body: %s", d)

	request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(d))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}
	request.Header.Set("Authorization", r.APIToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := r.client.Do(request)
	if err != nil {
		return fmt.Errorf("http request error: %s", err)
	}
	r.Logger.Printf("Response status %s", response.Status)
	if response.StatusCode != http.StatusAccepted {
		defer response.Body.Close()
		data, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("unable to read http response body error: %s", err)
		}
		r.Logger.Printf("Response body: %s", data)
		receiver := CCApiErrors{}
		err = json.Unmarshal(data, &receiver)
		if err != nil {
			return fmt.Errorf("http_error: %s response_body: %s", response.Status, string(data))
		}
		err = fmt.Errorf("http_error: %s", response.Status)
		for i := range receiver.Errors {
			err = fmt.Errorf("%w capi_error_code: %d capi_error_title: %s capi_error_detail: %s", err, receiver.Errors[i].Code, receiver.Errors[i].Title, receiver.Errors[i].Detail)
		}
		return err
	}

	return nil
}

type nullLogger struct{}

func (nullLogger) Printf(string, ...any) {
	// do nothing
}
