package internal

import (
	"context"
	"fmt"
	"github.com/NETWAYS/go-check"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type Client struct {
	Client influxdb2.Client
}


func (c *Client) GetStatus()  {
	ctx := context.Background()
	h, err := c.Client.Health(ctx)
	if err != nil {
		err := fmt.Errorf("could not get health: %w", err)
		check.ExitError(err)
	}

	if h.Status != "pass" {

	}
}
