package request

import "net/http"

// IsSuccessStatus reports whether the status is a 2xx status code.
func IsSuccessStatus(status int) bool {
	return status >= http.StatusOK && status < http.StatusMultipleChoices
}
