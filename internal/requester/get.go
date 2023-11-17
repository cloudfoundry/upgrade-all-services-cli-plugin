package requester

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"code.cloudfoundry.org/jsonry"
)

func (r Requester) Get(url string, receiver any) error {
	receiverType := reflect.ValueOf(receiver)
	switch {
	case receiverType.Kind() != reflect.Ptr:
		return fmt.Errorf("receiver must be a pointer to a struct, got non-pointer")
	case receiverType.Elem().Kind() != reflect.Struct:
		return fmt.Errorf("receiver must be a pointer to a struct, got non-struct")
	}

	url = fmt.Sprintf("%s/%s", r.baseURL, url)
	r.Logger.Printf("HTTP GET: %s", url)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %s", err)
	}
	request.Header.Set("Authorization", r.token)

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
	err = jsonry.Unmarshal(data, receiver)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response into receiver error: %s", err)
	}

	return nil
}
