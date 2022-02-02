package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/NETWAYS/go-check"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"strconv"
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
	Username string
	Password string
	Token    string
	Insecure bool
	Client   influxdb2.Client
	LogLevel uint
	Average  int
	Critical int
	Warning  int
}

func NewClient(url, username, password, token string) *Client {
	return &Client{
		Url:      url,
		Username: username,
		Password: password,
		Token:    token,
		Insecure: false,
	}
}

func (c *Client) Connect() error {
	cfg := influxdb2.NewClientWithOptions(
		c.Url,
		c.Token,
		influxdb2.DefaultOptions().SetTLSConfig(&tls.Config{
			InsecureSkipVerify: c.Insecure,
		}).SetLogLevel(c.LogLevel))

	c.Client = cfg

	return nil
}

func (c *Client) ExecuteQuery(ifClient influxdb2.Client, query string) (rc int, output string) {
	var value int
	var field string
	var measurement string

	if c.Token != "" {
		value, measurement, field = c.parseQueryValueType(ifClient, query)
		//value, measurement, field = c.parseQueryValueType(ifClient, c.Query)
		fmt.Println(value)
		fmt.Println(ByteToHumanread(value, 2))

		defer ifClient.Close()
	} else {
		err := ifClient.UsersAPI().SignIn(context.Background(), c.User, c.Pass)
		if err != nil {
			check.ExitError(err)
		}

		value, measurement, field = c.parseQueryValueType(ifClient, query)
		//value, measurement, field = c.parseQueryValueType(ifClient, c.Query)

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

func (c *Client) ExecQuery(ifClient influxdb2.Client, query string) {
	//queryApi := ifClient.QueryAPI(c.Organization)

	if c.Token == "" {

	} else {

	}

	return
}

func (c *Client) parseQueryValueType(ifClient influxdb2.Client, query string) (value int, measurement string, field string) {
	queryApi := ifClient.QueryAPI(c.Organization)

	fmt.Println(query)

	//result, err := queryApi.Query(context.Background(), c.Query)
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

// https://mrwaggel.be/post/golang-human-readable-byte-sizes/
func ByteToHumanread(length int, decimals int) (out string) {
	var unit string
	var i int
	var remainder int

	// Get whole number, and the remainder for decimals
	if length > TB {
		unit = "TB"
		i = length / TB
		remainder = length - (i * TB)
	} else if length > GB {
		unit = "GB"
		i = length / GB
		remainder = length - (i * GB)
	} else if length > MB {
		unit = "MB"
		i = length / MB
		remainder = length - (i * MB)
	} else if length > KB {
		unit = "KB"
		i = length / KB
		remainder = length - (i * KB)
	} else {
		return strconv.Itoa(int(length)) + " B"
	}

	if decimals == 0 {
		return strconv.Itoa(int(i)) + " " + unit
	}

	// This is to calculate missing leading zeroes
	width := 0
	if remainder > GB {
		width = 12
	} else if remainder > MB {
		width = 9
	} else if remainder > KB {
		width = 6
	} else {
		width = 3
	}

	// Insert missing leading zeroes
	remainderString := strconv.Itoa(int(remainder))
	for iter := len(remainderString); iter < width; iter++ {
		remainderString = "0" + remainderString
	}
	if decimals > len(remainderString) {
		decimals = len(remainderString)
	}

	return fmt.Sprintf("%d.%s%s", i, remainderString[:decimals], unit)
}
