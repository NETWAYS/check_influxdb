package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/NETWAYS/go-check"
	// "github.com/NETWAYS/go-check/perfdata"
	// "github.com/NETWAYS/go-check/result"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/spf13/cobra"
)

type QueryV1Config struct {
	Database    string
	Critical    string
	Warning     string
	QueryFile   string
	QueryString string
}

var warnv1 *check.Threshold
var critv1 *check.Threshold

var cliQueryV1Config QueryV1Config

func queryV1(query, database string, influxClient client.Client) {
	// var (
	// 	perfData perfdata.PerfdataList
	// 	rc       int
	// 	states   []int
	// )

	q := client.NewQuery(query, database, "")

	response, errQuery := influxClient.Query(q)

	if errQuery != nil {
		check.ExitError(errQuery)
	}

	for _, result := range response.Results {
		for _, series := range result.Series {
			for _, row := range series.Values {
				fmt.Println(row)
			}
		}
	}
}

var queryv1Cmd = &cobra.Command{
	Use:   "queryv1",
	Short: "Checks one specific or multiple values from the database using InfluxQL",
	Long:  `Checks one specific or multiple values from the database using InfluxQL`,
	Run: func(_ *cobra.Command, _ []string) {
		var dbQuery string
		var err error

		if cliQueryV1Config.QueryFile == "" && cliQueryV1Config.QueryString == "" {
			check.ExitError(errors.New("Query needs to be defined with either --query-file or --query-string"))
		}

		crit, err = check.ParseThreshold(cliQueryV1Config.Critical)
		if err != nil {
			check.ExitError(err)
		}

		warn, err = check.ParseThreshold(cliQueryV1Config.Warning)
		if err != nil {
			check.ExitError(err)
		}

		// Load query from file.
		if cliQueryV1Config.QueryFile != "" {
			fq, err := os.ReadFile(cliQueryV1Config.QueryFile)

			if err != nil {
				check.ExitError(fmt.Errorf("unable to read query file %s: %w", cliQueryV1Config.QueryFile, err))
			}

			dbQuery = string(fq)
		}

		// Load query from CLI.
		if cliQueryV1Config.QueryString != "" {
			dbQuery = cliQueryV1Config.QueryString
		}

		c := cliConfig.NewV1Client()

		queryV1(dbQuery, cliQueryV1Config.Database, c)
	},
}

func init() {
	rootCmd.AddCommand(queryv1Cmd)
	fs := queryv1Cmd.Flags()

	fs.StringVarP(&cliQueryV1Config.Database, "database", "d", "",
		"The database to use (required)")
	fs.StringVarP(&cliQueryV1Config.Critical, "critical", "c", "",
		"The critical threshold (required)")
	fs.StringVarP(&cliQueryV1Config.Warning, "warning", "w", "",
		"The warning threshold (required)")

	fs.StringVarP(&cliQueryV1Config.QueryFile, "query-file", "f", "",
		"Path to query file")
	fs.StringVarP(&cliQueryV1Config.QueryString, "query-string", "q", "",
		"Query script as string")

	// fs.StringVar(&cliQueryConfig.PerfdataLabelByKey, "perfdata-label-by-key", "",
	// 	"Sets the label for the perfdata of the given column key for the record.\nWill skip perfdata output if the key is not found")
	// fs.StringVar(&cliQueryConfig.PerfdataLabel, "perfdata-label", "",
	// 	"Sets as custom label for the perfdata")

	queryv1Cmd.MarkFlagsMutuallyExclusive("query-file", "query-string")
	// queryCmd.MarkFlagsMutuallyExclusive("perfdata-label-by-key", "perfdata-label")

	_ = queryv1Cmd.MarkFlagRequired("database")
	_ = queryv1Cmd.MarkFlagRequired("warning")
	_ = queryv1Cmd.MarkFlagRequired("critical")

	fs.SortFlags = false
}
