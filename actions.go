package main

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"
	"syscall"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
)

func listFilesAction(c *cli.Context) error {
	filenames, err := listFiles()
	if err != nil {
		c.App.Run([]string{c.App.Name, "help"})
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "File Name"})
	for i, name := range filenames {
		t.AppendRow([]interface{}{i + 1, name})
	}
	t.Render()
	return nil
}

func deleteFileAction(c *cli.Context) error {
	err := deleteFIle(c.Args().Get(0))
	if err != nil {
		return err
	}
	return nil
}

func uploadFileAction(c *cli.Context) error {
	err := uploadFile(c.Args().Get(0))
	if err != nil {
		return err
	}
	return nil
}

func loginAction(c *cli.Context) error {
	client := &http.Client{}
	url := "http://woven-box.bharathk.in/api/token"

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter email")
	fmt.Println("---------------------")
	email_input, _ := reader.ReadString('\n')
	email := strings.Replace(email_input, "\n", "", -1)

	fmt.Println("Enter password")
	fmt.Println("---------------------")

	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	password := string(bytePassword)
	password = strings.TrimSpace(password)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	fmt.Println("Logging in...")

	basic_token := b64.StdEncoding.EncodeToString([]byte(email + ":" + password))
	req.Header.Add("Authorization", "Basic "+basic_token)
	resp, err := client.Do(req)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		var datbodyBytes map[string]interface{}
		if errbodyBytes := json.Unmarshal(bodyBytes, &datbodyBytes); errbodyBytes != nil {
			fmt.Println(errbodyBytes)
		}

		return errors.New(fmt.Sprint(datbodyBytes["error"]))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var respBody map[string]string
	if err := json.Unmarshal(body, &respBody); err != nil {
		return err
	}

	token_file, err := json.MarshalIndent(respBody, "", " ")
	if err != nil {
		return err
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	token_file_path := usr.HomeDir + "/.woven_box_authn_token.json"
	f, err := os.Create(token_file_path)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(token_file_path, token_file, 0600)
	if err != nil {
		return err
	}
	fmt.Println("Done.")

	defer f.Close()

	return nil
}
