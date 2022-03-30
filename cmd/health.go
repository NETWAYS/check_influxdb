package cmd

import (
	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health status of InfluxDB",
	Long: `Checks the health status of InfluxDB

The health status is:
  pass = OK
  fail = CRITICAL`,
	Run: func(cmd *cobra.Command, args []string) {
		client := cliConfig.Client()
		err := client.Connect()
		if err != nil {
			check.ExitError(err)
		}

		health, err := client.Health()
		if err != nil {
			check.ExitError(err)
		}

		var rc int
		output := health.Name + ": " + string(health.Status) + " - " + *health.Message

		switch string(health.Status) {
		case "pass":
			rc = 0
		case "fail":
			rc = 2
		default:
			rc = 3
		}

		check.ExitRaw(rc, output)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.DisableFlagsInUseLine = true
}
