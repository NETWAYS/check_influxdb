package client

import (
	"context"
	"crypto/tls"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"time"
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

//nolint: gosec
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
