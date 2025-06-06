# check_influxdb

Icinga check plugin to check InfluxDB

## Usage

### Health

Supported InfluxDB versions:

* v1
* v2

Checks the health status of an InfluxDB instance.

```
Usage:
  check_influxdb health

Flags:
  -h, --help   help for health

Global Flags:
  -H, --hostname string    Address of the InfluxDB instance (CHECK_INFLUXDB_HOSTNAME) (default "localhost")
  -p, --port int           Port of the InfluxDB instance (default 8086)
  -s, --secure             Use a HTTPS connection
  -T, --token string       Token for server authentication (CHECK_INFLUXDB_TOKEN)
  -u, --user string        Specify the user name and password for server authentication <user:password> (CHECK_INFLUXDB_BASICAUTH)
      --ca-file string     Specify the CA File for TLS authentication (CHECK_INFLUXDB_CA_FILE)
      --cert-file string   Specify the Certificate File for TLS authentication (CHECK_INFLUXDB_CERT_FILE)
      --key-file string    Specify the Key File for TLS authentication (CHECK_INFLUXDB_KEY_FILE)
  -i, --insecure           Skip the verification of the server's TLS certificate
  -t, --timeout int        Timeout in seconds for the CheckPlugin (default 30)
      --header strings     Additional HTTP header to include in the request. Can be used multiple times.
                           Keys and values are separated by a colon (--header "X-Custom: example").
  -h, --help               help for check_influxdb
  -v, --version            version for check_influxdb
```

The check plugin respects the environment variables HTTP_PROXY, HTTPS_PROXY and NO_PROXY.

Various flags can be set with environment variables, refer to the help to see which flags.

Examples:

```
check_influxdb health
[OK] - influxdb: pass - ready for queries and writes
```

### Query

Supported InfluxDB versions:

* v2

```
Usage:
  check_influxdb query [flags]

Flags:
  -o, --org string                     The organization to use (required)
  -b, --bucket string                  The bucket to use (required)
  -c, --critical string                The critical threshold (required)
  -w, --warning string                 The warning threshold (required)
  -f, --flux-file string               Path to flux file
  -q, --flux-string string             Flux script as string
      --perfdata-label-by-key string   Sets the label for the perfdata of the given column key for the record.
                                       Will skip perfdata output if the key is not found
      --perfdata-label string          Sets as custom label for the perfdata
  -h, --help                           help for query
```

Examples:

```
check_influxdb query --token "${INFLUX_TOKEN}" --org influx --bucket telegraf \
--warning 1 --critical 2 \
--flux-string 'from(bucket:"monitor")|>range(start:-1h)|>filter(fn:(r)=>r["_measurement"]=="cpu")|>filter(fn:(r)=>r["_field"]=="usage_user")|>aggregateWindow(every:1h,fn:mean)'

[CRITICAL] - InfluxDB Query Status | influxdb.cpu.usage_user=0.078;1;2 influxdb.cpu.usage_user=0.04;1;2 influxdb.cpu.usage_user=0.078;1;2 influxdb.cpu.usage_user=0.04;1;2
exit status 2
```

```
check_influxdb query --token "${INFLUX_TOKEN}" --org influx --bucket telegraf \
--warning 50 --critical 100 \
--flux-file mem.flux

[WARNING] - InfluxDB Query Status | influxdb.mem.active=45.5;50;100 influxdb.mem.active=68.9;50;100
exit status 1
```

# License

Copyright (c) 2022 [NETWAYS GmbH](mailto:info@netways.de)

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public
License as published by the Free Software Foundation, either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not,
see [gnu.org/licenses](https://www.gnu.org/licenses/).
