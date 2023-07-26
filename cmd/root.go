package cmd

import (
	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
	"os"
)

var Timeout = 30

var rootCmd = &cobra.Command{
	Use:   "check_influxdb",
	Short: "Icinga check plugin to check InfluxDB",
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

	pfs := rootCmd.PersistentFlags()
	pfs.StringVarP(&cliConfig.Hostname, "hostname", "H", "localhost",
		"Address of the InfluxDB instance")
	pfs.IntVarP(&cliConfig.Port, "port", "p", 8086,
		"Port of the InfluxDB instance")
	pfs.StringVarP(&cliConfig.Token, "token", "T", "",
		"Specify the token for server authenticatio")
	pfs.BoolVarP(&cliConfig.TLS, "tls", "S", false,
		"Use a HTTPS connection")
	pfs.BoolVar(&cliConfig.Insecure, "insecure", false,
		"Skip the verification of the server's TLS certificate")
	pfs.IntVarP(&Timeout, "timeout", "t", Timeout,
		"Timeout in seconds for the CheckPlugin")

	rootCmd.Flags().SortFlags = false
	pfs.SortFlags = false
}

func Help(cmd *cobra.Command, strings []string) {
	_ = cmd.Usage()

	os.Exit(3)
}
