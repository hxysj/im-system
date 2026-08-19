package models

import (
	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
)

type UserBasic struct{
	gorm.Model
	UserId int64 `gorm:"not null;uniqueIndex" json:"user_id"`
	Name string
	Password string
	Phone string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email string `valid:"email"`
	Identity string
	Salt string //随机数，用于加密和解密密码的
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

// 获取用户列表
func GetUserList() []UserBasic{
	data := make([]UserBasic, 10)
	utils.DB.Find(&data)
	return data
}

func FindUserById(id int) UserBasic{
	user := UserBasic{}
	utils.DB.Where("id = ?",id).First(&user)
	return user
}

func FindUserByName(name string) UserBasic{
	user := UserBasic{}
	utils.DB.Where("name = ?",name).First(&user)
	return user
}

func FindUserByPhone(phone string) UserBasic{
	user := UserBasic{}
	utils.DB.Where("phone = ?",phone).First(&user)
	return user
}

func FindUserByEmail(email string) UserBasic{
	user := UserBasic{}
	utils.DB.Where("email = ?",email).First(&user)
	return user
}

// 创建用户
func CreateUser(user UserBasic) *gorm.DB{
	return utils.DB.Create(&user)
}

// 删除用户-软删除
func DeleteUser(user UserBasic) *gorm.DB{
	return utils.DB.Delete(&user)
}

// 更新用户信息
func UpdateUser(user UserBasic) *gorm.DB{
	return utils.DB.Model(&user).Updates(UserBasic{Name: user.Name,Password: user.Password,Phone: user.Phone,Email: user.Email,Identity: user.Identity})
}