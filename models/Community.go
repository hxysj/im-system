package models

import (
	"fmt"
	"time"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Community struct {
	gorm.Model
	CommunityId int64 `gorm:"not null;uniqueIndex" json:"community_id"`
	Name        string
	OwnerId     uint
	Img         string
	Desc        string
}

func (Community) TableName() string {
	return "community"
}

func CreateCommunity(community *Community) (int, string, int64) {
	if len(community.Name) == 0 {
		return -1, "群名称不能为空", 0
	}

	if community.OwnerId == 0 {
		return -1, "请先登录", 0
	}

	contact := Contact{}
	contact.OwenId = community.OwnerId
	contact.TargetId = uint(community.CommunityId)
	contact.ContactId = utils.NextId()
	contact.Type = 2
	contact.Desc = ""

	tx := utils.DB.Begin()

	// 事务开启后，不论出现什么异常最终都会Rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&community).Error; err != nil {
		return -1, "建群失败", 0
	}

	if err := tx.Model(&Contact{}).Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "建群失败", 0
	}

	conversation_id, _, err := CreateConversationTx(tx, int64(community.OwnerId), community.CommunityId, 1)

	if err != nil {
		tx.Rollback()
		return -1, "建群失败", 0
	}

	tx.Commit()

	return 0, "建群成功", conversation_id
}

type LoadCommunityResult struct {
	CommunityId   int64  `json:"community_id"`
	CommunityName string `json:"community_name"`
	Img           string `json:"img"`
	Desc          string `json:"desc"`
	IsOwner       bool   `json:"is_owner"`
}

func LoadCommunity(owner_id uint) (int, []LoadCommunityResult, string) {

	contact_list := make([]*Contact, 0)

	// 获取用户有关联的群聊
	if err := utils.DB.Model(&Contact{}).Where("owen_id = ? AND type = 2", owner_id).Find(&contact_list).Error; err != nil {
		fmt.Println(err)
		return -1, nil, "查询失败"
	}
	// 获取群里的id列表
	communityIds := make([]int64, 0, len(contact_list))
	for _, contact := range contact_list {
		communityIds = append(communityIds, int64(contact.TargetId))
	}
	// 如果是空的直接返回空值
	if len(communityIds) == 0 {
		return 0, nil, "查询成功"
	}
	// 获取群信息
	communityInfoList := make([]*Community, 0, len(communityIds))

	if err := utils.DB.Where("community_id IN ?", communityIds).Find(&communityInfoList).Error; err != nil {
		fmt.Println(err)
		return -1, nil, "查询失败"
	}

	data := make([]LoadCommunityResult, 0)

	for _, community := range communityInfoList {
		data = append(data, LoadCommunityResult{
			CommunityId:   community.CommunityId,
			CommunityName: community.Name,
			Img:           community.Img,
			Desc:          community.Desc,
			IsOwner:       community.OwnerId == owner_id,
		})
	}

	return 0, data, "查询成功"
}

func DeleteCommunity(user_id int64, community_id int64) error {
	var community Community
	if err := utils.DB.Model(&Community{}).
		Where("owner_id = ? AND community_id = ? ADN delete_at IS NULL").
		Take(&community).Error; err != nil {
		return err
	}

	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var conversation Conversation
		if err := tx.Model(&Conversation{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("community_id = ? AND type = ? AND status != 3", community_id, 1).
			Take(&conversation).Error; err != nil {
			return err
		}

		nowTime := time.Now()
		updateMemberRes := tx.Model(&ConversationMember{}).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("conversation_id = ?", conversation.ConversationId).Update("left_at", nowTime)

		if updateMemberRes.Error != nil {
			return updateMemberRes.Error
		}
		if updateMemberRes.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		
	})

	return nil
}
