# check_influxdb

Icinga check plugin to check InfluxDB v2

## Usage

### Health

Checks the health status of an InfluxDB instance

````
Usage:
  check_influxdb health

Flags:
  -h, --help   help for health

Global Flags:
  -H, --hostname string   Address of the InfluxDB instance (default "localhost")
      --insecure          Allow use of self signed certificates when using SSL
  -p, --port int          Port of the InfluxDB instance (default 8086)
  -t, --timeout int       Timeout for the check (default 30)
  -S, --tls               Use secure connection
  -T, --token string      The token which allows access to the API
````

````
$ check_influxdb health
OK - influxdb: pass - ready for queries and writes
````

### Query

Checks a specific value from the database. The query must return only ONE value.

````
Usage:
  check_influxdb query [flags]

Flags:
  -o, --org string               The organization which will be used
  -b, --bucket string            The bucket where time series data is stored
  -q, --raw-query string         An InfluxQL query which will be performed. Note: Only ONE value result will be evaluated
      --start duration           Specifies a start time range for your query. (default -1h0m0s)
      --end duration             Specifies the end of a time range for your query.
  -m, --measurement string       The data stored in the associated fields, e.g. 'disk'
  -f, --field string             The key-value pair that records metadata and the actual data value. (default "value")
  -a, --aggregation string       Function that returns an aggregated value across a set of points.
                                 Viable values are 'mean', 'median', 'last' (default "last")
  -F, --filter stringArray       Add a key=value filter to the query, e.g. 'hostname=example.com'
      --raw-filter stringArray   A fully customizable filter which will be added to the query.
                                 e.g. 'filter(fn: (r) => r["hostname"] == "example.com")'
  -c, --critical string          The critical threshold for a value (default "500")
  -w, --warning string           The warning threshold for a value (default "1000")
  -h, --help                     help for query

Global Flags:
  -H, --hostname string   Address of the InfluxDB instance (default "localhost")
      --insecure          Allow use of self signed certificates when using SSL
  -p, --port int          Port of the InfluxDB instance (default 8086)
  -t, --timeout int       Timeout for the check (default 30)
  -S, --tls               Use secure connection
  -T, --token string      The token which allows access to the API
````

````
 $ check_influxdb query -H 'example.host' --port 443 --token 'example_token' -S \
                        --org "example_org" \
                        --bucket 'example_bucket' \
                        --measurement "example_measurement" \
                        --filter 'hostname=example.com' \
                        --raw-filter 'filter(fn: (r) => r["example_key"] == "exmaple_value")' \
                        --aggregation 'median' \
                        --start -1h --end 0
 CRITICAL - value is: 12623000000 | value=12623000000;1000;500
````

## Further Documentation

[Query data with Flux](https://docs.influxdata.com/influxdb/v2.1/query-data/flux/)

## License

Copyright (c) 2022 [NETWAYS GmbH](mailto:info@netways.de)

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public
License as published by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied
warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program. If not,
see [gnu.org/licenses](https://www.gnu.org/licenses/).