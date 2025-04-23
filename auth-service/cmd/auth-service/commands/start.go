package commands

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/auth"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/storage"
	"team6-managing.mni.thm.de/Commz/auth-service/internal/utils"

	_ "team6-managing.mni.thm.de/Commz/auth-service/docs"
	server "team6-managing.mni.thm.de/Commz/auth-service/internal/http"
)

var (
	logger = utils.GetLogger("start")
)

func init() {

	startCmd.Flags().IntVar(&port, "port", 4244, "API server port")
	startCmd.Flags().BoolVar(&prometheus, "prometheus", false, "Enable prometheus metrics")
	startCmd.Flags().BoolVar(&swagger, "swagger", false, "Enable swagger documentation")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug log info")
	startCmd.Flags().StringVar(&jwtKey, "jwt", "your-secret-key", "Set a JWT token")
	startCmd.Flags().StringVar(&mongoURI, "mongo-uri", "mongodb://localhost:27017", "MongoDB URI")
	viper.BindPFlag("server.port", startCmd.Flags().Lookup("port"))

	viper.BindEnv("mongo-uri", "MONGO_URI")
	viper.BindPFlag("mongo-uri", startCmd.Flags().Lookup("mongo-uri"))

	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the chat service",
	Run: func(cmd *cobra.Command, args []string) {
		mongoURI = viper.GetString("mongo-uri")

		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		utils.JwtKey = []byte(useEnvIfNull(jwtKey, "JWT_KEY", "your-secret-key"))

		storage, err := storage.NewMongoDBStorage(mongoURI)

		if err != nil {
			logger.Fatal().Err(err).Msg("Failed to connect to MongoDB")
			return
		}

		chatService := auth.New(storage)
		router := server.New(&chatService)

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

func useEnvIfNull(val, key string, defaultVal ...string) string {
	if val == "" || len(defaultVal) == 1 && defaultVal[0] == val {
		if os.Getenv(key) != "" {
			return os.Getenv(key)
		}
		if len(defaultVal) == 1 {
			return defaultVal[0]
		}
		return ""
	}
	return val
}
