package cmd

import (
	"check_influxdb/internal/client"
	"net/url"
	"strconv"
)

type Config struct {
	Hostname string
	Port     int
	TLS      bool
	Username string
	Password string
	Insecure bool
	Token    string
}

var cliConfig Config

func (c *Config) Client() *client.Client {
	u := url.URL{
		Scheme: "http",
		Host:   c.Hostname + ":" + strconv.Itoa(c.Port),
	}

	if c.TLS {
		u.Scheme = "https"
	}

	cl := client.NewClient(u.String(), c.Username, c.Password, c.Token)
	cl.Insecure = c.Insecure

	return cl
}
