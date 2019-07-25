package cmd

import (
	"github.com/spf13/cobra"
	//"os"
)

var rootCmd = &cobra.Command{Use: "connMgr"}

func init() {
	//rootCmd.PersistentFlags().StringVarP(&pluginServerUrl, "plugin_server_url", "p", "", "url of the plugin server")
	//rootCmd.PersistentFlags().StringVarP(&vcUrl, "vc_host", "", "", "hostname/ip of the vcenter")

	//rootCmd.AddCommand(cmdExtension)
	//rootCmd.AddCommand(cmdLun)
	//rootCmd.AddCommand(cmdDatastore)
	//rootCmd.AddCommand(cmdVcenter)
	//rootCmd.AddCommand(cmdPlugin)
	rootCmd.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdList)
	rootCmd.AddCommand(cmdRemove)
}

func Execute() {
	rootCmd.Execute()
}
