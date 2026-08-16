package models

import "gorm.io/gorm"

// 人员关系表
type Contact struct {
	gorm.Model
	OwenId uint //拥有者的id
	TargetId uint //对应的用户
	Type int // 对应的关系 0 1 3
	Desc string

}

func (table *Contact) TableName() string{
	return "contact"
}
