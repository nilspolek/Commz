package main

import commands "team6-managing.mni.thm.de/Commz/media-service/cmd/media-service"

// @title Media Service API
// @version 1.0
// @description This is the API for the Media service
// go:generate swag init
func main() {
	commands.Execute()
}
