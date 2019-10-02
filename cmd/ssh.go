package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var cmdConnect = &cobra.Command{
	Use:   "c",
	Short: "Connect to a host",
	Long:  `Connect to a host`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := connect()
		if err != nil {
			os.Exit(1)
		}
	},
}

var cmdEnableSsh = &cobra.Command{
	Use:   "e",
	Short: "Enable root ssh access to a host",
	Long:  `Enable root ssh access to a host`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		err := enableRootAccess()
		if err != nil {
			os.Exit(1)
		}
	},
}

var cmdCommand = &cobra.Command{
	Use:   "cmd <command>",
	Short: "Send remote command to a host",
	Long:  `Send remote command to a host`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := command(args[0])
		if err != nil {
			os.Exit(1)
		}
	},
}

var cmdScp = &cobra.Command{
	Use:   "cp <src> <remote_dst>",
	Short: "Copy file to a host",
	Long:  `Copy file to a host`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := scp(args[0], args[1])
		if err != nil {
			os.Exit(1)
		}
	},
}

var cmdIscp = &cobra.Command{
	Use:   "icp <remote_src> <dst>",
	Short: "Copy file from a host",
	Long:  `Copy file from a host`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		err := iscp(args[0], args[1])
		if err != nil {
			os.Exit(1)
		}
	},
}

var cmdPubKey = &cobra.Command{
	Use:   "pub",
	Short: "dump public key",
	Long:  `dump public key`,
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		key, err := dumpPubKey()
		if err != nil {
			os.Exit(1)
		}
		fmt.Println(key)
	},
}

var useUser = false

func init() {
	cmdEnableSsh.Flags().BoolVarP(&useUser, "use_user", "u", false, "Use user account ")
	//rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(cmdConnect)
	rootCmd.AddCommand(cmdEnableSsh)
	rootCmd.AddCommand(cmdCommand)
	rootCmd.AddCommand(cmdScp)
	rootCmd.AddCommand(cmdIscp)
	rootCmd.AddCommand(cmdPubKey)
}

func connect() error {
	conn, err := selectSingleConnection()
	if err != nil {
		return err
	}

	fmt.Println("Connection to", conn.Hostname)

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sshCmd := fmt.Sprintf("root@%s", conn.Hostname)
	cmd := exec.Command("ssh", sshCmd, "-i", home+"/.ssh/id_rsa")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	return nil
}

func getHostKey(host string) (ssh.PublicKey, error) {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return nil, fmt.Errorf("error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hostKey == nil {
		return nil, fmt.Errorf(fmt.Sprintf("no hostkey for %s", host))
	}
	return hostKey, nil
}

func getSigner() (ssh.Signer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	key, err := ioutil.ReadFile(home + "/.ssh/id_rsa")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return signer, nil
}

func dumpPubKey() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Get public key
	pub, err := ioutil.ReadFile(home + "/.ssh/id_rsa.pub")
	if err != nil {
		return "", err
	}
	pubKey := strings.TrimSuffix(string(pub), "\n")

	return pubKey, nil
}

// reference form remote-ssh-key-setup.sh
//
func enableRootAccess() error {
	// Tell SSH to read in the output of the provided script as the password.
	// We still have to use setsid to eliminate access to a terminal and thus avoid
	// it ignoring this and asking for a password.
	err := os.Setenv("SSH_ASKPASS", "/tmp/ssh-askpass.sh")
	if err != nil {
		log.Println(err)
		return err
	}

	// Set no display, necessary for ssh to play nice with setsid and SSH_ASKPASS.
	err = os.Setenv("DISPLAY", ":0")
	if err != nil {
		log.Println(err)
		return err
	}

	conns, err := selectConnections(true)
	if err != nil {
		log.Println(err)
		return err
	}

	pubKey, err := dumpPubKey()
	if err != nil {
		log.Println(err)
		return err
	}

	for _, conn := range conns {
		// Create a temp script to echo the SSH password, used by SSH_ASKPASS
		script := fmt.Sprintf(`
			#!/bin/sh
			echo "%s"
		`, conn.Pass)

		err = ioutil.WriteFile("/tmp/ssh-askpass.sh", []byte(script), 0755)
		if err != nil {
			log.Println(err)
			return err
		}

		addr := conn.User + "@" + conn.Hostname

		// LogLevel error is to suppress the hosts warning. The others are
		// necessary if working with development servers with self-signed
		// certificates.
		sshOptLogLevel := "-oLogLevel=error"
		sshOptHostkeyCheck := "-oStrictHostKeyChecking=no"
		sshOptKnownHostFile := "-oUserKnownHostsFile=/dev/null"

		bashCmd := fmt.Sprintf("echo %s | sudo -S whoami; sudo mkdir -p /root/.ssh/ /var/services/homes/admin; sudo /bin/bash -c 'echo %s >> /root/.ssh/authorized_keys'", conn.Pass, pubKey)
		cmd := exec.Command("setsid", "ssh", "-t", sshOptLogLevel, sshOptHostkeyCheck, sshOptKnownHostFile, addr, bashCmd)

		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}

	return nil
}

func scp(src, dst string) error {
	conn, err := selectSingleConnection()
	if err != nil {
		log.Println(err)
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	remote := fmt.Sprintf("root@%s:/%s", conn.Hostname, dst)
	cmd := exec.Command("scp", "-r", "-i", home+"/.ssh/id_rsa", src, remote)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	return nil
}

func iscp(src, dst string) error {
	conn, err := selectSingleConnection()
	if err != nil {
		log.Println(err)
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	remote := fmt.Sprintf("root@%s:/%s", conn.Hostname, src)
	cmd := exec.Command("scp", "-r", "-i", home+"/.ssh/id_rsa", remote, dst)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	return nil
}

func command(cmd string) error {
	conn, err := selectSingleConnection()
	if err != nil {
		log.Println(err)
		return err
	}

	signer, err := getSigner()
	if err != nil {
		log.Println(err)
		return err
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := conn.Hostname + ":22"
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Println(err)
		return err
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		log.Fatal("Failed to run: " + err.Error())
	}
	fmt.Println(b.String())

	return nil
}
