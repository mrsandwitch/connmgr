package cmd

import (
	"github.com/spf13/cobra"
	//"os"
	"log"
)

var rootCmd = &cobra.Command{Use: "connMgr"}

func init() {
	//rootCmd.PersistentFlags().StringVarP(&pluginServerUrl, "plugin_server_url", "p", "", "url of the plugin server")
	//rootCmd.PersistentFlags().StringVarP(&vcUrl, "vc_host", "", "", "hostname/ip of the vcenter")

	//log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetFlags(log.LstdFlags | log.Llongfile)

	rootCmd.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdList)
	rootCmd.AddCommand(cmdRemove)
	rootCmd.AddCommand(cmdConnect)
	rootCmd.AddCommand(cmdEnableSsh)
	rootCmd.AddCommand(cmdCommand)
	rootCmd.AddCommand(cmdScp)
	rootCmd.AddCommand(cmdIscp)
	rootCmd.AddCommand(cmdBackup)
}

func Execute() {
	rootCmd.Execute()
}
