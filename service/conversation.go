package service

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

// 创建/获取私聊会话
func CreatePrivateConversation(ctx *gin.Context) {
	user_id := ctx.GetInt64("current_user_id")
	target_id, err := strconv.Atoi(ctx.PostForm("target_id"))

	if err != nil {
		utils.RespFail(ctx.Writer, "创建会话失败")
		return
	}

	conversationId, msg, err := models.CreateConversation(user_id, int64(target_id), 2)

	if err != nil {
		fmt.Println(err)
		utils.RespFail(ctx.Writer, msg)
	} else {
		type Result struct {
			ConversationId int64 `json:"conversation_id"`
		}
		utils.RespOk(ctx.Writer, Result{
			ConversationId: conversationId,
		}, msg)
	}
}

// 创建/获取群聊会话
func CreateCommunityConversation(ctx *gin.Context) {
	user_id := ctx.GetInt64("current_user_id")
	community_id, err := strconv.Atoi(ctx.PostForm("community_id"))

	if err != nil || community_id == 0 {
		utils.RespFail(ctx.Writer, "参数有误")
		return
	}

	conversation_id, msg, err := models.CreateConversation(user_id, int64(community_id), 1)

	if err != nil {
		utils.RespFail(ctx.Writer, msg)
	} else {
		type Result struct {
			ConversationId int64 `json:"conversation_id"`
		}

		utils.RespOk(ctx.Writer, Result{
			ConversationId: conversation_id,
		}, "")
	}
}
