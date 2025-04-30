package middles

import (
	"github.com/clakeboy/golib/components"
	"github.com/gin-gonic/gin"
)

func Cache() gin.HandlerFunc {
	return func(c *gin.Context) {
		cache := components.NewMemCache()
		c.Set("cache", cache)
		c.Next()
	}
}
