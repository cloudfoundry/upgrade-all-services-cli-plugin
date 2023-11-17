package requester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"code.cloudfoundry.org/jsonry"
)

func (r Requester) Patch(url string, data any) error {
	if reflect.TypeOf(data).Kind() != reflect.Struct {
		return fmt.Errorf("input body must be a struct")
	}

	d, err := jsonry.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %s", err)
	}

	url = fmt.Sprintf("%s/%s", r.baseURL, url)
	r.Logger.Printf("HTTP PATCH: %s", url)
	r.Logger.Printf("Request body: %s", d)

	request, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(d))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}
	request.Header.Set("Authorization", r.token)
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

		var receiver ccAPIErrors
		err = json.Unmarshal(data, &receiver)
		if err != nil {
			return fmt.Errorf("http_error: %s response_body: %s", response.Status, string(data))
		}
		err = fmt.Errorf("http_error: %s", response.Status)
		for _, e := range receiver.Errors {
			err = fmt.Errorf("%w %s", err, e)
		}
		return err
	}

	return nil
}
