package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chat-service",
	Short: "Chat service is a service that allows users to send messages to each other",
}

var (
	port       int
	prometheus bool
	swagger    bool
	debug      bool
	mongoURI   string
	gatewayUrl string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
