package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterMessageRoutes(r *gin.Engine){
	message := r.Group("/msg")
	{
		// 发送消息
		message.GET("/sendMsg",service.SendMsg)
		message.GET("/sendUserMsg",service.SendUserMsg)
	}
}