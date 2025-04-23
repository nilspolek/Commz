package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-service",
	Short: "Ai service is a service that allows users to interact with Ollama",
}

var (
	port       int
	prometheus bool
	swagger    bool
	debug      bool
	ollamaUrl  string
	gatewayUrl string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
