package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterMessageRoutes(r *gin.RouterGroup) {
	message := r.Group("/msg")
	{
		// 发送消息
		message.GET("/sendMsg", service.SendMsg)
		// 获取表情包列表
		message.GET("/getEmojiList", service.GetEmojiList)
		// 获取消息列表
		message.GET("/getMessageList", service.GetMessageList)
		// 将会话的消息设置成已读
		message.POST("/readMessage", service.ReadMessage)
	}
}
