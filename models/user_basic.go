package models

import (
	"gorm.io/gorm"
)

type UserBasic struct{
	gorm.Model
	Name string
	Password string
	Phone string
	Email string
	Identity string
	ClientIP string
	ClientPort string
	LoginTime uint64
	HeartbeatTime uint64
	LoginOutTime uint64 `gorm:"column:login_out_time"`
	IsLogout bool
	DeviceInfo string
}

func (table *UserBasic)TableName() string{
	return "user_basic"
}