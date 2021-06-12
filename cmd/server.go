package cmd

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Bnei-Baruch/galaxy-monitor/api"
	"github.com/Bnei-Baruch/galaxy-monitor/common"
	"github.com/Bnei-Baruch/galaxy-monitor/utils"
	"github.com/Bnei-Baruch/galaxy-monitor/version"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Galaxy monitor api server",
	Run:   serverFn,
}

var bindAddress string

func init() {
	serverCmd.PersistentFlags().StringVar(&bindAddress, "bind_address", "", "Bind address for server.")
	RootCmd.AddCommand(serverCmd)
}

func serverFn(cmd *cobra.Command, args []string) {
	log.Infof("Starting monitoring api server version %s", version.Version)
	common.Init()

	api.Init()

	// cors
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, []string{"Authorization", "Content-Encoding", "Accept-Encoding"}...)
	corsConfig.AllowAllOrigins = true

	// Setup gin
	gin.SetMode(common.Config.GinMode)
	router := gin.New()
	router.Use(
		utils.LoggerMiddleware(),
		utils.ErrorHandlingMiddleware(),
		cors.New(corsConfig),
		utils.RecoveryMiddleware())

	api.SetupRoutes(router)

	log.Infof("Running application [%s]", common.Config.ListenAddress)
	if cmd != nil {
		router.Run(common.Config.ListenAddress)
	}
}
