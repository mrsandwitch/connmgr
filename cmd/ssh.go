package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
)

var cmdConnect = &cobra.Command{
	Use:   "c",
	Short: "Connect to a host",
	Long:  `Connect to a host`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		if err != nil {
			log.Println(err)
		}
	},
}

var cmdEnableSsh = &cobra.Command{
	Use:   "e",
	Short: "Enable ssh connection to a host",
	Long:  `Enable ssh connection to a host`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := enableSsh()
		if err != nil {
			log.Println(err)
		}
	},
}

func connect() error {
	conn, err := selectConnection()
	if err != nil {
		return err
	}

	fmt.Println("Connection to", conn.Hostname)

	sshCmd := fmt.Sprintf("root@%s", conn.Hostname)
	cmd := exec.Command("ssh", sshCmd)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	return nil
}

func enableSsh() error {

	return nil
}
