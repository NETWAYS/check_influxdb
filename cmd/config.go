package cmd

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/NETWAYS/check_influxdb/internal/client"
	"github.com/NETWAYS/check_influxdb/internal/config"
	"github.com/NETWAYS/go-check"
)

type Config struct {
	Hostname     string
	CAFile       string
	CertFile     string
	KeyFile      string
	Token        string
	Organization string
	Port         int
	Insecure     bool
	Secure       bool
}

var cliConfig Config

func (c *Config) NewClient() *client.Client {
	u := url.URL{
		Scheme: "http",
		Host:   c.Hostname + ":" + strconv.Itoa(c.Port),
	}

	if c.Secure {
		u.Scheme = "https"
	}

	// Create TLS configuration for default RoundTripper
	tlsConfig, err := config.NewTLSConfig(&config.TLSConfig{
		InsecureSkipVerify: c.Insecure,
		CAFile:             c.CAFile,
		KeyFile:            c.KeyFile,
		CertFile:           c.CertFile,
	})

	if err != nil {
		check.ExitError(err)
	}

	var rt http.RoundTripper = &http.Transport{
		TLSClientConfig:       tlsConfig,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	}

	return client.NewClient(u.String(), c.Token, c.Organization, rt)
}
