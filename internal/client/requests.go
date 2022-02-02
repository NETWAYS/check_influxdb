package client

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func (c *Client) Health() (*domain.HealthCheck, error) {
	health, err := c.Client.Health(context.Background())
	if err != nil {
		err = fmt.Errorf("could not fetch health: %w", err)
		return nil, err
	}

	return health, nil
}
