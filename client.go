package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
