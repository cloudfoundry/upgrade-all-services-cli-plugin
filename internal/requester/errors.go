package requester

import "fmt"

type ccAPIErrors struct {
	Errors []ccAPIError
}

type ccAPIError struct {
	Code   int    `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func (c ccAPIError) String() string {
	return fmt.Sprintf("capi_error_code: %d capi_error_title: %s capi_error_detail: %s", c.Code, c.Title, c.Detail)
}
