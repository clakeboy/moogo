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
	case "conn":
		return controllers.NewConnectController(c)
	case "exec":
		return controllers.NewExecController(c)
	case "index":
		return controllers.NewIndexesController(c)
	case "coll":
		return controllers.NewCollectionController(c)
	case "database":
		return controllers.NewDatabaseController(c)
	case "backup":
		return controllers.NewBackupController(c)
	default:
		return nil
	}
}
