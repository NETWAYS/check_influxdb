package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	influxdb "github.com/NETWAYS/check_influxdb/internal/api"
)

// Client is a small wrapper for HTTP connections,
// that is configured and used in the subcommands.
type Client struct {
	Organization string
	URL          string
	Token        string
	Client       *http.Client
}

func NewClient(url, token, org string, rt http.RoundTripper) *Client {
	c := &http.Client{
		Transport: rt,
	}

	return &Client{
		URL:          url,
		Token:        token,
		Organization: org,
		Client:       c,
	}
}

// Version returns the Version of the InfluxDB API.
func (c *Client) Version() (influxdb.APIVersion, error) {
	var v influxdb.APIVersion

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	u, _ := url.JoinPath(c.URL, "/health")

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)

	// Do the HTTP Request to the URL.
	resp, err := c.Client.Do(req)
	if resp == nil {
		return v, errors.New("could not reach influxdb instance")
	}

	if err != nil {
		return v, err
	}

	if resp.StatusCode != http.StatusOK {
		return v, fmt.Errorf("could not get %s - Error: %d", u, resp.StatusCode)
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&v)

	if err != nil {
		return v, err
	}

	return v, nil
}

type headersRoundTripper struct {
	headers map[string]string
	rt      http.RoundTripper
}

// NewHeadersRoundTripper adds the given headers to a request
func NewHeadersRoundTripper(headers map[string]string, rt http.RoundTripper) http.RoundTripper {
	return &headersRoundTripper{headers, rt}
}

func (rt *headersRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// RoundTrip should not modify the request, except for
	// consuming and closing the Request's Body.
	req = cloneRequest(req)

	for key, value := range rt.headers {
		req.Header.Add(key, value)
	}

	return rt.rt.RoundTrip(req)
}

// cloneRequest returns a clone of the provided *http.Request
func cloneRequest(r *http.Request) *http.Request {
	// Shallow copy of the struct.
	r2 := new(http.Request)
	*r2 = *r
	// Deep copy of the Header.
	r2.Header = make(http.Header)
	for k, s := range r.Header {
		r2.Header[k] = s
	}

	return r2
}
