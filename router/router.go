package router

import (
	"github.com/gin-gonic/gin"
	docs "github.com/hxysj/im-system/docs"
	"github.com/hxysj/im-system/middleware"
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
	protected := r.Group("")
	protected.Use(middleware.AuthRequired())
	RegisterPublicUserRoutes(r)
	// 注册用户模块
	RegisterUserRoutes(protected)
	// 注册消息模块
	RegisterMessageRoutes(protected)
	// 注册关系模块
	RegisterContactRoutes(protected)
	// 注册上传模块
	RegisterAttachRoutes(protected)
	// 注册群聊模块
	RegisterCommunityRoutes(protected)
	// 注册申请模块
	RegisterRelationRequest(protected)

	// 建立websocket连接
	r.GET("/msg/chat", middleware.WebSocketAuthRequired(), service.Chat)
	return r
}
