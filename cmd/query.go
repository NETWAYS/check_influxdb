package cmd

import (
	"fmt"
	"github.com/NETWAYS/check_influxdb/internal/client"
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

type QueryConfig struct {
	RawQuery    string
	Bucket      string
	RangeStart  time.Duration
	RangeEnd    time.Duration
	Measurement string
	Field       string
	RawFilter   []string
	Filter      []string
	Aggregation string
	Critical    string
	Warning     string
}

var cliQueryConfig QueryConfig

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Checks a specific value from the database",
	Long: `Checks a specific value from the database. The query must return only ONE value.
`,
	Example: ``,
	Run: func(cmd *cobra.Command, args []string) {
		c := cliConfig.Client()
		err := c.Connect()
		if err != nil {
			check.ExitError(err)
		}

		query := cliQueryConfig.RawQuery

		if query == "" {
			query, err = cliQueryConfig.BuildQuery()
			if err != nil {
				check.ExitError(err)
			}
		}

		res, err := c.GetSingleQueryResult(query)
		if err != nil {
			check.ExitError(err)
		}

		result, err := client.AssertFloat64(res.Value())
		if err != nil {
			check.ExitError(err)
		}

		var rc int

		crit, err := check.ParseThreshold(cliQueryConfig.Critical)
		if err != nil {
			check.ExitError(err)
		}

		warn, err := check.ParseThreshold(cliQueryConfig.Warning)
		if err != nil {
			check.ExitError(err)
		}

		if crit.DoesViolate(result) {
			rc = 2
		} else if warn.DoesViolate(result) {
			rc = 1
		} else {
			rc = 0
		}

		output := fmt.Sprintf("value is: %s", check.FormatFloat(result))

		p := perfdata.PerfdataList{
			{Label: "value", Value: result,
				Warn: warn,
				Crit: crit,
			},
		}

		check.ExitRaw(rc, output, "|", p.String())
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
	fs := queryCmd.Flags()

	fs.StringVarP(&cliConfig.Organization, "org", "o", "",
		"The organization which will be used")
	_ = queryCmd.MarkFlagRequired("org")
	fs.StringVarP(&cliQueryConfig.Bucket, "bucket", "b", "",
		"The bucket where time series data is stored")
	_ = queryCmd.MarkFlagRequired("bucket")
	fs.StringVarP(&cliQueryConfig.RawQuery, "raw-query", "q", "",
		"An InfluxQL query which will be performed. Note: Only ONE value result will be evaluated")
	fs.DurationVar(&cliQueryConfig.RangeStart, "start", -time.Hour,
		"Specifies a start time range for your query.")
	fs.DurationVar(&cliQueryConfig.RangeEnd, "end", 0,
		"Specifies the end of a time range for your query.")
	fs.StringVarP(&cliQueryConfig.Measurement, "measurement", "m", "",
		"The data stored in the associated fields, e.g. 'disk'")
	fs.StringVarP(&cliQueryConfig.Field, "field", "f", "value",
		"The key-value pair that records metadata and the actual data value.")
	fs.StringVarP(&cliQueryConfig.Aggregation, "aggregation", "a", "last",
		"Function that returns an aggregated value across a set of points.\nViable values are 'mean', 'median', 'last'")
	fs.StringArrayVarP(&cliQueryConfig.Filter, "filter", "F", []string{},
		"Add a key=value filter to the query, e.g. 'hostname=example.com'")
	fs.StringArrayVar(&cliQueryConfig.RawFilter, "raw-filter", []string{},
		"A fully customizable filter which will be added to the query.\ne.g. 'filter(fn: (r) => r[\"hostname\"] == \"example.com\")'")
	fs.StringVarP(&cliQueryConfig.Critical, "critical", "c", "500",
		"The critical threshold for a value")
	fs.StringVarP(&cliQueryConfig.Warning, "warning", "w", "1000",
		"The warning threshold for a value")

	fs.SortFlags = false

	// In icinga2 beispiel Hostname, service, metric abfragen
}

func (q *QueryConfig) BuildQuery() (string, error) {
	query := fmt.Sprintf(
		`from(bucket: "%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "%s")
		|> filter(fn: (r) => r["_field"] == "%s")`,
		q.Bucket,
		q.RangeStart,
		q.RangeEnd,
		q.Measurement,
		q.Field,
	)

	for _, rawFilter := range q.RawFilter {
		query += "\n |> " + rawFilter
	}

	for _, filter := range q.Filter {
		pair := strings.SplitN(filter, "=", 2)
		if len(pair) < 2 {
			return "", fmt.Errorf("filter must be 'key=value'")
		}

		query += "\n"
		query += fmt.Sprintf(`|> filter(fn: (r) => r["%s"] == "%s")`,
			pair[0],
			pair[1],
		)
	}

	switch q.Aggregation {
	case "median", "mean", "last":
		query += "\n |> " + q.Aggregation + "()"
	default:
		return "", fmt.Errorf("unknown aggreation function")
	}

	return query, nil
}
