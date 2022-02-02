package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// TODO: Metriken werden von der Clientlib nicht zur Verfügung gestellt ...
var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("metrics called")
	},
}

func init() {
	rootCmd.AddCommand(metricsCmd)
}
