package cmd

import (
	"github.com/NETWAYS/go-check"
	"github.com/spf13/cobra"
)

type QueryConfig struct {
	Organization string
	Bucket       string
	Query        string
	Average      uint
	Critical     uint
	Warning      uint
}

var cliQueryConfig QueryConfig

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Check for a specific value",
	Long:  `Checks for a specific value from the database. The query has to return only ONE value.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		//if Client.User != "" && Client.Pass == "" {
		//	err := fmt.Errorf("please specify a password")
		//	check.ExitError(err)
		//} else if Client.User == "" && Client.Pass != "" {
		//	err := fmt.Errorf("please specify a username")
		//	check.ExitError(err)
		//} else if Client.Token != "" && (Client.User != "" || Client.Pass != "") {
		//	err := fmt.Errorf(`either specify a "token" or "username and password"`)
		//	check.ExitError(err)
		//} else if Client.Token == "" && Client.User == "" && Client.Pass == "" {
		//	err := fmt.Errorf(`either specify a "token" or "username and password"`)
		//	check.ExitError(err)
		//}
	},
	Run: func(cmd *cobra.Command, args []string) {
		//address := Client.Proto + "://" + Client.Host + ":" + strconv.Itoa(Client.Port)

		client := cliConfig.Client()
		err := client.Connect()
		if err != nil {
			check.ExitError(err)
		}

		if client.Token != "" {
			//client := influxdb2.NewClient(address, Client.Token)
			check.Exitf(client.ExecuteQuery(client.Client, cliQueryConfig.Query))
		} else if client.Token == "" && (client.User != "" && client.Pass != "") {
			//client := influxdb2.NewClient(address, "")
			check.Exitf(client.ExecuteQuery(client.Client, cliQueryConfig.Query))
		}
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
	fs := queryCmd.Flags()

	fs.StringVarP(&cliQueryConfig.Organization, "org", "", "", "")
	_ = queryCmd.MarkFlagRequired("org")
	fs.StringVarP(&cliQueryConfig.Bucket, "bucket", "", "", "")
	fs.StringVarP(&cliQueryConfig.Query, "query", "", "", "")
	_ = queryCmd.MarkFlagRequired("query")
	fs.UintVarP(&cliQueryConfig.Average, "average", "", 0, "If 0 only the last value will be used")
	fs.UintVarP(&cliQueryConfig.Critical, "critical", "", 500, "")
	fs.UintVarP(&cliQueryConfig.Warning, "warning", "", 1000, "")

	fs.SortFlags = false

}
