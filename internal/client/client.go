package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/NETWAYS/check_influxdb/internal/api"
)

type Client struct {
	Organization string
	URL          string
	Token        string
	Client       *http.Client
}

func NewClient(url, token, org string, rt http.RoundTripper) *Client {
	// Small wrapper
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
		return v, fmt.Errorf("could not reach influxdb instance")
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
