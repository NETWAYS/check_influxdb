package cmd

import (
	"github.com/spf13/cobra"
	"github.com/NETWAYS/go-check"
	"os"
)

var Timeout = 30


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "check_influxdb",
	Short: "A brief description of your application",
	Long: `Long text`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		go check.HandleTimeout(Timeout)
	},
	Run: Help,
}

func Execute(version string) {
	defer check.CatchPanic()

	rootCmd.Version = version
	rootCmd.VersionTemplate()

	if err := rootCmd.Execute(); err != nil {
		check.ExitError(err)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.DisableAutoGenTag = true

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

func Help(cmd *cobra.Command, strings []string)  {
	_ = cmd.Usage()

	os.Exit(3)
}
