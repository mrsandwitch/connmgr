package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var rootCmd = &cobra.Command{Use: "connMgr"}

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

}

func Execute() {
	rootCmd.Execute()
}
