package common

import (
	"os"

	"github.com/gin-gonic/gin"
)

type config struct {
	ListenAddress     string
	GinMode           string
	HttpPProfPassword string
}

func newConfig() *config {
	return &config{
		ListenAddress:     ":8081",
		GinMode:           gin.DebugMode,
		HttpPProfPassword: "",
	}
}

var Config *config

func Init() {
	Config = newConfig()

	if val := os.Getenv("LISTEN_ADDRESS"); val != "" {
		Config.ListenAddress = val
	}
	if val := os.Getenv(gin.EnvGinMode); val != "" {
		Config.GinMode = val
	}
	if val := os.Getenv("HTTP_PPROF_PASSWORD"); val != "" {
		Config.HttpPProfPassword = val
	}
}
