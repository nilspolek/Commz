package commands

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	server "team6-managing.mni.thm.de/Commz/gateway/internal/http"
	"team6-managing.mni.thm.de/Commz/gateway/internal/utils"
)

var (
	logger = utils.GetLogger("start")
)

func init() {
	startCmd.Flags().IntVar(&port, "port", 4242, "API server port")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug log info")
	startCmd.Flags().BoolVar(&metrics, "metrics", true, "Enable metrics")

	startCmd.Flags().StringVar(&mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB URI")
	startCmd.Flags().StringVar(&chatServiceUrl, "chat-service", "http://localhost:4243", "Chat service URL")
	startCmd.Flags().StringVar(&authServiceUrl, "auth-service", "http://localhost:4244", "Auth service URL")
	startCmd.Flags().StringVar(&aiServiceUrl, "ai-service", "http://localhost:4245", "AI service URL")
	startCmd.Flags().StringVar(&mediaServiceUrl, "media-service", "http://localhost:4246", "Media service URL")

	viper.BindPFlag("server.port", startCmd.Flags().Lookup("port"))
	viper.BindPFlag("authService", startCmd.Flags().Lookup("auth-service"))
	viper.BindPFlag("chatService", startCmd.Flags().Lookup("chat-service"))
	viper.BindPFlag("aiService", startCmd.Flags().Lookup("ai-service"))
	viper.BindPFlag("mediaService", startCmd.Flags().Lookup("media-service"))
	viper.BindPFlag("mongo-uri", startCmd.Flags().Lookup("mongo-uri"))

	viper.BindEnv("mongo-uri", "MONGO_URI")
	viper.BindEnv("chatService", "CHAT_SERVICE_URL")
	viper.BindEnv("authService", "AUTH_SERVICE_URL")
	viper.BindEnv("aiService", "AI_SERVICE_URL")
	viper.BindEnv("mediaService", "MEDIA_SERVICE_URL")

	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the gateway",
	Run: func(cmd *cobra.Command, args []string) {

		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		chatServiceUrl := viper.GetString("chatService")
		authServiceUrl := viper.GetString("authService")
		aiServiceUrl := viper.GetString("aiService")
		mediaService := viper.GetString("mediaService")
		mongoURI := viper.GetString("mongo-uri")

		router, err := server.New(chatServiceUrl, authServiceUrl, aiServiceUrl, mediaService, mongoURI)

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to parse auth or chat service urls")
		}

		logger.Info().Int("port", port).Msg("Starting server")
		address := fmt.Sprintf(":%d", port)

		cors := cors.New(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173"},
			AllowCredentials: true,
			AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE"},
		})
		handler := cors.Handler(router.Router)

		if metrics {
			logger.Info().Msg("Started metrics on /metrics")
			router.Router.Handle("/metrics", promhttp.Handler())
		}

		if err := http.ListenAndServe(address, handler); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	},
}
