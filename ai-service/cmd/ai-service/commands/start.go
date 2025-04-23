package commands

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ory/viper"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "team6-managing.mni.thm.de/Commz/ai-service/docs"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/ai"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/auth"
	server "team6-managing.mni.thm.de/Commz/ai-service/internal/http"
	"team6-managing.mni.thm.de/Commz/ai-service/internal/utils"
)

var (
	logger = utils.GetLogger("start")
)

const (
	SWAGGER_PATH = "/swagger/"
)

func init() {
	startCmd.Flags().IntVar(&port, "port", 4245, "API server port")
	startCmd.Flags().BoolVar(&prometheus, "prometheus", false, "Enable prometheus metrics")
	startCmd.Flags().BoolVar(&swagger, "swagger", false, "Enable swagger documentation")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug log info")
	startCmd.Flags().StringVar(&ollamaUrl, "ollamaUrl", "http://localhost:11434", "Ollama URL")
	startCmd.Flags().StringVar(&gatewayUrl, "gatewayUrl", "http://localhost:4242", "Gateway URL")

	viper.BindEnv("ollamaUrl", "OLLAMA_URL")
	viper.BindPFlag("ollamaUrl", startCmd.Flags().Lookup("ollamaUrl"))

	viper.BindEnv("gatewayUrl", "GATEWAY_URL")
	viper.BindPFlag("gatewayUrl", startCmd.Flags().Lookup("gatewayUrl"))

	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the ai service",
	Run: func(cmd *cobra.Command, args []string) {
		ollamaUrl = viper.GetString("ollamaUrl")
		gatewayUrl = viper.GetString("gatewayUrl")

		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		url, err := url.Parse(ollamaUrl)
		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to parse ollama URL")
		}
		aiService := ai.New(url)
		authService := auth.New(gatewayUrl)
		router := server.New(&aiService, &authService)

		// serve generated swagger documentation
		if swagger {
			router.Router.PathPrefix(SWAGGER_PATH).Handler(httpSwagger.WrapHandler)
		}

		logger.Info().Int("port", port).Msg("Starting server")
		address := fmt.Sprintf(":%d", port)
		if err := http.ListenAndServe(address, router.Router); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	},
}
