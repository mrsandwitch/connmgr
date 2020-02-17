package cmd

import (
	"connmgr/model"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
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

		key, err := findUnusedKey()
		if err != nil {
			os.Exit(1)
		}

		conn := model.Conn{
			Key:      key,
			Hostname: args[0],
			User:     user,
			Pass:     pass,
			Desc:     desc,
		}

		err = entryAdd(conn)
		if err != nil {
			os.Exit(1)
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
			os.Exit(1)
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
			os.Exit(1)
		}
	},
}

var cmdBackup = &cobra.Command{
	Use:   "back",
	Short: "Backup connection config",
	Long:  `Backup connection config`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := backup()
		if err != nil {
			os.Exit(1)
		}
	},
}

var cmdVimEdit = &cobra.Command{
	Use:   "vim",
	Short: "Use vim to edit config",
	Long:  `Use vim to edit config`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := vimEdit()
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdAdd)
	rootCmd.AddCommand(cmdList)
	rootCmd.AddCommand(cmdRemove)
	rootCmd.AddCommand(cmdBackup)
	rootCmd.AddCommand(cmdVimEdit)
}

const DEFAULT_CONF_PATH = "/.connmgr/conn.json"
const DEFAULT_BACKUP_CONF_PATH = "/.connmgr/conn.json.bkp"

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/root" + DEFAULT_CONF_PATH
	}

	return home + DEFAULT_CONF_PATH
}

func getBackupConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/root" + DEFAULT_BACKUP_CONF_PATH
	}

	return home + DEFAULT_BACKUP_CONF_PATH
}

func findUnusedKey() (string, error) {
	conns, err := readConfig()
	if err != nil {
		return "", err
	}

	m := make(map[string]struct{})
	for _, conn := range conns {
		m[conn.Key] = struct{}{}
	}

	for i := 1; i < len(conns)+20; i++ {
		key := "node-" + strconv.Itoa(i)

		_, ok := m[key]
		if ok {
			continue
		}

		return key, nil
	}

	return "", fmt.Errorf("Failed to find unused key")
}

func entryUpdate(entry model.Conn) error {
	conns, err := readConfig()
	if err != nil {
		return err
	}

	for ix, conn := range conns {
		if conn.Key == entry.Key {
			conns[ix] = entry
		}
	}

	err = writeConfig(conns, getConfigPath())
	if err != nil {
		return err
	}

	return nil
}

func entryAdd(entry model.Conn) error {
	conns, err := readConfig()
	if err != nil {
		return err
	}

	conns = append(conns, entry)

	err = writeConfig(conns, getConfigPath())
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
		fmt.Fprintf(tw, "%-8s", conn.Key)
		fmt.Fprintf(tw, "%-16s", conn.Hostname)
		fmt.Fprintf(tw, "%-10.10s\t", conn.User)
		fmt.Fprintf(tw, "%-10.10s\t", conn.Pass)
		fmt.Fprintf(tw, "%s", conn.Desc)
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

	rms, err := selectConnections(true)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, r := range rms {
		fmt.Printf("%s(%s) is removed\n", r.Key, r.Hostname)
	}

	rmHosts := make(map[string]struct{})
	for _, r := range rms {
		rmHosts[r.Key] = struct{}{}
	}

	newConns := []model.Conn{}
	for _, c := range conns {
		if _, ok := rmHosts[c.Key]; ok {
			// skip removed conn
			continue
		}
		newConns = append(newConns, c)
	}

	err = writeConfig(newConns, getConfigPath())
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func readConfig() ([]model.Conn, error) {
	js, err := ioutil.ReadFile(getConfigPath())
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

func selectConnections(multi bool) ([]model.Conn, error) {
	conns, err := readConfig()
	if err != nil {
		return nil, err
	}

	connMap := make(map[string]model.Conn)
	for _, conn := range conns {
		connMap[conn.Key] = conn
	}

	sels, err := fuzzyFilter(func(in io.WriteCloser) {
		tw := tabwriter.NewWriter(in, 2, 0, 2, ' ', 0)
		for _, conn := range conns {
			fmt.Fprintf(tw, "%-8s,", conn.Key)
			fmt.Fprintf(tw, "%-16s,", conn.Hostname)
			fmt.Fprintf(tw, "%-10.10s\t,", conn.User)
			fmt.Fprintf(tw, "%-10.10s\t,", conn.Pass)
			fmt.Fprintf(tw, "%s", conn.Desc)
			fmt.Fprintf(tw, "\n")

			_ = tw.Flush()
		}
	}, multi)

	selConns := []model.Conn{}

	for _, sel := range sels {
		splits := strings.Split(sel, ",")
		if len(splits) <= 1 {
			continue
		} else if len(splits) < 5 {
			log.Printf("Bad entry:[%s]", sel)
			continue
		}

		key := strings.TrimSpace(splits[0])
		conn, ok := connMap[key]
		if !ok {
			log.Printf("Entry not found with key:[%s]", key)
			continue
		}

		selConns = append(selConns, conn)
	}

	if len(selConns) < 1 {
		err = fmt.Errorf("No connection selected")
		log.Println(err)
		return nil, err
	}

	return selConns, nil
}

func selectSingleConnection() (*model.Conn, error) {
	conns, err := selectConnections(false)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &conns[0], nil
}

func writeConfig(conns []model.Conn, dst string) error {
	data, err := json.MarshalIndent(conns, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func backup() error {
	conns, err := readConfig()
	if err != nil {
		log.Println(err)
		return err
	}

	err = writeConfig(conns, getBackupConfigPath())
	if err != nil {
		log.Println(err)
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

func vimEdit() error {
	cmd := exec.Command("vim", getConfigPath())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return nil
}
