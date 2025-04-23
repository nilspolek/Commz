package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/viper"
)

func VerifyToken(token string) (*User, error) {

	type VerifyTokenRequest struct {
		Token string `json:"token"`
	}

	// get user id from auth service
	request := VerifyTokenRequest{
		Token: token,
	}

	jsonBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("internal sever error: failed to marshal verify token request")
	}
	bodyReader := bytes.NewReader(jsonBody)

	authService := viper.GetString("authService")
	response, err := http.Post(authService+"/verify", "application/json", bodyReader)
	if err != nil {
		return nil, fmt.Errorf("token invalid")
	}

	var user User
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("verify token response invalid from auth service")
	}
	return &user, nil
}
