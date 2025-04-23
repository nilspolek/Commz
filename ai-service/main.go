package main

import "team6-managing.mni.thm.de/Commz/ai-service/cmd/ai-service/commands"

// @title AI Service API
// @version 1.0
// @description This is the API for the AI service
// go:generate swag init
func main() {
	commands.Execute()
}
