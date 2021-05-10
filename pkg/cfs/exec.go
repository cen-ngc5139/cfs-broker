package cfs

import (
	"fmt"

	traitv1 "github.com/ghostbaby/cfs-broker/pkg/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
)

func Exec(log logr.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		spec := &traitv1.CfsTrait{}
		if err := c.Bind(spec); HandleError(c, err) {
			log.Error(err, "fail to bind post param")
			return
		}

		fmt.Println(spec)

		if err := Validate(spec); HandleError(c, err) {
			log.Error(err, "fail to check post param")
			return
		}

		if err := Controller(spec); HandleError(c, err) {
			log.Error(err, "fail to update cfs config")
			return
		}

		c.AbortWithStatusJSON(200, gin.H{"ok": true})
	}
}
