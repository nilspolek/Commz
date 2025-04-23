package utils

import (
	"bytes"
	"encoding/json"
	"net/http"

	"golang.org/x/net/websocket"
)

func PostRequest[T any, R any](url string, requestBody T) (*R, error) {
	// Marshal request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, NewError("failed to marshal request", http.StatusInternalServerError)
	}
	bodyReader := bytes.NewReader(jsonBody)

	// Make POST request
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, NewError("failed to create request", http.StatusInternalServerError)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://localhost/")
	response, err := client.Do(req)
	if err != nil {
		return nil, NewError("failed to make POST request", http.StatusBadRequest)
	}
	defer response.Body.Close()

	// Check response status
	if response.StatusCode >= 400 {
		return nil, NewError("server returned error", response.StatusCode)
	}

	// Decode response
	var result R
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, NewError("failed to decode response", http.StatusInternalServerError)
	}

	return &result, nil
}

func GetRequest[T any](url string) (*T, error) {
	// Make POST request
	response, err := http.Get(url)
	if err != nil {
		return nil, NewError("failed to make POST request", http.StatusBadRequest)
	}
	defer response.Body.Close()

	// Check response status
	if response.StatusCode >= 400 {
		return nil, NewError("server returned error", response.StatusCode)
	}

	// Decode response
	var result T
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, NewError("failed to decode response", http.StatusInternalServerError)
	}

	return &result, nil
}

func WebsocketRequest[T any, R any](url string, request T, responseFunction func(response R)) error {
	socket, err := websocket.Dial(url, "", "http://localhost/")
	if err != nil {
		return NewError("failed to connect to websocket", http.StatusBadRequest)
	}
	defer socket.Close()
	err = websocket.JSON.Send(socket, request)

	if err != nil {
		return err
	}

	for {

		var result R
		err = websocket.JSON.Receive(socket, &result)
		if err != nil {
			return nil
		}
		responseFunction(result)
	}
}
