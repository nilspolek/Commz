package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "auth-service",
	Short: "Auth service is a service that allows users to authenticate and authorize themselves",
}

var (
	port       int
	prometheus bool
	swagger    bool
	debug      bool
	mongoURI   string
	jwtKey     string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
