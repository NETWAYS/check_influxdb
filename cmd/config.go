package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NETWAYS/check_influxdb/internal/client"
	"github.com/NETWAYS/check_influxdb/internal/config"
	"github.com/NETWAYS/go-check"
)

// Central Configuration for CLI
type Config struct {
	BasicAuth    string
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
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
	}

	// Using a Bearer Token for authentication
	if c.Token != "" {
		var t = config.Secret(c.Token)
		rt = config.NewAuthorizationCredentialsRoundTripper("Token", t, rt)
	}

	// Using a BasicAuth for authentication
	if c.BasicAuth != "" {
		s := strings.Split(c.BasicAuth, ":")
		if len(s) != 2 {
			check.ExitError(fmt.Errorf("specify the user name and password for server authentication <user:password>"))
		}

		var u = s[0]

		var p = config.Secret(s[1])

		rt = config.NewBasicAuthRoundTripper(u, p, "", rt)
	}

	return client.NewClient(u.String(), c.Token, c.Organization, rt)
}

// Central timeout configuration for anything that needs it
func (c *Config) timeoutContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
