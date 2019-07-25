package main

import (
	//"github.com/spf13/cobra"
	"connmgr/cmd"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

//func findFileInDir() ([]os.FileInfo, error) {
//	home, err := os.UserHomeDir()
//	if err != nil {
//		return nil, err
//	}
//
//	f, err := os.Open(home + "/.connmgr")
//	//f, err := os.Open("/home/bushyang/.connmgr")
//	if err != nil {
//		return nil, err
//	}
//	defer f.Close()
//
//	entries, err := f.Readdir(0) // 0 => no limit; read all entries
//	if err != nil {
//		return nil, err
//	}
//	return entries, nil
//}

func main() {
	cmd.Execute()
}

type Conn struct {
	Desc     string `json:"desc"`
	Hostname string `json:"hostname"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
}

func readConfig() ([]Conn, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	js, err := ioutil.ReadFile(home + "/.connmgr/conn.json")
	if err != nil {
		return nil, err
	}

	conns := []Conn{}
	err = json.Unmarshal(js, &conns)
	if err != nil {
		return nil, err
	}

	return conns, nil
}

func writeConfig(conns []Conn) error {
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

func main3() {
	conns := []Conn{
		{"415p", "bush415p.syno", "admin", "aaaaaa"},
		{"vdsm7", "10.17.19.111", "admin", "aaaaaa"},
	}

	err := writeConfig(conns)
	if err != nil {
		log.Fatal(err)
	}

	//conns2, err := readConfig()
	//if err != nil {
	//	log.Fatal(err)
	//}

	//log.Println(conns2[0])

	//cmd := exec.Command("bash", "-c", "fzf -m")
	cmd := exec.Command("fzf", "-m")
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for _, conn := range conns {
			fmt.Fprintln(in, conn.Hostname)
		}

		//err := json.NewEncoder(os.Stdin).Encode(conns2)
		//if err != nil {
		//	log.Fatal(err)
		//}
	}()

	result, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}

	//return strings.Split(string(result), "\n")

	//log.Println(string(result))
	results := strings.Split(string(result), "\n")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(results)
	fmt.Println(len(results))

	//files, err := findFileInDir()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//for _, e := range files {
	//	log.Print(e.Name())
	//}
	//fmt.Println("vim-go")
	//os.list
}

func withFilter(command string, input func(in io.WriteCloser)) []string {
	//shell := os.Getenv("SHELL")
	//if len(shell) == 0 {
	//	shell = "sh"
	//}
	//cmd := exec.Command("bash", "-c", command)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stderr = os.Stderr
	in, _ := cmd.StdinPipe()
	go func() {
		input(in)
		in.Close()
	}()
	result, _ := cmd.Output()
	return strings.Split(string(result), "\n")
}

func main2() {
	filtered := withFilter("fzf -m", func(in io.WriteCloser) {
		for i := 0; i < 1000; i++ {
			fmt.Fprintln(in, i)
		}
	})
	fmt.Println(filtered)
}

//
//
//
//
//
//
//
//
//
