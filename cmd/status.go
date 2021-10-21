package cmd

import (
	"context"
	"fmt"
	"github.com/influxdata/influxdb-client-go/v2"
	"github.com/spf13/cobra"
	"log"
)

// statusCmd represents the health command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		userName := "neuerUser"
		password := "12345678"

		foo := fmt.Sprintf("%s:%s", userName, password)
		fmt.Println(foo)

		//ctx := context.Background()
		client := influxdb2.NewClient("http://10.211.55.94:8086/", fmt.Sprintf("%s:%s",userName, password))
		defer client.Close()

		//h, err := client.Health(ctx)
		//if err != nil {
		//
		//}
		//
		//fmt.Println(h.Status)
		//fmt.Println(*h.Message)

		result, err := client.QueryAPI("Netways").QueryRaw(context.Background(), `buckets()`, nil)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.DisableFlagsInUseLine = true


}
