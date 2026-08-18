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
		// 建立websocket连接
		message.GET("/chat",service.Chat)
		// 获取表情包列表
		message.GET("/getEmojiList",service.GetEmojiList)
	}
}