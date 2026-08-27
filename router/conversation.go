package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/service"
)

func RegisterConversationRouter(r *gin.RouterGroup) {
	conversation := r.Group("/conversation")
	{
		conversation.POST("/getPrivateConversationId", service.CreatePrivateConversation)
		conversation.POST("/getCommunityConversationId", service.CreateCommunityConversation)
	}
}
