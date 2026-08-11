package router

import (
	"github.com/gin-gonic/gin"
	docs "github.com/hxysj/im-system/docs"
	"github.com/hxysj/im-system/service"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)


func Router() *gin.Engine{
	r := gin.Default()
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any",ginSwagger.WrapHandler(swaggerfiles.Handler))
	
	r.GET("/index",service.GetIndex)
	r.GET("/user/getUserList",service.GetUserList)
	return r
}