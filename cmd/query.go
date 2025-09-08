package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
	v2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/cobra"
)

type QueryConfig struct {
	Organization       string
	Bucket             string
	PerfdataLabel      string
	PerfdataLabelByKey string
	FluxFile           string
	FluxString         string
	Critical           string
	Warning            string
}

var warn *check.Threshold
var crit *check.Threshold

var cliQueryConfig QueryConfig

// Check of we can convert a record's value to compare it
// to the warn/crit threshold.
func convertToFloat64(value interface{}) (float64, error) {
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
		return 0, errors.New("string value can not be evaluated")
	default:
		return 0, errors.New("unknown data type")
	}
}

func queryFluxV2(fluxQuery, url, org, token string, c *http.Client) {
	client := v2.NewClientWithOptions(url, token,
		v2.DefaultOptions().SetHTTPClient(c),
	)

	ctx, cancel := cliConfig.timeoutContext()

	defer cancel()

	var (
		perfData     perfdata.PerfdataList
		rc           int
		recordStatus int
		states       []int
	)

	queryAPI := client.QueryAPI(org)
	queryResult, queryErr := queryAPI.Query(ctx, fluxQuery)

	if queryErr != nil {
		check.ExitError(queryErr)
	}

	// Evaluate query results.
	for queryResult.Next() {
		record := queryResult.Record()
		recordValue, err := convertToFloat64(record.Value())

		if err != nil {
			continue
		}

		switch {
		case crit.DoesViolate(recordValue):
			recordStatus = 2
		case warn.DoesViolate(recordValue):
			recordStatus = 1
		default:
			recordStatus = 0
		}

		states = append(states, recordStatus)

		// Default performance data label.
		if cliQueryConfig.PerfdataLabel == "" {
			cliQueryConfig.PerfdataLabel = record.Measurement() + "." + record.Field()
		}

		// Use LabelByKey.
		if cliQueryConfig.PerfdataLabelByKey != "" {
			cliQueryConfig.PerfdataLabel = fmt.Sprint(record.ValueByKey(cliQueryConfig.PerfdataLabelByKey))
		}

		// Skip perfdata if no key was found <nil>
		if cliQueryConfig.PerfdataLabel == "<nil>" {
			continue
		}

		perfData.Add(&perfdata.Perfdata{
			Label: cliQueryConfig.PerfdataLabel,
			Value: recordValue,
			Warn:  warn,
			Crit:  crit,
		})
	}

	// When the data from the query cannot be parsed.
	if queryResult.Err() != nil {
		check.ExitRaw(check.Unknown, "query error:", queryResult.Err().Error())
	}

	switch result.WorstState(states...) {
	case 0:
		rc = check.OK
	case 1:
		rc = check.Warning
	case 2:
		rc = check.Critical
	default:
		rc = check.Unknown
	}

	// If we got perfdata we print the only the last value
	if len(perfData) >= 1 {
		check.ExitRaw(rc, "InfluxDB Query Status", perfData[len(perfData)-1].String(), "|", perfData[len(perfData)-1].String())
	}

	check.ExitRaw(rc, "InfluxDB Query Status")
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Checks one specific or multiple values from the database using flux",
	Long:  `Checks one specific or multiple values from the database using flux`,
	Run: func(_ *cobra.Command, _ []string) {
		var fluxQuery string
		var err error

		if cliQueryConfig.FluxFile == "" && cliQueryConfig.FluxString == "" {
			check.ExitError(errors.New("flux script needs to be defined with either --flux-file or --flux-string"))
		}

		crit, err = check.ParseThreshold(cliQueryConfig.Critical)
		if err != nil {
			check.ExitError(err)
		}

		warn, err = check.ParseThreshold(cliQueryConfig.Warning)
		if err != nil {
			check.ExitError(err)
		}

		// Load flux script from file.
		if cliQueryConfig.FluxFile != "" {
			fq, err := os.ReadFile(cliQueryConfig.FluxFile)

			if err != nil {
				check.ExitError(fmt.Errorf("unable to read flux file %s: %w", cliQueryConfig.FluxFile, err))
			}

			fluxQuery = string(fq)
		}

		// Load flux from CLI.
		if cliQueryConfig.FluxString != "" {
			fluxQuery = cliQueryConfig.FluxString
		}

		// Create API Client.
		c := cliConfig.NewClient()

		apiversion, versionErr := c.Version()

		if versionErr != nil {
			check.ExitError(versionErr)
		}

		// Uses the major version to determine which API to call.
		// Can be extended in the future.
		switch apiversion.MajorVersion {
		case 2:
			queryFluxV2(fluxQuery, c.URL, c.Organization, c.Token, c.Client)
		default:
			check.ExitError(fmt.Errorf("InfluxDB Version %d not supported", apiversion.MajorVersion))
		}
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
	fs := queryCmd.Flags()

	fs.StringVarP(&cliConfig.Organization, "org", "o", "",
		"The organization to use (required)")
	fs.StringVarP(&cliQueryConfig.Bucket, "bucket", "b", "",
		"The bucket to use (required)")
	fs.StringVarP(&cliQueryConfig.Critical, "critical", "c", "",
		"The critical threshold (required)")
	fs.StringVarP(&cliQueryConfig.Warning, "warning", "w", "",
		"The warning threshold (required)")

	fs.StringVarP(&cliQueryConfig.FluxFile, "flux-file", "f", "",
		"Path to flux file")
	fs.StringVarP(&cliQueryConfig.FluxString, "flux-string", "q", "",
		"Flux script as string")

	fs.StringVar(&cliQueryConfig.PerfdataLabelByKey, "perfdata-label-by-key", "",
		"Sets the label for the perfdata of the given column key for the record.\nWill skip perfdata output if the key is not found")
	fs.StringVar(&cliQueryConfig.PerfdataLabel, "perfdata-label", "",
		"Sets as custom label for the perfdata")

	queryCmd.MarkFlagsMutuallyExclusive("flux-file", "flux-string")
	queryCmd.MarkFlagsMutuallyExclusive("perfdata-label-by-key", "perfdata-label")

	_ = queryCmd.MarkFlagRequired("bucket")
	_ = queryCmd.MarkFlagRequired("org")
	_ = queryCmd.MarkFlagRequired("warning")
	_ = queryCmd.MarkFlagRequired("critical")

	fs.SortFlags = false
}
