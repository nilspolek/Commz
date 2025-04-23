package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "media-service",
	Short: "Media service is a service that allows users to save medias like Pictures",
}

var (
	port            int
	prometheus      bool
	swagger         bool
	debug           bool
	minioUrl        string
	gatewayUrl      string
	accessKeyID     string
	secretAccessKey string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
