package models

import (
	"fmt"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
)

type RelationRequest struct {
	gorm.Model
	RequestId   int64 `gorm:"not null;uniqueIndex" json:"request_id"` //申请人id / 被邀请人的id
	Type        int   // 1好友申请  2 加群申请  3群邀请
	RequesterId int64
	TargetId    int64
	Status      int   // 1待处理 2通过  3拒绝
	InviteFrom  int64 // 邀请人的id
	Message     string
}

type UserInfo struct {
	UserId int64  `json:"user_id"`
	Name   string `json:"name"`
}

type CommunityInfo struct {
	CommunityId int64  `json:"community_id"`
	Name        string `json:"name"`
	Img         string `json:"img"`
}

type RelationRequestItem struct {
	RequestId   int64          `json:"request_id"`
	Type        int            `json:"type"`
	Status      int            `json:"status"`
	Message     string         `json:"message"`
	Requester   *UserInfo      `json:"requester,omitempty"`
	TargetUser  *UserInfo      `json:"target_user,omitempty"`
	RequestUser *UserInfo      `json:"request_user,omitempty"`
	Community   *CommunityInfo `json:"community,omitempty"`
	InviteFrom  *UserInfo      `json:"invite_from,omitempty"`
}

// 获取申请记录
func GetRelationRequestList(user_id int64) ([]RelationRequestItem, error) {
	var communities []Community

	if err := utils.DB.Where("owner_id = ?", user_id).Find(&communities).Error; err != nil {
		return nil, err
	}

	communityIds := make([]int64, 0, len(communities))
	for _, community := range communities {
		communityIds = append(communityIds, community.CommunityId)
	}

	var requests []RelationRequest
	// 搜素type=1（好友申请）并且target = 用户的 （对方发起申请） 或者requester 是自己的（自己发起申请） 或者 type = 3（群邀请）并且requester_id = 用户的
	query := utils.DB.Where("(type = ? AND target_id = ?) OR"+
		"(type = ? AND requester_id = ?) OR"+
		"(type = ? AND requester_id = ?)",
		1,
		user_id,
		1,
		user_id,
		3,
		user_id,
	)
	// 如果用户自己拥有的群
	if len(communityIds) > 0 {
		// 找出type = 2（群申请）的数据
		query = query.Or("type = ? AND target_id IN ?", 2, communityIds)
	}

	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}

	userIdSet := make(map[int64]struct{})
	communityIdSet := make(map[int64]struct{})

	for _, request := range requests {
		switch request.Type {
		case 1:
			// 自己发送申请的
			if request.RequesterId == user_id {
				userIdSet[request.TargetId] = struct{}{}
			} else {
				// 对方发送申请的
				// 记录下要申请添加好友的id
				userIdSet[request.RequesterId] = struct{}{}
			}
		case 2:
			// 记录下申请入群的用户id
			userIdSet[request.RequesterId] = struct{}{}
			// 记录下申请群的群id
			communityIdSet[request.TargetId] = struct{}{}
		case 3:
			// 记录下被要求入的群的id
			communityIdSet[request.TargetId] = struct{}{}
			// 记录下邀请人的id
			userIdSet[request.InviteFrom] = struct{}{}
		}
	}

	userIds := make([]int64, 0, len(userIdSet))
	for id := range userIdSet {
		userIds = append(userIds, id)
	}

	communityIdsForQuery := make([]int64, 0, len(communityIdSet))
	for id := range communityIdSet {
		communityIdsForQuery = append(communityIdsForQuery, id)
	}

	var users []UserInfo
	if len(userIds) > 0 {
		if err := utils.DB.Model(&UserBasic{}).Select("user_id,name").Where("user_id IN ?", userIds).Find(&users).Error; err != nil {
			return nil, err
		}
	}

	var communityList []CommunityInfo
	if len(communityIdsForQuery) > 0 {
		if err := utils.DB.Model(&Community{}).Select("community_id,name,img").Where("community_id IN ?", communityIdsForQuery).Find(&communityList).Error; err != nil {
			return nil, err
		}
	}

	userMap := make(map[int64]UserInfo, len(users))
	for _, user := range users {
		userMap[user.UserId] = user
	}

	communityMap := make(map[int64]CommunityInfo, len(communityList))
	for _, community := range communityList {
		communityMap[community.CommunityId] = community
	}

	result := make([]RelationRequestItem, 0, len(requests))

	for _, request := range requests {
		item := RelationRequestItem{
			RequestId: request.RequestId,
			Type:      request.Type,
			Status:    request.Status,
			Message:   request.Message,
		}

		switch request.Type {
		case 1:
			// 如果是自己发送请求的
			if request.RequesterId == user_id {
				if user, ok := userMap[request.TargetId]; ok {
					userCopy := user
					item.TargetUser = &userCopy
				}
			} else {
				if user, ok := userMap[request.RequesterId]; ok {
					userCopy := user
					item.RequestUser = &userCopy
				}
			}
		case 2:
			if user, ok := userMap[request.RequesterId]; ok {
				userCopy := user
				item.TargetUser = &userCopy
			}
			if community, ok := communityMap[request.TargetId]; ok {
				communityCopy := community
				item.Community = &communityCopy
			}
		case 3:
			if user, ok := userMap[request.InviteFrom]; ok {
				userCopy := user
				item.InviteFrom = &userCopy
			}
			if community, ok := communityMap[request.TargetId]; ok {
				communityCopy := community
				item.Community = &communityCopy
			}
		}
		result = append(result, item)
	}

	return result, nil
}

// 创建记录
func CreateRelationRequest(user_id int64, target_id int64, desc string, req_type int, inviteFrom int64) (int, string) {
	if req_type != 1 && req_type != 2 && req_type != 3 {
		return -1, "不支持的申请类型"
	}
	if req_type == 1 {
		var user UserBasic
		utils.DB.Model(&UserBasic{}).Where("user_id = ?", target_id).Find(&user)
		if user.UserId == 0 {
			return -1, "参数有误"
		}
	}
	if req_type == 2 || req_type == 3 {
		var community Community
		utils.DB.Model(&Community{}).Where("community_id = ?", target_id).Find(&community)

		if community.CommunityId == 0 {
			return -1, "参数有误"
		}

		if req_type == 3 {
			var contact Contact
			utils.DB.Model(&Contact{}).Where("target_id = ? AND owen_id = ?", target_id, user_id).Find(&contact)
			if contact.ContactId != 0 {
				return -1, "该好友已加入群聊"
			}
			utils.DB.Model(&Contact{}).Where("target_id = ? AND owen_id = ?", target_id, inviteFrom).First(&contact)
			if contact.ContactId == 0 {
				return -1, "邀请人不存在该改群中"
			}
		}
	}

	var relationRequestInfo RelationRequest

	utils.DB.Model(&RelationRequest{}).Where("requester_id = ? AND target_id = ?  AND type = ? AND status != ?", user_id, target_id, req_type, 3).Find(&relationRequestInfo)

	if relationRequestInfo.RequestId != 0 {
		return -1, "记录已存在"
	}

	relationRequestInfo.RequestId = utils.NextId()
	relationRequestInfo.Type = req_type
	relationRequestInfo.RequesterId = user_id
	relationRequestInfo.TargetId = target_id
	relationRequestInfo.Status = 1
	relationRequestInfo.Message = desc
	relationRequestInfo.InviteFrom = inviteFrom
	if err := utils.DB.Create(&relationRequestInfo).Error; err != nil {
		fmt.Println(err)
		return -1, "操作失败"
	}
	return 0, "操作成功"
}

// 修改申请记录
func ToggleRelationRequestStatus(r_id int64, user_id int64, status int) (int, string) {

	var relationRequestInfo RelationRequest

	utils.DB.Model(&RelationRequest{}).Where("request_id = ? AND status = 1", r_id).First(&relationRequestInfo)

	if relationRequestInfo.RequestId == 0 {
		return -1, "参数有误"
	}

	// 拒绝
	if status == 3 {
		if relationRequestInfo.Type == 1 && user_id != relationRequestInfo.TargetId {
			return -1, "参数有误"
		}

		if relationRequestInfo.Type == 2 || relationRequestInfo.Type == 3 {
			var community Community
			query := utils.DB.Model(&Community{}).Where("owner_id = ? AND community_id = ?", user_id, relationRequestInfo.TargetId)
			if relationRequestInfo.Type == 3 {
				query = utils.DB.Model(&Community{}).Where("community_id = ?", relationRequestInfo.TargetId)
			}

			query.First(&community)

			if community.CommunityId == 0 || (relationRequestInfo.Type == 3 && user_id != relationRequestInfo.RequesterId) {
				return -1, "参数有误"
			}
		}

		if err := utils.DB.Model(&RelationRequest{}).Where("request_id = ? AND status = ?", r_id, 1).Update("status", 3).Error; err != nil {
			fmt.Println(err)
			return -1, "修改失败"
		}
		return 0, "修改成功"
	}

	if relationRequestInfo.Type == 1 {
		if user_id != relationRequestInfo.TargetId {
			return -1, "参数有误"
		}

		res := AddFriend(uint(relationRequestInfo.RequesterId), uint(relationRequestInfo.TargetId))

		if res != 0 {
			return -1, "修改失败"
		}

		if err := utils.DB.Model(&RelationRequest{}).Where("request_id = ? AND status = ?", r_id, 1).Update("status", 2).Error; err != nil {
			fmt.Println(err)
			return -1, "修改失败"
		}

	} else {

		var community Community
		query := utils.DB.Model(&Community{}).Where("owner_id = ? AND community_id = ?", user_id, relationRequestInfo.TargetId)
		if relationRequestInfo.Type == 3 {
			query = utils.DB.Model(&Community{}).Where("community_id = ?", relationRequestInfo.TargetId)
		}
		query.First(&community)

		if community.CommunityId == 0 || (relationRequestInfo.Type == 3 && user_id != relationRequestInfo.RequesterId) {
			return -1, "参数有误"
		}

		tx := utils.DB.Begin()

		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		contact := Contact{}
		// 创建群记录
		contact.OwenId = uint(relationRequestInfo.RequesterId)
		contact.TargetId = uint(relationRequestInfo.TargetId)
		contact.Type = 2
		contact.ContactId = utils.NextId()

		if err := tx.Create(&contact).Error; err != nil {
			fmt.Println(err)
			tx.Rollback()
			return -1, "修改失败"
		}

		if err := tx.Model(&RelationRequest{}).Where("request_id = ? AND status = ?", r_id, 1).Update("status", 2).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			return -1, "修改失败"
		}

		tx.Commit()
	}

	return 0, "修改成功"
}
