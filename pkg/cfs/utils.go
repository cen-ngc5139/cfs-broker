package cfs

import "github.com/gin-gonic/gin"

func HandleError(c *gin.Context, err error) bool {
	if err != nil {
		jsonError(c, err.Error())
		return true
	}
	return false
}

func jsonError(c *gin.Context, msg interface{}) {
	c.AbortWithStatusJSON(200, gin.H{"ok": false, "msg": msg})
}

func JsonResult(c *gin.Context, data interface{}) {
	c.AbortWithStatusJSON(200, gin.H{"ok": true,
		"result": data,
	})
}
