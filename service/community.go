package service

import (
	"encoding/json"
	"fmt"
	"strconv"

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

func DeleteCommunity(c *gin.Context) {
	user_id := c.GetInt64("current_user_id")
	community_id, err := strconv.Atoi(c.PostForm("community_id"))

	if err != nil || community_id <= 0 || user_id <= 0 {
		utils.RespFail(c.Writer, "参数有误")
		return
	}

	delErr := models.DeleteCommunity(user_id, int64(community_id))
	if delErr != nil {
		fmt.Println("删除群聊失败:", delErr)
		utils.RespFail(c.Writer, "删除失败！")
	} else {
		utils.RespOk(c.Writer, nil, "删除成功！")
	}
}

func ChangeCommunityInfo(c *gin.Context) {
	user_id := c.GetInt64("current_user_id")
	var changeData models.CommunityInfoItem
	err := json.Unmarshal([]byte(c.PostForm("info")), &changeData)
	if err != nil {
		utils.RespFail(c.Writer, "参数有误！")
		return
	}
	if user_id <= 0 || changeData.CommunityId <= 0 {
		utils.RespFail(c.Writer, "参数有误！")
		return
	}

	err = models.ChangeCommunityInfo(user_id, changeData)
	if err != nil {
		utils.RespFail(c.Writer, "修改失败！")
	} else {
		utils.RespOk(c.Writer, nil, "修改成功")
	}
}
