package models

import (
	"time"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserBasic struct {
	gorm.Model
	UserId        int64 `gorm:"not null;uniqueIndex" json:"user_id"`
	Name          string
	Password      string
	Phone         string `valid:"matches(^1[3-9]{1}\\d{9}$)"`
	Email         string `valid:"email"`
	Identity      string
	Salt          string //随机数，用于加密和解密密码的
	ClientIP      string
	ClientPort    string
	LoginTime     uint64
	HeartbeatTime uint64
	LoginOutTime  uint64 `gorm:"column:login_out_time"`
	IsLogout      bool
	DeviceInfo    string
	Avatar        string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

// 获取用户列表
func GetUserList() []UserBasic {
	data := make([]UserBasic, 10)
	utils.DB.Find(&data)
	return data
}

func FindUserById(id int64) UserBasic {
	user := UserBasic{}
	utils.DB.Where("user_id = ?", id).Take(&user)
	return user
}

func FindUserByName(name string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ?", name).First(&user)
	return user
}

func FindUserByPhone(phone string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("phone = ?", phone).First(&user)
	return user
}

func FindUserByEmail(email string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("email = ?", email).First(&user)
	return user
}

// 创建用户
func CreateUser(user UserBasic) *gorm.DB {
	return utils.DB.Create(&user)
}

// 删除用户-软删除
func DeleteUser(userId int64) error {

	return utils.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var user UserBasic
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userId).Take(&user).Error; err != nil {
			return err
		}
		// 设置用户离开了会话
		if err := tx.Model(&ConversationMember{}).Where("user_id = ? AND left_at IS NULL", userId).
			Updates(map[string]interface{}{
				"left_at":      now,
				"visible_at":   nil,
				"unread_count": 0,
			}).Error; err != nil {
			return err
		}
		// 删除对应的关系表
		if err := tx.Where("owen_id = ? OR (type = 1 AND target_id = ?)", userId, userId).
			Delete(&Contact{}).Error; err != nil {
			return err
		}
		// 为通过的申请记录设置成失效
		if err := tx.Model(&RelationRequest{}).Where("status = 1 AND (requester_id = ? OR target_id = ? OR invite_from = ?)", userId, userId, userId).
			Update("status", 4).Error; err != nil {
			return err
		}
		return tx.Delete(&user).Error
	})
}

// 更新用户信息
func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&UserBasic{}).
		Where("user_id = ?", user.UserId).
		Updates(map[string]interface{}{
			"name":     user.Name,
			"phone":    user.Phone,
			"email":    user.Email,
			"identity": user.Identity,
		})
}

// 通过手机号或者名称搜索用户
func FindUserByPhoneOrNameOrEmail(key string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ? or phone = ? or email = ?", key, key, key).First(&user)
	return user
}

// 修改用户密码
func UpdateUserPassword(user_id int64, oldPassword string, newPassword string) (int, string) {

	user := FindUserById(user_id)

	oldPwd := utils.MakePassword(oldPassword, user.Salt)

	if oldPwd != user.Password {
		return -1, "密码错误"
	}

	newPwd := utils.MakePassword(newPassword, user.Salt)

	if newPwd == user.Password {
		return -1, "新密码不能与旧密码一样"
	}

	if err := utils.DB.Model(&user).Update("password", newPwd).Error; err != nil {
		return -1, "修改失败"
	}

	return 0, "修改成功"
}
