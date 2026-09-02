package service

import (
	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

func CreateCommunity(c *gin.Context) {
	ownerId := c.GetInt64("current_user_id")
	name := c.PostForm("name")
	desc := c.PostForm("desc")
	community := models.Community{}
	community.OwnerId = uint(ownerId)
	community.Name = name
	community.Desc = desc
	community.CommunityId = utils.NextId()
	status, msg, conversation_id := models.CreateCommunity(&community)

	if status == 0 {
		data := map[string]int64{
			"conversation_id": conversation_id,
		}
		utils.RespOk(c.Writer, data, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func LoadCommunity(c *gin.Context) {
	user_id := c.GetInt64("current_user_id")

	status, data, msg := models.LoadCommunity(uint(user_id))

	if status == -1 {
		utils.RespFail(c.Writer, msg)
	} else {
		utils.RespOkList(c.Writer, data, len(data))
	}
}
