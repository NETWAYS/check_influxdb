package client

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/query"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"time"
)

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

func (c *Client) GetSingleQueryResult(query string) (record *query.FluxRecord, err error) {
	res, err := c.GetQueryResult(query)
	if err != nil {
		return
	}

	if res.Err() != nil {
		err = fmt.Errorf("could not parse query: %s", res.Err().Error())
		return
	}

	for res.Next() {
		if record != nil {
			err = fmt.Errorf("more than one record has been returned")
			return
		}

		record = res.Record()
	}

	if record == nil {
		err = fmt.Errorf("no record has been returned")
		return
	}

	return
}
