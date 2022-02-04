package cmd

import (
	"github.com/NETWAYS/check_influxdb/internal/client"
	"net/url"
	"strconv"
)

type Config struct {
	Hostname     string
	Port         int
	TLS          bool
	Insecure     bool
	Token        string
	Organization string
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

	cl := client.NewClient(u.String(), c.Token, c.Organization)
	cl.Insecure = c.Insecure

	return cl
}
