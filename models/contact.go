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