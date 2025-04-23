package main

import "github.com/nilspolek/DevOps/Chat/cmd/chat-service/commands"

// @title Chat Service API
// @version 1.0
// @description This is the API for the Chat service
// go:generate swag init
func main() {
	commands.Execute()
}
