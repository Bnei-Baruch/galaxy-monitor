package api

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"github.com/Bnei-Baruch/galaxy-monitor/common"
)

func SetupRoutes(router *gin.Engine) {
	router.POST("/update", UpdateHandler)
	router.GET("/users", UsersHandler)
	router.POST("/user_data", UserDataHandler)
	router.POST("/user_metrics", UserMetricsHandler)
	router.GET("/metrics", MetricsHandler)
	router.POST("/spec", SpecHandlerPost)
	router.GET("/spec", SpecHandlerGet)
	router.GET("/users_data", UsersDataHandler)
	router.GET("/health_check", HealthCheckHandler)

	if common.Config.HttpPProfPassword != "" {
		pRouter := router.Group("debug/pprof", gin.BasicAuth(gin.Accounts{"debug": common.Config.HttpPProfPassword}))
		pRouter.GET("/", pprofHandler(pprof.Index))
		pRouter.GET("/cmdline", pprofHandler(pprof.Cmdline))
		pRouter.GET("/profile", pprofHandler(pprof.Profile))
		pRouter.POST("/symbol", pprofHandler(pprof.Symbol))
		pRouter.GET("/symbol", pprofHandler(pprof.Symbol))
		pRouter.GET("/trace", pprofHandler(pprof.Trace))
		pRouter.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
		pRouter.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
		pRouter.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
		pRouter.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
		pRouter.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
	}
}

func pprofHandler(h http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
