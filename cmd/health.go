package cmd

import (
	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Checks the health status of InfluxDB",
	Long: `Checks the health status of InfluxDB

API translation:
	pass = OK
	fail = CRITICAL`,
	Run: func(cmd *cobra.Command, args []string) {
		client := cliConfig.NewClient()
		err := client.Connect()

		if err != nil {
			check.ExitError(err)
		}

		var (
			rc     int
			output string
		)

		// Getting the preconfigured context
		ctx, cancel := cliConfig.timeoutContext()
		defer cancel()

		health, err := client.Client.Health(ctx)

		if err != nil {
			check.ExitError(err)
		}

		switch string(health.Status) {
		case "pass":
			rc = 0
		case "fail":
			rc = 2
		default:
			rc = 3
		}

		output = health.Name + ": " + string(health.Status) + " - " + *health.Message

		check.ExitRaw(rc, output)
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.DisableFlagsInUseLine = true
}
