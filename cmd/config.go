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
			fmt.Fprintf(tw, "%-16s,", conn.Hostname)
			fmt.Fprintf(tw, "%-10s,", conn.User)
			fmt.Fprintf(tw, "%-10s,", conn.Pass)
			fmt.Fprintf(tw, "%s", conn.Desc)
			fmt.Fprintf(tw, "\n")

			_ = tw.Flush()
		}
	}, true)
	//fmt.Println(rms)

	newConns := []model.Conn{}
	rmHosts := make(map[string]struct{})
	for _, r := range rms {
		splits := strings.Split(r, ",")
		if len(splits) < 4 {
			continue
		}
		rmHosts[strings.TrimSpace(splits[0])] = struct{}{}
	}

	for _, c := range conns {
		_, ok := rmHosts[c.Hostname]
		if !ok {
			newConns = append(newConns, c)
		}
	}

	//err = writeConfig(conns)
	//if err != nil {
	//	return err
	//}

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

func selectConnection(multi bool) (*model.Conn, error) {
	conns, err := readConfig()
	if err != nil {
		return nil, err
	}
	connMap := make(map[string]model.Conn)
	for _, conn := range conns {
		connMap[conn.Hostname] = conn
	}

	sel, err := fuzzyFilter(func(in io.WriteCloser) {
		tw := tabwriter.NewWriter(in, 2, 0, 2, ' ', 0)
		for _, conn := range conns {
			fmt.Fprintf(tw, "%-16s,", conn.Hostname)
			fmt.Fprintf(tw, "%-10s,", conn.User)
			fmt.Fprintf(tw, "%-10s,", conn.Pass)
			fmt.Fprintf(tw, "%s", conn.Desc)
			fmt.Fprintf(tw, "\n")

			_ = tw.Flush()
		}
	}, multi)

	splits := strings.Split(sel[0], ",")
	if len(splits) < 4 {
		return nil, fmt.Errorf("Bad entry:[%s]", sel)
	}

	conn, ok := connMap[strings.TrimSpace(splits[0])]
	if !ok {
		return nil, fmt.Errorf("Entry not found with key:[%s]", splits[0])
	}
	return &conn, nil
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

func fuzzyFilter(input func(in io.WriteCloser), multi bool) ([]string, error) {
	var cmd *exec.Cmd
	if multi {
		cmd = exec.Command("fzf", "-m")
	} else {
		cmd = exec.Command("fzf")
	}
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

func fuzzySearch(inputs []string, multi bool) ([]string, error) {
	var cmd *exec.Cmd
	if multi {
		cmd = exec.Command("fzf", "-m")
	} else {
		cmd = exec.Command("fzf")
	}
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
