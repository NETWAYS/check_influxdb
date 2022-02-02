package cmd

import (
	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
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

		rc := 3
		output := health.Name + " " + string(health.Status) + " - " + *health.Message

		switch health.Status {
		case "pass":
			rc = 0
		case "fail":
			rc = 2
		default:
			rc = 2
		}

		// pass = 0
		// fail = 2
		p := perfdata.PerfdataList{
			{Label: "status", Value: rc},
		}

		check.ExitRaw(rc, output, "|", p.String())
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
	healthCmd.DisableFlagsInUseLine = true
}
