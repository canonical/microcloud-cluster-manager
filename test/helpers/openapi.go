package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/canonical/lxd/shared/api"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
)

// TestOpenAPIOperation validates a request with a JSON body and its response against the OpenAPI spec.
func TestOpenAPIOperation(t *testing.T, env *Environment, router routers.Router, method string, path string, body any) error {
	reqInput, res, err := validateAndSendRequest(env, router, method, path, body)
	if err != nil {
		return err
	}
	return validateResponse(reqInput, res)
}

// TestOpenAPIOperationInvalid sends an intentionally invalid request (skipping request validation)
// and validates only that the error response conforms to the OpenAPI spec.
func TestOpenAPIOperationInvalid(t *testing.T, env *Environment, router routers.Router, method string, path string, body any) error {
	req, _, err := buildRawRequest(method, path, body)
	if err != nil {
		return err
	}

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return err
	}

	// Build reqInput for response validation only — intentionally skip request validation.
	reqInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc:                openapi3filter.NoopAuthenticationFunc,
			MultiError:                        true,
			RejectWhenRequestBodyNotSpecified: false,
		},
	}

	res, err := sendRequest(env, req)
	if err != nil {
		return err
	}

	return validateResponse(reqInput, res)
}

// validateAndSendRequest validates the given HTTP request with a JSON body against the OpenAPI specification and sends it to the server.
func validateAndSendRequest(env *Environment, router routers.Router, method string, path string, body any) (*openapi3filter.RequestValidationInput, *http.Response, error) {
	reqInput, req, err := validateRequest(router, method, path, body)
	if err != nil {
		return nil, nil, err
	}

	res, err := sendRequest(env, req)
	if err != nil {
		return nil, nil, err
	}

	return reqInput, res, nil
}

// validateRequest validates the given HTTP request against the OpenAPI specification.
// Returns the validation input and the request.
func validateRequest(router routers.Router, method string, path string, body any) (*openapi3filter.RequestValidationInput, *http.Request, error) {
	req, err := buildRequest(method, path, body)
	if err != nil {
		return nil, nil, err
	}

	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		return nil, nil, err
	}

	reqInput := getRequestValidationInput(req, pathParams, route)
	if err = openapi3filter.ValidateRequest(context.Background(), reqInput); err != nil {
		return nil, nil, err
	}

	return reqInput, req, nil
}

// buildRequest constructs an HTTP request with an optional JSON body.
// Returns the request and the raw body bytes (nil if body is nil).
func buildRequest(method string, path string, body any) (*http.Request, error) {
	if body == nil {
		req, err := http.NewRequest(method, path, nil)
		return req, err
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// buildRawRequest constructs an HTTP request with an optional JSON body, bypassing URL parsing validation.
// This is useful for testing error responses triggered by intentionally malformed URLs (e.g. invalid percent-escapes).
// Returns the request and the raw body bytes (nil if body is nil).
func buildRawRequest(method string, path string, body any) (*http.Request, []byte, error) {
	var bodyReader io.Reader
	var bodyBytes []byte

	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Construct the URL directly without parsing to allow malformed paths through.
	u := &url.URL{
		Path:    path,
		RawPath: path,
	}

	req := &http.Request{
		Method: method,
		URL:    u,
		Header: make(http.Header),
	}

	if bodyReader != nil {
		req.Body = io.NopCloser(bodyReader)
		req.ContentLength = int64(len(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
	}

	return req, bodyBytes, nil
}

// getRequestValidationInput creates a new RequestValidationInput for the given HTTP request, path parameters, and route.
func getRequestValidationInput(req *http.Request, pathParams map[string]string, route *routers.Route) *openapi3filter.RequestValidationInput {
	return &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
		Options: &openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			MultiError:         true,
		},
	}
}

// sendRequest sends a test request to the specified endpoint and returns the raw HTTP response.
func sendRequest(env *Environment, req *http.Request) (*http.Response, error) {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	certPublicKey, err := env.ManagementAPICert().PublicKeyX509()
	if err != nil {
		return nil, err
	}

	headers, err := env.ManagementAPILoginHeaders(certPublicKey)
	if err != nil {
		return nil, err
	}

	tlsClient, err := NewTLSHTTPClient(api.URL{}, nil, certPublicKey, env.ManagementAPIHost())
	if err != nil {
		return nil, err
	}

	if err := headers(req); err != nil {
		return nil, err
	}

	req.URL.Scheme = "https"
	req.URL.Host = env.ManagementAPIHostPort()

	return tlsClient.Do(req)
}

// validateResponse validates the given HTTP response against the OpenAPI specification.
func validateResponse(reqInput *openapi3filter.RequestValidationInput, res *http.Response) error {
	respInput := getResponseValidationInput(reqInput, res)

	err := openapi3filter.ValidateResponse(context.Background(), respInput)
	if err != nil {
		return err
	}

	return nil
}

// getResponseValidationInput creates a new ResponseValidationInput for the given HTTP request and response.
func getResponseValidationInput(reqInput *openapi3filter.RequestValidationInput, res *http.Response) *openapi3filter.ResponseValidationInput {
	return &openapi3filter.ResponseValidationInput{
		RequestValidationInput: reqInput,
		Status:                 res.StatusCode,
		Header:                 res.Header,
		Body:                   res.Body,
		Options: &openapi3filter.Options{
			IncludeResponseStatus: true,
		},
	}
}
