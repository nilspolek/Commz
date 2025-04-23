package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gateway",
	Short: "The gateway is responsible for connecting the different micro services",
}

var (
	port            int
	debug           bool
	metrics         bool
	authServiceUrl  string
	chatServiceUrl  string
	aiServiceUrl    string
	mediaServiceUrl string
	mongoURI        string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
