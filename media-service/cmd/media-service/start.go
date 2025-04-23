package commands

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "team6-managing.mni.thm.de/Commz/media-service/docs"
	"team6-managing.mni.thm.de/Commz/media-service/internal/auth"
	server "team6-managing.mni.thm.de/Commz/media-service/internal/http"
	"team6-managing.mni.thm.de/Commz/media-service/internal/media"
	"team6-managing.mni.thm.de/Commz/media-service/internal/utils"
)

var (
	logger = utils.GetLogger("start")
)

const (
	SWAGGER_PATH = "/swagger/"
)

func init() {
	startCmd.Flags().IntVar(&port, "port", 4246, "API server port")
	startCmd.Flags().BoolVar(&prometheus, "prometheus", false, "Enable prometheus metrics")
	startCmd.Flags().BoolVar(&swagger, "swagger", false, "Enable swagger documentation")
	startCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug log info")
	startCmd.Flags().StringVar(&gatewayUrl, "gatewayUrl", "http://localhost:4242", "Gateway URL")
	startCmd.Flags().StringVar(&minioUrl, "minioURL", "localhost:9000", "Minio Endpoint")
	startCmd.Flags().StringVar(&accessKeyID, "accessKeyID", "access-key-id", "Minio Access Key ID")
	startCmd.Flags().StringVar(&secretAccessKey, "secretKey", "secret-access-key", "Minio Secret Access Key")

	viper.BindEnv("minioURL", "MINIO_URL")
	viper.BindPFlag("minioURL", startCmd.Flags().Lookup("minioURL"))

	viper.BindEnv("gatewayUrl", "GATEWAY_URL")
	viper.BindPFlag("gatewayUrl", startCmd.Flags().Lookup("gatewayUrl"))

	viper.BindEnv("accessKeyID", "ACCESS_KEY_ID")
	viper.BindPFlag("accessKeyID", startCmd.Flags().Lookup("accessKeyID"))

	viper.BindEnv("secretKey", "SECRET_ACCESS_KEY")
	viper.BindPFlag("secretKey", startCmd.Flags().Lookup("secretKey"))

	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the media service",
	Run: func(cmd *cobra.Command, args []string) {
		minioUrl = viper.GetString("minioUrl")
		gatewayUrl = viper.GetString("gatewayUrl")
		accessKeyID = viper.GetString("accessKeyID")
		secretAccessKey = viper.GetString("secretKey")

		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
		}

		mediaService, err := media.New(minioUrl, accessKeyID, secretAccessKey)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to create media service")
			panic(err)
		}

		authService := auth.New(gatewayUrl)
		router := server.New(mediaService, &authService)
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
