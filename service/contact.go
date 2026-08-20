package service

import (
	"strconv"

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

// AddFriend
// Summary 添加好友
// @Tags 关系模块
// @Param user_id formData string false "用户id"
// @Param target_id fromData string false "目标id"
// @Success 200 {string} json{"code","message"}
// @Router /contact/addFriend
func AddFriend(ctx *gin.Context){
	user_id,_ := strconv.Atoi(ctx.PostForm("user_id"))
	target_id,_ := strconv.Atoi(ctx.PostForm("target_id"))

	if user_id == target_id{
		utils.RespFail(ctx.Writer,"添加失败")
		return
	}

	result := models.AddFriend(uint(user_id),uint(target_id))

	if result != 0{
		utils.RespOk(ctx.Writer,nil,"添加成功")
	}else{
		utils.RespFail(ctx.Writer,"添加失败")
	}
}