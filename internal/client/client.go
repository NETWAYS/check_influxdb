package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Client struct {
	Organization string
	Url          string
	Token        string
	Insecure     bool
	Client       influxdb2.Client
	LogLevel     uint
	Average      int
	Critical     int
	Warning      int
}

func NewClient(url, token, org string) *Client {
	return &Client{
		Url:          url,
		Token:        token,
		Insecure:     false,
		Organization: org,
	}
}

// nolint: gosec
func (c *Client) Connect() error {
	cfg := influxdb2.NewClientWithOptions(
		c.Url,
		c.Token,
		influxdb2.DefaultOptions().SetTLSConfig(&tls.Config{
			InsecureSkipVerify: c.Insecure,
		}).SetLogLevel(c.LogLevel))

	ctx, cancel := c.timeoutContext()
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

func (c *Client) timeoutContext() (context.Context, func()) {
	// TODO Add timeout config
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (c *Client) Health() (*domain.HealthCheck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := c.Client.Health(ctx)
	if err != nil {
		err = fmt.Errorf("could not fetch health: %w", err)
		return nil, err
	}

	return health, nil
}

func (c *Client) GetQueryResult(query string) (res *api.QueryTableResult, err error) {
	queryApi := c.Client.QueryAPI(c.Organization)

	ctx, cancel := c.timeoutContext()
	defer cancel()

	res, err = queryApi.Query(ctx, query)
	if err != nil {
		err = fmt.Errorf("could build query: %w", err)
		return
	}

	return
}

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
