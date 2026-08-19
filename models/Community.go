package models

import (
	"strconv"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
)

type Community struct {
	gorm.Model
	CommunityId int64 `gorm:"not null;uniqueIndex" json:"community_id"`
	Name string
	OwnerId uint
	Img string
	Desc string
}

func CreateCommunity(community *Community)(int,string){
	if len(community.Name) == 0{
		return -1,"群名称不能为空"
	}
	
	if community.OwnerId == 0{
		return -1,"请先登录"
	}

	if err := utils.DB.Create(&community).Error;err !=nil{
		return -1,"建群失败"
	}
	return 0,"建群成功"
}

func LoadCommunity(owner_id uint)(int, []*Community,string){
	data := make([]*Community,0)
	
	utils.DB.Where("owner_id = ?",owner_id).Find(&data)

	return 0,data,"查询成功"
}

func JoinGroups(userId uint,com string)(int,string){
	contact := Contact{}
	contact.OwenId = userId
	contact.Type = 2

	community := Community{}

	utils.DB.Where("id = ? or name = ? ",com,com).Find(&community)
	if community.Name == ""{
		return -1,"群聊不存在"
	}
	if _, err := strconv.Atoi(com); err == nil {
		utils.DB.Where("owen_id = ? and target_id = ? and type = 2",userId,com,com).Find(&contact)
	} else {
		utils.DB.Where("owen_id = ? and target_id = ? and type = 2",userId,community.ID).Find(&contact)
	}
	if !contact.CreatedAt.IsZero(){
		return -1,"已存在于该群聊中"
	}else{
		contact.TargetId = uint(community.CommunityId)
		utils.DB.Create(&contact)
		return 0,"加入成功"
	}
}