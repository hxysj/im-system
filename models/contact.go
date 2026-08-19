package models

import (
	"fmt"
	"strconv"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
)

// 人员关系表
type Contact struct {
	gorm.Model
	ContactId int64 `gorm:"not null;uniqueIndex" json:"contact_id"`
	OwenId uint //拥有者的id
	TargetId uint //对应的用户
	Type int // 对应的关系  1好友 2群组 3
	Desc string

}

func (table *Contact) TableName() string{
	return "contact"
}

func SearchFriend(userId string) ([]UserBasic){

	contacts := make([]Contact,0)
	userIds := make([]uint64,0)

	id, err := strconv.ParseUint(userId, 10, 64)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	// 获取关系表中用户相关连的信息
	utils.DB.Where("owen_id = ? and type = ?",id,1).Find(&contacts)

	for _,v := range contacts{
		userIds = append(userIds, uint64(v.TargetId))
	}
	
	users :=  make([]UserBasic,0)
	// 获取用户详情
	utils.DB.Where("id in ?",userIds).Find(&users)

	return users
}


func AddFriend(userId uint,targetId uint) int{
	user := FindUserById(int(userId))

	if targetId != 0 && user.Salt != ""{
		contact := Contact{}
		
		utils.DB.Where("owen_id = ? and target_id = ? and type = 1",userId,targetId).Find(&contact)

		if contact.ID != 0{
			return 0
		}

		// 开启事务
		tx := utils.DB.Begin()
		// 事务开启后，不论出现什么异常最终都会Rollback
		defer func(){
			if r := recover();r != nil{
				tx.Rollback()
			}
		}()

		contact.OwenId = userId
		contact.TargetId = targetId
		contact.Type = 1
		contact.ContactId = utils.NextId()
		if err := utils.DB.Create(&contact).Error; err != nil{
			tx.Rollback()  //回滚数据库
			return 0
		}

		owner_contact := Contact{}
		owner_contact.OwenId = targetId
		owner_contact.TargetId = userId
		owner_contact.Type = 1
		owner_contact.ContactId = utils.NextId()
		utils.DB.Create(&owner_contact)
		// 两个操作都成功了之后才提交
		tx.Commit()
		return 1
	}
	return 0
}