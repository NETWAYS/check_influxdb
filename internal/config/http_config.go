// Copyright 2016 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"bytes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"sync"
)

type closeIdler interface {
	CloseIdleConnections()
}

type TLSVersion uint16

var TLSVersions = map[string]TLSVersion{
	"TLS13": (TLSVersion)(tls.VersionTLS13),
	"TLS12": (TLSVersion)(tls.VersionTLS12),
	"TLS11": (TLSVersion)(tls.VersionTLS11),
	"TLS10": (TLSVersion)(tls.VersionTLS10),
}

// NewTLSConfig creates a new tls.Config from the given TLSConfig.
func NewTLSConfig(cfg *TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         uint16(cfg.MinVersion),
		MaxVersion:         uint16(cfg.MaxVersion),
	}

	if cfg.MaxVersion != 0 && cfg.MinVersion != 0 {
		if cfg.MaxVersion < cfg.MinVersion {
			return nil, fmt.Errorf("tls_config.max_version must be greater than or equal to tls_config.min_version if both are specified")
		}
	}

	// If a CA cert is provided then let's read it in so we can validate the
	// scrape target's certificate properly.
	if len(cfg.CAFile) > 0 {
		b, err := readCAFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}
		if !updateRootCA(tlsConfig, b) {
			return nil, fmt.Errorf("unable to use specified CA cert %s", cfg.CAFile)
		}
	}

	if len(cfg.ServerName) > 0 {
		tlsConfig.ServerName = cfg.ServerName
	}
	// If a client cert & key is provided then configure TLS config accordingly.
	if len(cfg.CertFile) > 0 && len(cfg.KeyFile) == 0 {
		return nil, fmt.Errorf("client cert file %q specified without client key file", cfg.CertFile)
	} else if len(cfg.KeyFile) > 0 && len(cfg.CertFile) == 0 {
		return nil, fmt.Errorf("client key file %q specified without client cert file", cfg.KeyFile)
	} else if len(cfg.CertFile) > 0 && len(cfg.KeyFile) > 0 {
		// Verify that client cert and key are valid.
		if _, err := cfg.getClientCertificate(nil); err != nil {
			return nil, err
		}
		tlsConfig.GetClientCertificate = cfg.getClientCertificate
	}

	return tlsConfig, nil
}

// TLSConfig configures the options for TLS connections.
type TLSConfig struct {
	// The CA cert to use for the targets.
	CAFile string `yaml:"ca_file,omitempty" json:"ca_file,omitempty"`
	// The client cert file for the targets.
	CertFile string `yaml:"cert_file,omitempty" json:"cert_file,omitempty"`
	// The client key file for the targets.
	KeyFile string `yaml:"key_file,omitempty" json:"key_file,omitempty"`
	// Used to verify the hostname for the targets.
	ServerName string `yaml:"server_name,omitempty" json:"server_name,omitempty"`
	// Disable target certificate validation.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
	// Minimum TLS version.
	MinVersion TLSVersion `yaml:"min_version,omitempty" json:"min_version,omitempty"`
	// Maximum TLS version.
	MaxVersion TLSVersion `yaml:"max_version,omitempty" json:"max_version,omitempty"`
}

// readCertAndKey reads the cert and key files from the disk.
func readCertAndKey(certFile, keyFile string) ([]byte, []byte, error) {
	certData, err := os.ReadFile(certFile)
	if err != nil {
		return nil, nil, err
	}

	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, nil, err
	}

	return certData, keyData, nil
}

// getClientCertificate reads the pair of client cert and key from disk and returns a tls.Certificate.
func (c *TLSConfig) getClientCertificate(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
	certData, keyData, err := readCertAndKey(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read specified client cert (%s) & key (%s): %s", c.CertFile, c.KeyFile, err)
	}

	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		return nil, fmt.Errorf("unable to use specified client cert (%s) & key (%s): %s", c.CertFile, c.KeyFile, err)
	}

	return &cert, nil
}

// readCAFile reads the CA cert file from disk.
func readCAFile(f string) ([]byte, error) {
	data, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("unable to load specified CA cert %s: %s", f, err)
	}
	return data, nil
}

// updateRootCA parses the given byte slice as a series of PEM encoded certificates and updates tls.Config.RootCAs.
func updateRootCA(cfg *tls.Config, b []byte) bool {
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(b) {
		return false
	}
	cfg.RootCAs = caCertPool
	return true
}

// tlsRoundTripper is a RoundTripper that updates automatically its TLS
// configuration whenever the content of the CA file changes.
type tlsRoundTripper struct {
	caFile   string
	certFile string
	keyFile  string

	// newRT returns a new RoundTripper.
	newRT func(*tls.Config) (http.RoundTripper, error)

	mtx          sync.RWMutex
	rt           http.RoundTripper
	hashCAFile   []byte
	hashCertFile []byte
	hashKeyFile  []byte
	tlsConfig    *tls.Config
}

func NewTLSRoundTripper(
	cfg *tls.Config,
	caFile, certFile, keyFile string,
	newRT func(*tls.Config) (http.RoundTripper, error),
) (http.RoundTripper, error) {
	t := &tlsRoundTripper{
		caFile:    caFile,
		certFile:  certFile,
		keyFile:   keyFile,
		newRT:     newRT,
		tlsConfig: cfg,
	}

	rt, err := t.newRT(t.tlsConfig)
	if err != nil {
		return nil, err
	}
	t.rt = rt
	_, t.hashCAFile, t.hashCertFile, t.hashKeyFile, err = t.getTLSFilesWithHash()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *tlsRoundTripper) getTLSFilesWithHash() ([]byte, []byte, []byte, []byte, error) {
	b1, err := readCAFile(t.caFile)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	h1 := sha256.Sum256(b1)

	var h2, h3 [32]byte
	if t.certFile != "" {
		b2, b3, err := readCertAndKey(t.certFile, t.keyFile)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		h2, h3 = sha256.Sum256(b2), sha256.Sum256(b3)
	}

	return b1, h1[:], h2[:], h3[:], nil
}

// RoundTrip implements the http.RoundTrip interface.
func (t *tlsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	caData, caHash, certHash, keyHash, err := t.getTLSFilesWithHash()
	if err != nil {
		return nil, err
	}

	t.mtx.RLock()
	equal := bytes.Equal(caHash[:], t.hashCAFile) &&
		bytes.Equal(certHash[:], t.hashCertFile) &&
		bytes.Equal(keyHash[:], t.hashKeyFile)
	rt := t.rt
	t.mtx.RUnlock()
	if equal {
		// The CA cert hasn't changed, use the existing RoundTripper.
		return rt.RoundTrip(req)
	}

	// Create a new RoundTripper.
	// The cert and key files are read separately by the client
	// using GetClientCertificate.
	tlsConfig := t.tlsConfig.Clone()
	if !updateRootCA(tlsConfig, caData) {
		return nil, fmt.Errorf("unable to use specified CA cert %s", t.caFile)
	}
	rt, err = t.newRT(tlsConfig)
	if err != nil {
		return nil, err
	}
	t.CloseIdleConnections()

	t.mtx.Lock()
	t.rt = rt
	t.hashCAFile = caHash[:]
	t.hashCertFile = certHash[:]
	t.hashKeyFile = keyHash[:]
	t.mtx.Unlock()

	return rt.RoundTrip(req)
}

func (t *tlsRoundTripper) CloseIdleConnections() {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	if ci, ok := t.rt.(closeIdler); ok {
		ci.CloseIdleConnections()
	}
}
