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

}

func Execute() {
	rootCmd.Execute()
}
