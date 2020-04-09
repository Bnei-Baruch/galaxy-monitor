package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.POST("/update", UpdateHandler)
	router.GET("/users", UsersHandler)
	router.POST("/user_data", UserDataHandler)
}
