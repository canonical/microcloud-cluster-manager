package helpers

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"

	"github.com/canonical/lxd/shared"
	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/logger"
)

// Client is a wrapper around the http.Client.
type Client struct {
	*http.Client
	url api.URL
}

// NewUnixHTTPClient creates a new http client for Unix sockets.
func NewUnixHTTPClient(url api.URL) (*Client, error) {
	unixPath := shared.HostPath(url.Hostname())
	url.Host(filepath.Base(url.Hostname()))

	// Setup a Unix socket dialer
	unixDial := func(ctx context.Context, network string, addr string) (net.Conn, error) {
		raddr, err := net.ResolveUnixAddr("unix", unixPath)
		if err != nil {
			return nil, err
		}

		var d net.Dialer
		return d.DialContext(ctx, "unix", raddr.String())
	}

	// Define the http transport
	transport := &http.Transport{
		DialContext: unixDial,
	}

	// Define the http client
	client := &http.Client{Transport: transport}

	return &Client{
		Client: client,
		url:    url,
	}, nil
}

// NewTLSHTTPClient creates a new http client for TLS connections with site manager.
func NewTLSHTTPClient(url api.URL, clientCert *shared.CertInfo, serverCert *x509.Certificate) (*Client, error) {
	var tlsConfig *tls.Config
	// if a server cert is provided, we need to setup the client to trust it
	if serverCert != nil {
		tlsConfig = shared.InitTLSConfig()
		tlsConfig.Certificates = []tls.Certificate{}
		var keypair tls.Certificate
		if clientCert != nil {
			keypair = clientCert.KeyPair()
			tlsConfig.Certificates = append(tlsConfig.Certificates, keypair)
		}

		tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &keypair, nil
		}

		// Add the public key to the CA pool to make it trusted.
		tlsConfig.RootCAs = x509.NewCertPool()
		serverCert.IsCA = true
		serverCert.KeyUsage = x509.KeyUsageCertSign
		tlsConfig.RootCAs.AddCert(serverCert)

		// Always use public key DNS name rather than server cert, so that it matches.
		if len(serverCert.DNSNames) > 0 {
			tlsConfig.ServerName = serverCert.DNSNames[0]
		}
	}

	transport := &http.Transport{
		TLSClientConfig:   tlsConfig,
		DisableKeepAlives: true,
	}

	client := &http.Client{Transport: transport}

	return &Client{
		Client: client,
		url:    url,
	}, nil
}

// Query makes a query using the http client (unix or tls).
func (c *Client) Query(ctx context.Context, method string, path *api.URL, input any, output any, adjustHeaders func(*http.Request) error) error {
	// Merge the provided URL with the one we have for the client.
	url := api.NewURL()
	url.URL.Host = c.url.URL.Host
	url.URL.Scheme = c.url.URL.Scheme
	url.URL.Path = path.URL.Path
	url.URL.RawPath = path.URL.RawPath

	if path.URL.Host != "" {
		url.URL.Host = path.URL.Host
	}

	if path.URL.Scheme != "" {
		url.URL.Scheme = path.URL.Scheme
	}

	localQuery := url.URL.Query()
	clientQuery := c.url.URL.Query()
	for q := range url.URL.Query() {
		clientQuery.Set(q, localQuery.Get(q))
	}

	url.URL.RawQuery = clientQuery.Encode()

	// Make the request
	req, err := makeRequest(ctx, method, url, input)
	if err != nil {
		return err
	}

	if adjustHeaders != nil {
		err := adjustHeaders(req)
		if err != nil {
			return err
		}
	}

	// Perform the request
	rawResponse, err := c.Do(req)
	if err != nil {
		return err
	}

	// Decode the response assuming LXD response structure
	parsedResponse, err := parseResponse(rawResponse)
	if err != nil {
		return err
	}

	defer rawResponse.Body.Close()
	_, err = io.Copy(io.Discard, rawResponse.Body)
	if err != nil {
		logger.Error("Failed to read response body", logger.Ctx{"error": err})
	}

	err = json.Unmarshal(parsedResponse.Metadata, &output)
	if err != nil {
		return err
	}

	return nil
}

func makeRequest(ctx context.Context, method string, url *api.URL, data any) (req *http.Request, err error) {
	if data != nil {
		reqBody, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}

		req, err = http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(reqBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url.String(), nil)
		if err != nil {
			return nil, err
		}
	}

	return req, nil
}

func parseResponse(resp *http.Response) (*api.Response, error) {
	// Decode the response
	decoder := json.NewDecoder(resp.Body)
	response := api.Response{}

	err := decoder.Decode(&response)
	if err != nil {
		// Check the return value for a cleaner error
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Failed to fetch %q: %q", resp.Request.URL.String(), resp.Status)
		}

		return nil, err
	}

	// Handle errors
	if response.Type == api.ErrorResponse {
		return nil, api.StatusErrorf(resp.StatusCode, response.Error)
	}

	return &response, nil
}
