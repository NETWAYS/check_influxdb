package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
	"github.com/spf13/cobra"
)

type QueryConfig struct {
	RawQuery      string
	Bucket        string
	RangeStart    time.Duration
	RangeEnd      time.Duration
	Measurement   string
	Field         string
	RawFilter     []string
	Filter        []string
	Aggregation   string
	PerfdataLabel string
	ValueByKey    string
	Critical      string
	Warning       string
	Verbose       bool
}

var cliQueryConfig QueryConfig

// Converts return from client into float64.
func assertFloat64(value interface{}) (float64, error) {
	switch res := value.(type) {
	case float64:
		return res, nil
	case float32:
		return float64(res), nil
	case int64:
		return float64(res), nil
	case int32:
		return float64(res), nil
	case int16:
		return float64(res), nil
	case int8:
		return float64(res), nil
	case int:
		return float64(res), nil
	case uint64:
		return float64(res), nil
	case uint32:
		return float64(res), nil
	case uint16:
		return float64(res), nil
	case uint8:
		return float64(res), nil
	case uint:
		return float64(res), nil
	case string:
		return 0, fmt.Errorf("string value can not be evaluated")
	default:
		return 0, fmt.Errorf("unknown data type")
	}
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Checks one specific or multiple values from the database",
	Long: `Checks one specific or multiple values from the database. It's possible to set custom labels for
the perfdata via '--perfdata-label', or set the key name from the database via '--value-by-key'.
IMPORTANT: the filter, aggregation and raw-filter parameters has a specific evaluation order, which is:
	1. --bucket
	2. --start --end
	3. --measurement
	4. --field
	5. --filter (can be repeated)
	6. --raw-filter (can be repeated)
	7. --aggregation

Use the '--verbose' parameter to see the query which will be evaluated.`,
	Example: ``,
	Run: func(cmd *cobra.Command, args []string) {
		c := cliConfig.NewClient()
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
			if cliQueryConfig.Verbose {
				fmt.Println(query)
			}
		}

		res, err := c.GetQueryRecords(query)
		if err != nil {
			check.ExitError(err)
		}

		var (
			output  string
			rc      int
			perf    perfdata.PerfdataList
			states  []int
			summary float64
		)

		for _, result := range res {
			record, err := assertFloat64(result.Value())
			if err != nil {
				check.ExitError(err)
			}

			summary += record

			crit, err := check.ParseThreshold(cliQueryConfig.Critical)
			if err != nil {
				check.ExitError(err)
			}

			warn, err := check.ParseThreshold(cliQueryConfig.Warning)
			if err != nil {
				check.ExitError(err)
			}

			if crit.DoesViolate(record) {
				rc = 2
			} else if warn.DoesViolate(record) {
				rc = 1
			} else {
				rc = 0
			}

			output += fmt.Sprintf("value is: %s; ", check.FormatFloat(record))

			states = append(states, rc)

			if cliQueryConfig.ValueByKey != "" {
				cliQueryConfig.PerfdataLabel = fmt.Sprint(result.ValueByKey(cliQueryConfig.ValueByKey))
			}

			// TODO: Validate characters in label. Might need filtering
			p := perfdata.Perfdata{
				Label: cliQueryConfig.PerfdataLabel,
				Value: record,
				Warn:  warn,
				Crit:  crit,
			}

			perf.Add(&p)
		}

		if result.WorstState(states...) == 0 {
			output = "All values are OK"
		}

		check.ExitRaw(result.WorstState(states...), output, "|", perf.String())
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
		"Function that returns an aggregated value across a set of points.\nViable values are 'mean', 'median', 'last', 'max', 'sum'")
	fs.StringArrayVarP(&cliQueryConfig.Filter, "filter", "F", []string{},
		"Add a key=value filter to the query, e.g. 'hostname=example.com'")
	fs.StringArrayVar(&cliQueryConfig.RawFilter, "raw-filter", []string{},
		"A fully customizable filter which will be added to the query.\ne.g. 'filter(fn: (r) => r[\"hostname\"] == \"example.com\")'")
	fs.StringVar(&cliQueryConfig.ValueByKey, "value-by-key", "",
		"Sets the label for the perfdata of the given column key for the record.\ne.g. --value-by-key 'hostname', which will be rendered out of the database to 'exmaple.int.host.com'")
	fs.StringVar(&cliQueryConfig.PerfdataLabel, "perfdata-label", "",
		"Sets as custom label for the perfdata")
	fs.BoolVarP(&cliQueryConfig.Verbose, "verbose", "v", false,
		"Display verbose output")
	fs.StringVarP(&cliQueryConfig.Critical, "critical", "c", "500",
		"The critical threshold for a value")
	fs.StringVarP(&cliQueryConfig.Warning, "warning", "w", "1000",
		"The warning threshold for a value")

	fs.SortFlags = false
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

	for _, rawFilter := range q.RawFilter {
		query += "\n |> " + rawFilter
	}

	switch q.Aggregation {
	case "median", "mean", "last", "sum", "max":
		query += "\n |> " + q.Aggregation + "()"
	default:
		return "", fmt.Errorf("unknown aggreation function")
	}

	return query, nil
}
