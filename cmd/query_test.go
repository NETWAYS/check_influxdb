package cmd

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os/exec"
	"strings"
	"testing"
)

func TestQuery_ConnectionRefused(t *testing.T) {

	cmd := exec.Command("go", "run", "../main.go", "query", "--port", "9999", "-o", "org", "-b", "bucket", "-T", "token", "-f", "../testdata/mem.flux", "-w", "0", "-c", "1")
	out, _ := cmd.CombinedOutput()

	actual := string(out)
	expected := "[UNKNOWN] - could not reach influxdb instance"

	if !strings.Contains(actual, expected) {
		t.Error("\nActual: ", actual, "\nExpected: ", expected)
	}
}

type QueryTest struct {
	name       string
	httpreturn http.HandlerFunc
	args       []string
	expected   string
}

func TestQueryCmd(t *testing.T) {
	tests := []QueryTest{
		{
			name: "query-ok",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,123,active,mem,influx
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,124,active,mem,influx
,_result,1,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,125,available,mem,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "-f", "../testdata/mem.flux", "-w", "200", "-c", "500"},
			expected: "[OK]",
		},
		{
			name: "query-warning",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,222,active,mem,influx
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,223,active,mem,influx
,_result,1,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,125,available,mem,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "-f", "../testdata/mem.flux", "-w", "200", "-c", "500"},
			expected: "[WARNING]",
		},
		{
			name: "query-critical",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,501,active,mem,influx
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,502,active,mem,influx
,_result,1,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,125,available,mem,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "-f", "../testdata/mem.flux", "-w", "200", "-c", "500"},
			expected: "[CRITICAL]",
		},
		{
			name: "query-file-output",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,501,active,mem,influx
,_result,0,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,502,active,mem,influx
,_result,1,2023-08-25T13:56:37.00917165Z,2023-08-25T14:56:37.00917165Z,2023-08-25T14:00:00Z,125,available,mem,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "--flux-file", "../testdata/mem.flux", "-w", "200", "-c", "500"},
			expected: "[CRITICAL] - InfluxDB Query Status | mem.active=501;200;500 mem.active=502;200;500 mem.active=125;200;500",
		},
		{
			name: "query-string-output",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,_result,0,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:00:00Z,0.07822905321070156,usage_user,cpu,cpu-total,influx
,_result,0,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,0.04436126250005433,usage_user,cpu,cpu-total,influx
,_result,1,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:00:00Z,0.07822905321070156,usage_user,cpu,cpu0,influx
,_result,1,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,0.04436126250005433,usage_user,cpu,cpu0,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "--flux-string", "from(bucket:\"monitor\")|>range(start:-1h)", "-w", "1", "-c", "2"},
			expected: "[OK] - InfluxDB Query Status | cpu.usage_user=0.078;1;2 cpu.usage_user=0.044;1;2 cpu.usage_user=0.078;1;2 cpu.usage_user=0.044;1;2",
		},
		{
			name: "query-perfdata",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,_result,0,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:00:00Z,0.07822905321070156,usage_user,cpu,cpu-total,influx
,_result,1,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,0.04436126250005433,usage_user,cpu,cpu0,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "--flux-string", "from(bucket:\"monitor\")|>range(start:-1h)", "-w", "1", "-c", "2", "--perfdata-label", "foobar"},
			expected: "InfluxDB Query Status | foobar=0.078;1;2 foobar=0.044;1;2",
		},
		{
			name: "query-perfdata-by-key",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,_result,0,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:00:00Z,0.07822905321070156,usage_user,cpu,cpu-total,influx
,_result,1,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,0.04436126250005433,usage_user,cpu,cpu0,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "--flux-string", "from(bucket:\"monitor\")|>range(start:-1h)", "-w", "1", "-c", "2", "--perfdata-label-by-key", "host"},
			expected: "InfluxDB Query Status | influx=0.078;1;2 influx=0.044;1;2",
		},
		{
			name: "query-perfdata-by-key-is-nil",
			httpreturn: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,_result,0,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:00:00Z,0.07822905321070156,usage_user,cpu,cpu-total,influx
,_result,1,2023-08-28T12:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,2023-08-28T13:03:46.777288171Z,0.04436126250005433,usage_user,cpu,cpu0,influx`,
				))
			},
			args:     []string{"run", "../main.go", "query", "-o", "org", "-b", "bucket", "-T", "token", "--flux-string", "from(bucket:\"monitor\")|>range(start:-1h)", "-w", "1", "-c", "2", "--perfdata-label-by-key", "foobar"},
			expected: "InfluxDB Query Status |",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mux := http.NewServeMux()

			// Just so that the Client can establish the connection
			mux.HandleFunc("/ping/", func(w http.ResponseWriter, r *http.Request) {})

			mux.HandleFunc("/health/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"name":"influxdb", "message":"", "status":"pass", "checks":[], "version": "v2.7.1", "commit": ""}`))
			})

			// Add the wanted return to the Handler
			mux.HandleFunc("/api/", test.httpreturn)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			u, _ := url.Parse(ts.URL)
			cmd := exec.Command("go", append(test.args, "--port", u.Port())...)
			out, _ := cmd.CombinedOutput()

			actual := string(out)

			if !strings.Contains(actual, test.expected) {
				t.Error("\nActual: ", actual, "\nExpected: ", test.expected)
			}

		})
	}
}
