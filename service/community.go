package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

func CreateCommunity(c *gin.Context){
	ownerId,_ := strconv.Atoi(c.PostForm("owner_id"))
	name := c.PostForm("name")
	desc := c.PostForm("desc")
	community := models.Community{}
	community.OwnerId = uint(ownerId)
	community.Name = name
	community.Desc = desc
	community.CommunityId = utils.NextId()
	status,msg := models.CreateCommunity(&community)

	if status == 0{
		utils.RespOk(c.Writer,nil,msg)
	}else{
		utils.RespFail(c.Writer,msg)
	}
}


func LoadCommunity(c *gin.Context){
	user_id,_ := strconv.Atoi(c.PostForm("user_id"))
	status,data,msg := models.LoadCommunity(uint(user_id))

	if status == -1{
		utils.RespFail(c.Writer,msg)
	}else{
		utils.RespOkList(c.Writer,data,len(data))
	}
}

func JoinGroups(c *gin.Context){
	userId,_ := strconv.Atoi(c.PostForm("user_id"))

	status,msg := models.JoinGroups(uint(userId),c.PostForm("community"))

	if status == -1{
		utils.RespFail(c.Writer,msg)
	}else{
		utils.RespOk(c.Writer,nil,msg)
	}
}