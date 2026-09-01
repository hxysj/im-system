package models

import (
	"fmt"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
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

func CreateCommunity(community *Community) (int, string) {
	if len(community.Name) == 0 {
		return -1, "群名称不能为空"
	}

	if community.OwnerId == 0 {
		return -1, "请先登录"
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
		return -1, "建群失败"
	}

	if err := tx.Model(&Contact{}).Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "建群失败"
	}

	tx.Commit()

	return 0, "建群成功"
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
