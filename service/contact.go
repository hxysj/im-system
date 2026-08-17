package service

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

// SearchFiends
// Summary 获取好友列表
// @Tags 关系模块
// @param user_id formData string false "用户id"
// @Success 200 {string} json{"code","message"}
// @Router /contact/getFriendsById [post]
func SearchFriends(ctx *gin.Context){
	users := models.SearchFriend(ctx.PostForm("user_id"))

	// ctx.JSON(200,gin.H{
	// 	"code":0,
	// 	"message":"查询成功！",
	// 	"data":users,
	// })
	utils.RespOkList(ctx.Writer,users,len(users))
}