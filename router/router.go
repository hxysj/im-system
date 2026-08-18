package router

import (
	"github.com/gin-gonic/gin"
	docs "github.com/hxysj/im-system/docs"
	"github.com/hxysj/im-system/service"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	r := gin.Default()
	// swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 静态资源可直接通过 /asset/emoji/1.gif 这类链接访问
	r.Static("/asset", "asset/")
	r.Static("/views", "views/")

	// 首页模块
	r.GET("/index", service.GetIndex)

	// 注册用户模块
	RegisterUserRoutes(r)
	// 注册消息模块
	RegisterMessageRoutes(r)
	// 注册关系模块
	RegisterContactRoutes(r)
	// 注册上传模块
	RegisterAttachRoutes(r)

	return r
}
