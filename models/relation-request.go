package models

import (
	"fmt"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
)

type RelationRequest struct {
	gorm.Model
	RequestId int64 `gorm:"not null;uniqueIndex" json:"request_id"`
	Type int  // 1好友申请  2 加群申请  3群邀请
	RequesterId int64
	TargetId int64
	Status int // 1待处理 2通过  3拒绝
	Message string
}

// 获取申请记录
func GetRelationRequestList (user_id int64) ([]RelationRequest,error){
	var communities []Community

	if err := utils.DB.Where("owner_id = ?",user_id).Find(&communities).Error;err != nil{
		return nil,err
	}

	communityIds := make([]int64,0,len(communities))
	for _,community := range communities{
		communityIds = append(communityIds, community.CommunityId)
	}

	var requests []RelationRequest

	query := utils.DB.Where("target_id = ?",user_id)
	if len(communityIds) > 0{
		query = utils.DB.Where("(target_id = ? AND type = 1) OR (type = 2 AND target_id IN ?)",user_id,communityIds)
	}

	if err:= query.Find(&requests).Error;err != nil{
		return nil,err
	}

	return requests,nil
}

// 创建记录
func CreateRelationRequest(user_id int64,target_id int64,desc string,req_type int)(int,string){
	if req_type == 1{
		var user UserBasic
		utils.DB.Model(&UserBasic{}).Where("user_id = ?",target_id).Find(&user)
		if user.UserId == 0{
			return -1,"参数有误"
		}
	}
	if req_type == 2{
		var community Community
		utils.DB.Model(&Community{}).Where("community_id = ?",target_id).Find(&community)

		if community.CommunityId == 0{
			return -1,"参数有误"
		}
	}
	var relationRequestInfo RelationRequest

	utils.DB.Model(&RelationRequest{}).Where("requester_id = ? AND target_id = ?",user_id,target_id).Find(&relationRequestInfo)

	if relationRequestInfo.RequestId != 0{
		return -1,"记录已存在"
	}

	relationRequestInfo.RequestId = utils.NextId()
	relationRequestInfo.Type = req_type
	relationRequestInfo.RequesterId = user_id
	relationRequestInfo.TargetId = target_id
	relationRequestInfo.Status = 1
	relationRequestInfo.Message = desc
	utils.DB.Create(&relationRequestInfo)
	return 0,"申请成功"
}

// 修改申请记录
func ToggleRelationRequestStatus(r_id int64,user_id int64,status int) (int,string) {

	var relationRequestInfo RelationRequest

	utils.DB.Model(&RelationRequest{}).Where("request_id = ?",r_id).Find(&relationRequestInfo)

	if relationRequestInfo.RequestId == 0{
		return -1,"参数有误"
	}

	if status == 3{
		if err := utils.DB.Model(&RelationRequest{}).Where("request_id = ? AND status = ?",r_id,1).Update("status",3).Error;err != nil{
			fmt.Println(err)
			return -1,"修改失败"
		}
		return 0,"修改成功"
	}

	if relationRequestInfo.Type == 1{
		if user_id != relationRequestInfo.TargetId{
			return -1,"参数有误"
		}

		res := AddFriend(uint(relationRequestInfo.RequesterId),uint(relationRequestInfo.TargetId))

		if res != 0{
			return -1,"修改失败"
		}

	}else{
		var community Community
		utils.DB.Model(&Community{}).Where("owner_id = ? AND community_id = ?",user_id,relationRequestInfo.TargetId).First(&community)

		if community.CommunityId == 0{
			return -1,"参数有误"
		}

		contact := Contact{}
		// 创建群记录
		contact.OwenId = uint(relationRequestInfo.RequesterId)
		contact.TargetId =uint(relationRequestInfo.TargetId)
		contact.Type = 2
		contact.ContactId = utils.NextId()

		if err := utils.DB.Create(&contact).Error;err != nil{
			fmt.Println(err)
			return -1,"修改失败"
		}
	}

	if err := utils.DB.Model(&RelationRequest{}).Where("request_id = ? AND status = ?",r_id,1).Update("status",2).Error;err != nil{
		fmt.Println(err)
		return -1,"修改失败"
	}
	return 0,"修改成功"
}