package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func GetChat(chatId uuid.UUID, cookie string) (*Chat, error) {
	chatService := viper.GetString("chatService")
	request, err := http.NewRequest("GET", chatService+"/"+chatId.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Cookie", cookie)
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("can not get chat from chat service")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat service returned status code: %d", response.StatusCode)
	}

	var chat Chat
	err = json.NewDecoder(response.Body).Decode(&chat)
	if err != nil {
		return nil, fmt.Errorf("get chat response invalid from chat service")
	}
	return &chat, nil
}

func SendMessage(chatId uuid.UUID, content string, cookie string) (*Message, error) {
	type SendMessageRequest struct {
		Message string `json:"message"`
	}

	body := SendMessageRequest{Message: content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	chatService := viper.GetString("chatService")
	request, err := http.NewRequest("POST", chatService+"/"+chatId.String()+"/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Cookie", cookie)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("can send message to chat service")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat service returned status code: %d", response.StatusCode)
	}

	var chat Message
	err = json.NewDecoder(response.Body).Decode(&chat)
	if err != nil {
		return nil, fmt.Errorf("send message response invalid from chat service")
	}
	return &chat, nil
}

func UpdateMessage(messageId uuid.UUID, content string, cookie string) (*Message, error) {
	type SendMessageRequest struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}

	body := SendMessageRequest{Message: content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	chatService := viper.GetString("chatService")
	request, err := http.NewRequest("PUT", chatService+"/messages/"+messageId.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Cookie", cookie)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("can update message to chat service")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat service returned status code: %d", response.StatusCode)
	}

	var chat Message
	err = json.NewDecoder(response.Body).Decode(&chat)
	if err != nil {
		return nil, fmt.Errorf("update message response invalid from chat service")
	}
	return &chat, nil
}

func DeleteMessage(messageId uuid.UUID, content string, cookie string) (*Message, error) {
	type SendMessageRequest struct {
		Message string `json:"message"`
		ID      string `json:"id"`
	}

	body := SendMessageRequest{Message: content}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	chatService := viper.GetString("chatService")
	request, err := http.NewRequest("DELETE", chatService+"/messages/"+messageId.String(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Cookie", cookie)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("cant delete message to chat service")
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chat service returned status code: %d", response.StatusCode)
	}

	var chat Message
	err = json.NewDecoder(response.Body).Decode(&chat)
	if err != nil {
		return nil, fmt.Errorf("delete message response invalid from chat service")
	}
	return &chat, nil

}
