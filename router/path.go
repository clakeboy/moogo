package router

import (
	"github.com/gin-gonic/gin"
	"moogo/controllers"
)

func GetController(controllerName string, c *gin.Context) interface{} {
	switch controllerName {
	case "account":
		return controllers.NewAccountController(c)
	case "login":
		return controllers.NewLoginController(c)
	case "server":
		return controllers.NewServerController(c)
	default:
		return nil
	}
}
