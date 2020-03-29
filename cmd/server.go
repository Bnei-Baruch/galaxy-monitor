package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bbfsdev/galaxy-monitor/api"
	"github.com/bbfsdev/galaxy-monitor/common"
	"github.com/bbfsdev/galaxy-monitor/utils"
	"github.com/bbfsdev/galaxy-monitor/version"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Galaxy monitor api server",
	Run:   serverFn,
}

var bindAddress string

func init() {
	serverCmd.PersistentFlags().StringVar(&bindAddress, "bind_address", "", "Bind address for server.")
	viper.BindPFlag("server.bind-address", serverCmd.PersistentFlags().Lookup("bind_address"))
	RootCmd.AddCommand(serverCmd)
}

func serverFn(cmd *cobra.Command, args []string) {
	log.Infof("Starting feed api server version %s", version.Version)
	common.Init()
	defer common.Shutdown()

	// Setup gin
	gin.SetMode(viper.GetString("server.mode"))
	router := gin.New()
	router.Use(
		utils.LoggerMiddleware(),
		utils.DataStoresMiddleware(common.DB),
		utils.ErrorHandlingMiddleware(),
		cors.Default(),
		utils.RecoveryMiddleware())

	api.SetupRoutes(router)

	log.Infoln("Running application")
	if cmd != nil {
		router.Run(viper.GetString("server.bind-address"))
	}
}
