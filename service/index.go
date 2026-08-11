package service

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
)

// GetIndex
// @Tags 首页
// @Success 200 {string} welcome
// @Router /index [get]
func GetIndex(ctx *gin.Context){
	ctx.JSON(200,gin.H{
		"message":"welcome!!",
	})
}

// GetUserList
// @Tags 获取用户列表
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(ctx *gin.Context){
	data := make([]models.UserBasic,10)
	data = models.GetUserList()

	ctx.JSON(200,gin.H{
		"message":data,
	})

}