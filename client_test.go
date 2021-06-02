package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"testing"
)

func TestGetToken(t *testing.T) {
	respBody := map[string]string{"accessToken": "accessToken", "refreshToken": "refreshToken"}
	token_file, err := json.MarshalIndent(respBody, "", " ")
	if err != nil {
		t.Error(err)
	}

	usr, err := user.Current()
	if err != nil {
		t.Error(err)
	}

	token_file_path := usr.HomeDir + "/.woven_box_authn_token.json"
	f, err := os.Create(token_file_path)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile(token_file_path, token_file, 0600)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Done.")

	defer f.Close()

	data, err := getToken()
	if err != nil {
		t.Error(err)
	}

	if data["accessToken"] != respBody["accessToken"] {
		t.Error("Test Failed")
	}

	if data["refreshToken"] != respBody["refreshToken"] {
		t.Error("Test Failed")
	}
	os.Remove(token_file_path)

}
