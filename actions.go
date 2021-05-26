package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
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
