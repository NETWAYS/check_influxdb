package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
)

func TestHealth_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "health", "--port", "9999")
	out, _ := cmd.CombinedOutput()

	actual := string(out)
	expected := "[UNKNOWN] - could not reach influxdb instance"

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

type HealthTest struct {
	name       string
	httpreturn http.HandlerFunc
	args       []string
	expected   string
}

func TestHealthCmd(t *testing.T) {
	tests := []HealthTest{
		{
			name: "health-v1-ok",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"checks":[],"message":"ready for queries and writes","name":"influxdb","status":"pass","version":"1.8.10"}`))
			},
			args:     []string{"run", "../main.go", "health"},
			expected: "[OK] - InfluxDB Status: pass\n",
		},
		{
			name: "health-v2-ok",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"influxdb", "message":"ready for queries and writes", "status":"pass", "checks":[], "version": "v2.7.1", "commit": "407fa622e9"}`))
			},
			args:     []string{"run", "../main.go", "health"},
			expected: "[OK] - InfluxDB Status: pass\n",
		},
		{
			name: "health-v2-fail",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"influxdb", "message":"Oh No!", "status":"fail", "checks":[], "version": "v2.7.1", "commit": "407fa622e9"}`))
			},
			args:     []string{"run", "../main.go", "health"},
			expected: "[CRITICAL] - InfluxDB Status: fail\nexit status 2\n",
		},
		{
			name: "health-v3-unknown",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"influxdb", "message":"", "status":"pass", "checks":[], "version": "v3.0.0", "commit": "000000"}`))
			},
			args:     []string{"run", "../main.go", "health"},
			expected: "[UNKNOWN] - InfluxDB Version 3 not supported (*errors.errorString)\nexit status 3\n",
		},
		{
			name: "health-extra-header",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				foobar := r.Header.Get("X-Foobar")
				if foobar == "Barfoo" {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"checks":[],"message":"ready for queries and writes","name":"influxdb","status":"pass","version":"1.8.10"}`))
					return
				}
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`Wrong Header!`))
			},
			args:     []string{"run", "../main.go", "--header", "X-Foobar: Barfoo", "health"},
			expected: "[OK] - InfluxDB Status: pass\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mux := http.NewServeMux()

			// Just so that the Client can establish the connection
			mux.HandleFunc("/ping/", func(w http.ResponseWriter, r *http.Request) {})
			// Add the wanted return to the Handler
			mux.HandleFunc("/health/", test.httpreturn)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			u, _ := url.Parse(ts.URL)
			cmd := exec.Command("go", append(test.args, "--port", u.Port())...)
			out, _ := cmd.CombinedOutput()

			actual := string(out)

			if actual != test.expected {
				t.Error("\nActual: ", actual, "\nExpected: ", test.expected)
			}

		})
	}
}
