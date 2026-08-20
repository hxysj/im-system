package service

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hxysj/im-system/models"
	"github.com/hxysj/im-system/utils"
)

func GetRelationRequestList(ctx *gin.Context){
	user_id,_ := strconv.Atoi(ctx.Query("user_id"))

	user := models.FindUserById(user_id)

	if user.Salt == ""{
		utils.RespFail(ctx.Writer,"参数有误")
		return
	}

	result,err := models.GetRelationRequestList(user.UserId)

	if err != nil{
		utils.RespFail(ctx.Writer,"获取失败")
		return
	}else{
		utils.RespOkList(ctx.Writer,result,len(result))
	}
}

func ToggleRelationRequestStatus(ctx *gin.Context){
	request_id,_ := strconv.Atoi(ctx.PostForm("request_id"))
	status,_ := strconv.Atoi(ctx.PostForm("status"))
	user_id,_ := strconv.Atoi(ctx.PostForm("user_id"))

	if status != 2 && status != 3{
		utils.RespFail(ctx.Writer,"参数有误")
		return
	}

	res,msg := models.ToggleRelationRequestStatus(int64(request_id),int64(user_id),status)

	if res == -1{
		utils.RespFail(ctx.Writer,msg)
	}else{
		utils.RespOk(ctx.Writer,nil,"修改成功")
	}
}

// 好友申请
func CreateFriendRelation(ctx *gin.Context){
	user_id,_ := strconv.Atoi(ctx.PostForm("user_id"))
	target_id,_ := strconv.Atoi(ctx.PostForm("target_id"))
	desc := ctx.PostForm("context")

	code,msg := models.CreateRelationRequest(int64(user_id),int64(target_id),desc,1)

	if code == -1{
		utils.RespFail(ctx.Writer,msg)
	}else{
		utils.RespOk(ctx.Writer,nil,msg)
	}
}

// 群申请
func CreateCommunityRelation(ctx *gin.Context){
	user_id,_ := strconv.Atoi(ctx.PostForm("user_id"))
	target_id,_ := strconv.Atoi(ctx.PostForm("target_id"))
	desc := ctx.PostForm("context")

	code,msg := models.CreateRelationRequest(int64(user_id),int64(target_id),desc,2)

	if code == -1{
		utils.RespFail(ctx.Writer,msg)
	}else{
		utils.RespOk(ctx.Writer,nil,msg)
	}
}