package cmd

import (
	"fmt"
	"net/http"

	"github.com/NETWAYS/go-check"
	v2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/cobra"
)

func checkV2Health(url, token string, c *http.Client) (string, error) {
	client := v2.NewClientWithOptions(url, token,
		v2.DefaultOptions().SetHTTPClient(c),
	)

	ctx, cancel := cliConfig.timeoutContext()

	defer cancel()

	health, err := client.Health(ctx)

	if err != nil {
		return "", err
	}

	return string(health.Status), err
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health status of InfluxDB",
	Long: `Checks the health status of InfluxDB

API translation:
	pass = OK
	fail = CRITICAL`,
	Run: func(_ *cobra.Command, _ []string) {
		// Creating an client and connecting to the API
		c := cliConfig.NewClient()

		apiversion, versionErr := c.Version()

		if versionErr != nil {
			check.ExitError(versionErr)
		}

		var (
			rc     int
			output string
			health string
			err    error
		)

		// Uses the major version to determine which API to call.
		// Can be extended in the future.
		switch apiversion.MajorVersion {
		case 1:
			health, err = checkV2Health(c.URL, c.Token, c.Client)
		case 2:
			health, err = checkV2Health(c.URL, c.Token, c.Client)
		default:
			check.ExitError(fmt.Errorf("InfluxDB Version %d not supported", apiversion.MajorVersion))
		}

		if err != nil {
			check.ExitError(err)
		}

		// Is this flexible enough? Might be better to use strings.Contains.
		switch health {
		case "pass":
			rc = 0
		case "fail":
			rc = 2
		default:
			rc = 3
		}

		output = "InfluxDB Status: " + health

		check.ExitRaw(rc, output)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.DisableFlagsInUseLine = true
}
