package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type APICLient struct {
	config *Config
	client *http.Client
}

func getDefaultAPIClient() *APICLient {
	return &APICLient{
		config: &Config{url: "http://api.bharathk.in/api"},
		client: &http.Client{},
	}
}

func (a *APICLient) getFiles() (map[string][]string, error) {
	token, err := getToken()
	if err != nil {
		return nil, err
	}

	url := a.config.url + "/uploads?path=asdf"
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
	fmt.Println(respData)

	return respData, nil
}

func (a *APICLient) deleteFIle(filename string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	client := &http.Client{}

	url := a.config.url + "/uploads/" + filename

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
	return nil
}

func (a *APICLient) uploadFile(filename string) error {
	token, err := getToken()
	if err != nil {
		return err
	}

	client := &http.Client{}

	url := "http://api.bharathk.in/api/uploads"

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
	return nil
}
