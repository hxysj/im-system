package models

import (
	"errors"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
)

// 人员关系表
type Contact struct {
	gorm.Model
	ContactId int64 `gorm:"not null;uniqueIndex" json:"contact_id"`
	OwenId    uint  //拥有者的id
	TargetId  uint  //对应的用户
	Type      int   // 对应的关系  1好友 2群组 3
	Desc      string
}

func (table *Contact) TableName() string {
	return "contact"
}

type SearchFriendResult struct {
	UserId int64  `json:"user_id"`
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
}

func SearchFriend(userId int64) []SearchFriendResult {

	contacts := make([]Contact, 0)
	userIds := make([]uint64, 0)

	// 获取关系表中用户相关连的信息
	utils.DB.Where("owen_id = ? and type = ?", userId, 1).Find(&contacts)

	for _, v := range contacts {
		userIds = append(userIds, uint64(v.TargetId))
	}

	users := make([]SearchFriendResult, 0)
	// 获取用户详情
	utils.DB.Model(&UserBasic{}).Select("user_id,name,phone,email").Where("user_id in ?", userIds).Find(&users)

	return users
}

func AddFriend(userId uint, targetId uint) int {
	err := utils.DB.Transaction(func(tx *gorm.DB) error {
		return AddFriendTx(*tx, userId, targetId)
	})
	if err != nil {
		return -1
	}
	return 0
}

func AddFriendTx(tx gorm.DB, userId uint, targetId uint) error {
	user := FindUserById(int64(userId))

	if targetId != 0 && user.UserId != 0 {
		contact := Contact{}

		utils.DB.Where("owen_id = ? and target_id = ? and type = 1", userId, targetId).Find(&contact)

		if contact.ContactId != 0 {
			return errors.New("invalid friend relation")
		}

		contact.OwenId = userId
		contact.TargetId = targetId
		contact.Type = 1
		contact.ContactId = utils.NextId()
		if err := tx.Create(&contact).Error; err != nil {
			return err
		}

		owner_contact := Contact{}
		owner_contact.OwenId = targetId
		owner_contact.TargetId = userId
		owner_contact.Type = 1
		owner_contact.ContactId = utils.NextId()
		if err := tx.Create(&owner_contact).Error; err != nil {
			tx.Rollback()
			return err
		}
		// 两个操作都成功了之后才提交
		tx.Commit()
		return nil
	}
	return errors.New("invalid friend relation")
}
