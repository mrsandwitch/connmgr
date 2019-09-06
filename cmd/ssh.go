package cmd

import (
	"bufio"
	"bytes"
	"connmgr/model"
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

func connect() error {
	conn, err := selectConnection(false)
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

// reference form remote-ssh-key-setup.sh
//
func enableRootAccess() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Get public key
	pub, err := ioutil.ReadFile(home + "/.ssh/id_rsa.pub")
	if err != nil {
		return err
	}
	pubKey := strings.TrimSuffix(string(pub), "\n")

	// Tell SSH to read in the output of the provided script as the password.
	// We still have to use setsid to eliminate access to a terminal and thus avoid
	// it ignoring this and asking for a password.
	err = os.Setenv("SSH_ASKPASS", "/tmp/ssh-askpass.sh")
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

	conn := model.Conn{
		Hostname: "bush415p.syno",
		User:     "admin",
		Pass:     "aaaaaa",
		Desc:     "",
	}
	//conn, err := selectConnection(true)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}

	script := fmt.Sprintf(`
		#!/bin/sh
		echo "%s"
	`, conn.Pass)

	err = ioutil.WriteFile("/tmp/ssh-askpass.sh", []byte(script), 0755)
	if err != nil {
		log.Println(err)
		return err
	}
	//return nil

	addr := conn.User + "@" + conn.Hostname
	//cmd := exec.Command("ssh", addr, "su", "cat >> /root/.ssh/authorized_keys")
	//cmd := exec.Command("ssh", addr, "sudo -i")
	//cmd.Stdin = keyFile

	//bashCmd := fmt.Sprintf("echo %s | sudo mkdir -p /root/.ssh/ /var/services/homes/admin; echo %s >> /root/.ssh/authorized_keys2", conn.Pass, string(pub))
	//bashCmd := fmt.Sprintf("echo hello1 > /tmp/test1")
	//bashCmd := fmt.Sprintf("mkdir -p /root/.ssh/ /var/services/homes/admin; echo %s >> /root/.ssh/authorized_keys2", conn.Pass, string(pub))
	//bashCmd := fmt.Sprintf("echo aaaaaa | sudo -S whoami; echo hello2 > /tmp/test1")
	bashCmd := fmt.Sprintf("echo %s | sudo -S whoami; sudo mkdir -p /root/.ssh/ /var/services/homes/admin; sudo /bin/bash -c 'echo %s >> /root/.ssh/authorized_keys'", conn.Pass, pubKey)
	//bashCmd := fmt.Sprintf("echo %s | sudo -S whoami; sudo mkdir -p /root/.ssh/ /var/services/homes/admin;", conn.Pass)
	//bashCmd2 := fmt.Sprintf("echo %s | sudo -S whoami; sudo /bin/bash -c 'echo %s >> /root/.ssh/authorized_keys2'", conn.Pass, pubKey)

	//cmd := exec.Command("setsid", "ssh", "-t", "-oLogLevel=error", "-oStrictHostKeyChecking=no", "-oUserKnownHostsFile=/dev/null", addr, "bash", "-c", "echo aaaaaa; echo hello2 > /tmp/test1")
	//cmd := exec.Command("setsid", "ssh", "-t", "-oLogLevel=error", "-oStrictHostKeyChecking=no", "-oUserKnownHostsFile=/dev/null", addr, "bash", "-c", bashCmd)
	//cmd := exec.Command("setsid", "ssh", "-t", "-oLogLevel=error", "-oStrictHostKeyChecking=no", "-oUserKnownHostsFile=/dev/null", addr, bashCmd)
	cmd := exec.Command("setsid", "ssh", "-t", "-oLogLevel=error", "-oStrictHostKeyChecking=no", "-oUserKnownHostsFile=/dev/null", addr, bashCmd)

	out, err := cmd.CombinedOutput()
	//out, err := cmd.Output()
	//err = cmd.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))

	return nil
}

func scp() error {
	return nil
}

func command(cmd string) error {
	conn, err := selectConnection(false)
	if err != nil {
		log.Println(err)
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Println(err)
		return err
	}

	key, err := ioutil.ReadFile(home + "/.ssh/id_rsa")
	if err != nil {
		log.Println(err)
		return err
	}

	signer, err := ssh.ParsePrivateKey(key)
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
