package cmd

import (
	//"bytes"
	"connmgr/model"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
)

var cmdAdd = &cobra.Command{
	Use:   "a <hostname> <user> <pass> [desc]",
	Short: "Add a connection entry",
	Long:  `Add a connection entry`,
	Args:  cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		var desc, user, pass string
		if len(args) > 3 {
			desc = args[3]
		}

		user = args[1]
		pass = args[2]
		if pass == "admin" {
			user, pass = pass, user
		}

		conn := model.Conn{
			Hostname: args[0],
			User:     user,
			Pass:     pass,
			Desc:     desc,
		}

		err := entryAdd(conn)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s is added\n", args[0])
	},
}

var cmdList = &cobra.Command{
	Use:   "l",
	Short: "list all connections",
	Long:  `List all connections`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := entryList()
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println("Vcenter: " + strings.Join(args, " "))
	},
}

var cmdRemove = &cobra.Command{
	Use:   "r",
	Short: "remove a connections",
	Long:  `Remove a connection`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := entryRemove()
		if err != nil {
			log.Fatal(err)
		}
	},
}

//var cmdEntryAdd = &cobra.Command{
//	Use:   "add <ip> <username> <password>",
//	Short: "Add a vcenter entry",
//	Args:  cobra.MinimumNArgs(3),
//	Run: func(cmd *cobra.Command, args []string) {
//		entry := vmware.Entry{
//			Ip:       args[0],
//			UserName: args[1],
//			PassWord: args[2],
//		}
//
//		err := vcEntryAdd(entry)
//		if err != nil {
//			fmt.Printf("Failed to add vcenter entry. %v\n", err)
//		}
//	},
//}

//var cmdEntryRemove = &cobra.Command{
//	Use:   "remove <ip>",
//	Short: "Remove a vcenter entry",
//	Args:  cobra.MinimumNArgs(1),
//	Run: func(cmd *cobra.Command, args []string) {
//		err := vcEntryRemove(args[0])
//		if err != nil {
//			fmt.Printf("Failed to remove vcenter entry. %v\n", err)
//		}
//	},
//}

func init() {
	//cmdAdd.AddCommand(cmdryAdd)
}

func entryAdd(entry model.Conn) error {
	conns, err := readConfig()
	if err != nil {
		return err
	}

	conns = append(conns, entry)

	err = writeConfig(conns)
	if err != nil {
		return err
	}

	return nil
}

func entryList() error {
	conns, err := readConfig()
	if err != nil {
		return err
	}
	tw := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	for _, conn := range conns {
		fmt.Fprintf(tw, "%-16s\t", conn.Hostname)
		fmt.Fprintf(tw, "%-10s\t", conn.User)
		fmt.Fprintf(tw, "%-10s\t", conn.Pass)
		fmt.Fprintf(tw, "%s\t", conn.Desc)
		fmt.Fprintf(tw, "\n")

		_ = tw.Flush()
	}

	return nil
}

func entryRemove() error {
	conns, err := readConfig()
	if err != nil {
		return err
	}

	rms, err := fuzzyFilter(func(in io.WriteCloser) {
		tw := tabwriter.NewWriter(in, 2, 0, 2, ' ', 0)
		for _, conn := range conns {
			fmt.Fprintf(tw, "%-16s\t", conn.Hostname)
			fmt.Fprintf(tw, "%-10s\t", conn.User)
			fmt.Fprintf(tw, "%-10s\t", conn.Pass)
			fmt.Fprintf(tw, "%s\t", conn.Desc)
			fmt.Fprintf(tw, "\n")

			_ = tw.Flush()
		}
	})
	fmt.Println(rms)

	return nil
}

func readConfig() ([]model.Conn, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	js, err := ioutil.ReadFile(home + "/.connmgr/conn.json")
	if err != nil {
		return nil, err
	}

	conns := []model.Conn{}
	err = json.Unmarshal(js, &conns)
	if err != nil {
		return nil, err
	}

	return conns, nil
}

func writeConfig(conns []model.Conn) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(conns, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(home+"/.connmgr/conn.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func fuzzyFilter(input func(in io.WriteCloser)) ([]string, error) {
	cmd := exec.Command("fzf", "-m")
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		input(in)
		in.Close()
	}()
	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return strings.Split(string(result), "\n"), nil
}

func fuzzySearch(inputs []string) ([]string, error) {
	cmd := exec.Command("fzf", "-m")
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		for _, input := range inputs {
			fmt.Fprintln(in, input)
		}
	}()

	result, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	results := strings.Split(string(result), "\n")
	if err != nil {
		return nil, err
	}
	return results, nil
}
