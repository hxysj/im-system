package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)


func Router() *gin.Engine{
	r := gin.Default()
	
	r.GET("/index",service.GetIndex)
	return r
}