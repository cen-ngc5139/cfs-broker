package router

import (
	"github.com/ghostbaby/cfs-broker/pkg/cfs"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func InitRouter(log logr.Logger) *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		//
		v1.POST("/cfs/broker", cfs.Exec(log))
		v1.POST("/cfs/config", cfs.Query(log))
	}

	return router
}

func Health() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, "health")
	}
}
