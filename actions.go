package main

import (
	"bufio"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
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
	url := "http://localhost:8088/api/uploads?path=asdf"
	client := &http.Client{}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var dat map[string][]string
	if err := json.Unmarshal(body, &dat); err != nil {
		return err
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "File Name"})
	for i, name := range dat["filenames"] {
		t.AppendRow([]interface{}{i + 1, name})
	}
	t.Render()
	return nil
}

func deleteFileAction(c *cli.Context) error {
	client := &http.Client{}
	filename := c.Args().Get(0)

	url := "http://localhost:8088/api/uploads?filename=" + filename

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var datbodyBytes map[string]interface{}
		if errbodyBytes := json.Unmarshal(bodyBytes, &datbodyBytes); errbodyBytes != nil {
			panic(errbodyBytes)
		}

		return errors.New(fmt.Sprint(datbodyBytes["detail"]))
	}
	fmt.Printf("File '%s' Deleted successfully\n", filename)
	return nil
}

func uploadFileAction(c *cli.Context) error {
	client := &http.Client{}
	filename := c.Args().Get(0)
	url := "http://localhost:8088/api/uploads"

	if _, err := os.Stat(filename); err != nil {
		return errors.New(fmt.Sprintf("File '%s' does not exist in current directory.\n", filename))
	}

	resp, err := UploadMultipartFile(client, url, "uploaded_files", filename)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var datbodyBytes map[string]interface{}
		if errbodyBytes := json.Unmarshal(bodyBytes, &datbodyBytes); errbodyBytes != nil {
			panic(errbodyBytes)
		}

		return errors.New(fmt.Sprint(datbodyBytes["detail"]))
	}
	fmt.Printf("The file %s uploaded successfully\n", filename)
	return nil
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func UploadMultipartFile(client *http.Client, uri, key, path string) (*http.Response, error) {
	body, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	mwriter := multipart.NewWriter(writer)
	req.Header.Add("Content-Type", mwriter.FormDataContentType())

	errchan := make(chan error)

	go func() {
		defer close(errchan)
		defer writer.Close()
		defer mwriter.Close()

		w, err := mwriter.CreateFormFile(key, path)
		if err != nil {
			errchan <- err
			return
		}

		in, err := os.Open(path)
		if err != nil {
			errchan <- err
			return
		}
		defer in.Close()

		if written, err := io.Copy(w, in); err != nil {
			errchan <- fmt.Errorf("error copying %s (%d bytes written): %v", path, written, err)
			return
		}

		if err := mwriter.Close(); err != nil {
			errchan <- err
			return
		}
	}()

	resp, err := client.Do(req)
	merr := <-errchan

	if err != nil || merr != nil {
		return resp, fmt.Errorf("http error: %v, multipart error: %v", err, merr)
	}

	return resp, nil
}

func loginAction(c *cli.Context) error {
	client := &http.Client{}
	url := "http://localhost:3000/api/token"

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
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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

	token_file_path := usr.HomeDir + "/.woven_box/authn_token.json"
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
