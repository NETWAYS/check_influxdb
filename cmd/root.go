package cmd

import (
	"os"

	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
)

var Timeout = 30

var rootCmd = &cobra.Command{
	Use:   "check_influxdb",
	Short: "An Icinga check plugin to check InfluxDB",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		go check.HandleTimeout(Timeout)
	},
	Run: Usage,
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
	pfs.BoolVarP(&cliConfig.Secure, "secure", "s", false,
		"Use a HTTPS connection")
	pfs.StringVarP(&cliConfig.Token, "token", "T", "",
		"Token for server authentication")
	pfs.StringVarP(&cliConfig.CAFile, "ca-file", "", "",
		"Specify the CA File for TLS authentication")
	pfs.StringVarP(&cliConfig.CertFile, "cert-file", "", "",
		"Specify the Certificate File for TLS authentication")
	pfs.StringVarP(&cliConfig.KeyFile, "key-file", "", "",
		"Specify the Key File for TLS authentication")
	pfs.BoolVarP(&cliConfig.Insecure, "insecure", "i", false,
		"Skip the verification of the server's TLS certificate")
	pfs.IntVarP(&Timeout, "timeout", "t", Timeout,
		"Timeout in seconds for the CheckPlugin")

	rootCmd.Flags().SortFlags = false
	pfs.SortFlags = false
}

func Usage(cmd *cobra.Command, _ []string) {
	_ = cmd.Usage()

	os.Exit(3)
}
