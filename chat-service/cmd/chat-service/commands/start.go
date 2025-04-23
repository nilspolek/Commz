package commands

import (
	"fmt"
	"net/http"

	_ "github.com/nilspolek/DevOps/Chat/docs"
	"github.com/nilspolek/DevOps/Chat/internal/ai"
	"github.com/nilspolek/DevOps/Chat/internal/auth"
	"github.com/nilspolek/DevOps/Chat/internal/chat"
	server "github.com/nilspolek/DevOps/Chat/internal/http"
	"github.com/nilspolek/DevOps/Chat/internal/storage"
	"github.com/nilspolek/DevOps/Chat/internal/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

var (
	logger = utils.GetLogger("start")
)

func init() {
	startCmd.Flags().IntVar(&port, "port", 4243, "API server port")
	startCmd.Flags().BoolVar(&prometheus, "prometheus", false, "Enable prometheus metrics")
	startCmd.Flags().BoolVar(&swagger, "swagger", false, "Enable swagger documentation")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug log info")
	startCmd.Flags().StringVar(&mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB URI")
	startCmd.Flags().StringVar(&gatewayUrl, "gatewayUrl", "http://localhost:4242", "Gateway URL")

	viper.BindPFlag("server.port", startCmd.Flags().Lookup("port"))
	viper.BindEnv("mongo-uri", "MONGO_URI")
	viper.BindPFlag("mongo-uri", startCmd.Flags().Lookup("mongo-uri"))

	viper.BindEnv("gatewayUrl", "GATEWAY_URL")
	viper.BindPFlag("gatewayUrl", startCmd.Flags().Lookup("gatewayUrl"))

	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the chat service",
	Run: func(cmd *cobra.Command, args []string) {

		mongoURI = viper.GetString("mongo-uri")
		gatewayUrl = viper.GetString("gatewayUrl")

		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		storage, err := storage.NewMongoDBStorage(mongoURI)

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect to MongoDB")
			return
		}

		aiService := ai.New(gatewayUrl)
		authService := auth.New(gatewayUrl)
		chatService := chat.New(storage, &authService, &aiService)
		router := server.New(&chatService, &authService)

		// serve generated swagger documentation
		if swagger {
			router.Router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
		}

		logger.Info().Int("port", port).Msg("Starting server")
		address := fmt.Sprintf(":%d", port)
		if err := http.ListenAndServe(address, router.Router); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	},
}
