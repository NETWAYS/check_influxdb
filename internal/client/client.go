package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/NETWAYS/go-check"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"time"
)

const (
	TB = 1099511627776
	GB = 1073741824
	MB = 1048576
	KB = 1024

	Percent      = "%"
	MegabyteChar = "MB"
	GigabyteChar = "GB"
	TerabyteChar = "TB"
)

type Client struct {
	Host         string
	Port         int
	Proto        string
	User         string
	Pass         string
	Organization string
	Bucket       string
	Query        string
	StatusMode   string

	Url      string
	Token    string
	Insecure bool
	Client   influxdb2.Client
	LogLevel uint
	Average  int
	Critical int
	Warning  int
}

func NewClient(url, token, org string) *Client {
	return &Client{
		Url:          url,
		Token:        token,
		Insecure:     false,
		Organization: org,
	}
}

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

func (c *Client) ExecuteQuery(ifClient influxdb2.Client, query string) (rc int, output string) {
	var value int
	var field string
	var measurement string

	if c.Token != "" {
		value, measurement, field = c.parseQueryValueType(ifClient, query)
		//value, measurement, field = c.parseQueryValueType(ifClient, c.RawQuery)
		fmt.Println(value)

		defer ifClient.Close()
	} else {
		err := ifClient.UsersAPI().SignIn(context.Background(), c.User, c.Pass)
		if err != nil {
			check.ExitError(err)
		}

		value, measurement, field = c.parseQueryValueType(ifClient, query)
		//value, measurement, field = c.parseQueryValueType(ifClient, c.RawQuery)

		// Sign out
		err = ifClient.UsersAPI().SignOut(context.Background())
		if err != nil {
			check.ExitError(err)
		}
		defer ifClient.Close()
	}

	output = fmt.Sprintf("%s %s: %v | '%v'=%v;%v;%v;;",
		measurement, field, value, measurement, value, c.Warning, c.Critical)

	if value >= c.Critical {
		rc = 2
	} else if value >= c.Warning {
		rc = 1
	} else if value < c.Warning {
		rc = 0
	} else {
		rc = 3
	}

	return
}

func (c *Client) parseQueryValueType(ifClient influxdb2.Client, query string) (value int, measurement string, field string) {
	queryApi := ifClient.QueryAPI(c.Organization)

	fmt.Println(query)

	//result, err := queryApi.RawQuery(context.Background(), c.RawQuery)
	result, err := queryApi.Query(context.Background(), query)
	if err == nil {
		counter := 0
		for result.Next() {
			if counter > 0 {
				err := fmt.Errorf("query has more then one value")
				check.ExitError(err)
			}

			measurement = result.Record().Measurement()
			field = result.Record().Field()

			switch res := result.Record().Value().(type) {
			case float64:
				value = int(res)
			case float32:
				value = int(res)
			case int64:
				value = int(res)
			case int32:
				value = int(res)
			case int16:
				value = int(res)
			case int8:
				value = int(res)
			case int:
				value = res
			case string:
				err := fmt.Errorf("string value can not be evaluated")
				check.ExitError(err)
			default:
				err := fmt.Errorf("unknown data type")
				check.ExitError(err)
			}
			counter = counter + 1
		}
		// check for an error
		if result.Err() != nil {
			err := fmt.Errorf("query parsing error: %s\n", result.Err().Error())
			check.ExitError(err)
		}
	} else {
		check.ExitError(err)
	}

	return value, measurement, field
}
