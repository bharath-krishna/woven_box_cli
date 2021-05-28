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
		fmt.Println(err)
	}
	defer jsonFile.Close()

	tokenrespDataa := &map[string]string{}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &tokenrespDataa)

	return *tokenrespDataa, nil
}

type Files struct {
	filenames []string
}

func listFiles() ([]string, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	url := "http://localhost:8088/api/uploads?path=asdf"
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token["accessToken"])

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData map[string][]string
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var bodyDaata map[string]interface{}
		if err := json.Unmarshal(body, &bodyDaata); err != nil {
			return nil, err
		}

		return nil, errors.New(fmt.Sprint(bodyDaata["detail"]))
	} else {
		if err := json.Unmarshal(body, &respData); err != nil {
			return nil, err
		}
	}

	return respData["filenames"], nil
}

func deleteFIle(filename string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	client := &http.Client{}

	url := "http://localhost:8088/api/uploads/" + filename

	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Add("Authorization", "Bearer "+token["accessToken"])

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
			return err
		}
		var datbodyBytes map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &datbodyBytes); err != nil {
			return err
		}
		return errors.New(fmt.Sprint(datbodyBytes["detail"]))
	}

	fmt.Printf("File '%s' Deleted successfully\n", filename)
	return nil
}

func uploadFile(filename string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	client := &http.Client{}

	url := "http://localhost:8088/api/uploads"

	if _, err := os.Stat(filename); err != nil {
		return errors.New(fmt.Sprintf("File '%s' does not exist in current directory.\n", filename))
	}

	resp, err := UploadMultipartFile(client, url, "uploaded_files", filename, token)
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
