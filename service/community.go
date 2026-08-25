package service

import (
	"fmt"
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
	user_id,err := strconv.Atoi(c.PostForm("user_id"))

	if err != nil || user_id <= 0{
		fmt.Println(err)
		utils.RespFail(c.Writer,"参数有误")
		return
	}

	status,data,msg := models.LoadCommunity(uint(user_id))

	if status == -1{
		utils.RespFail(c.Writer,msg)
	}else{
		utils.RespOkList(c.Writer,data,len(data))
	}
}