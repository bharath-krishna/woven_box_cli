package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
)

func getToken() (map[string]string, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	token_file_path := usr.HomeDir + "/.woven_box/authn_token.json"

	jsonFile, err := os.Open(token_file_path)
	if err != nil {
		return nil, errors.New("Unauthorized. Try login command.")
	}
	defer jsonFile.Close()

	tokenrespDataa := &map[string]string{}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &tokenrespDataa)

	return *tokenrespDataa, nil
}

func listFiles() ([]string, error) {
	a := getDefaultAPIClient()
	respData, err := a.getFiles()
	if err != nil {
		return nil, err
	}

	return respData["filenames"], nil
}

func deleteFIle(filename string) error {
	a := getDefaultAPIClient()
	err := a.deleteFIle(filename)
	if err != nil {
		return err
	}

	fmt.Printf("File '%s' Deleted successfully\n", filename)
	return nil
}

func uploadFile(filename string) error {
	a := getDefaultAPIClient()
	a.uploadFile(filename)
	fmt.Printf("The file %s uploaded successfully\n", filename)
	return nil
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func UploadMultipartFile(client *http.Client, uri, key, path string, token map[string]string) (*http.Response, error) {
	body, writer := io.Pipe()

	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}

	mwriter := multipart.NewWriter(writer)
	req.Header.Add("Content-Type", mwriter.FormDataContentType())
	req.Header.Add("Authorization", "Bearer "+token["accessToken"])

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
