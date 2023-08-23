package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Client struct {
	Organization string
	URL          string
	Token        string
	Client       influxdb2.Client
	RoundTripper http.RoundTripper
}

func NewClient(url, token, org string, rt http.RoundTripper) *Client {
	return &Client{
		URL:          url,
		Token:        token,
		Organization: org,
		RoundTripper: rt,
	}
}

func (c *Client) Connect() error {
	httpclient := &http.Client{
		Transport: c.RoundTripper,
	}

	options := influxdb2.DefaultOptions().SetHTTPClient(httpclient)

	cfg := influxdb2.NewClientWithOptions(
		c.URL,
		c.Token,
		options,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	conn, err := cfg.Ping(ctx)
	if err != nil {
		err = fmt.Errorf("could not reach influxdb instance: %w", err)
		return err
	}

	if conn {
		c.Client = cfg
	}

	return nil
}

// TODO Move to cmd/query.go.
func (c *Client) GetQueryResult(query string) (res *api.QueryTableResult, err error) {
	queryAPI := c.Client.QueryAPI(c.Organization)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	res, err = queryAPI.Query(ctx, query)
	if err != nil {
		err = fmt.Errorf("could build query: %w", err)
		return
	}

	return
}

// TODO Move to cmd/query.go.
func (c *Client) GetQueryRecords(query string) (records []*query.FluxRecord, err error) {
	res, err := c.GetQueryResult(query)
	if err != nil {
		return
	}

	if res.Err() != nil {
		err = fmt.Errorf("could not parse query: %w", res.Err())
		return
	}

	for res.Next() {
		records = append(records, res.Record())
	}

	if records == nil {
		err = fmt.Errorf("no record has been returned")
		return
	}

	return
}
