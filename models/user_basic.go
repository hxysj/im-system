package models

import (
	"errors"
	"os"
	"time"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserBasic struct {
	gorm.Model
	UserId        int64  `gorm:"not null;uniqueIndex" json:"user_id"`
	Name          string `gorm:"not null;uniqueIndex" json:"name"`
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

		// 将对应的会话 status设置成 2 表示无法发送消息
		var conversationIds []int64
		if err := tx.Table("conversation_member AS cm").
			Joins("JOIN conversation AS c ON c.conversation_id = cm.conversation_id").
			Select("cm.conversation_id").
			Where("cm.user_id = ? AND c.type = 2", userId).
			Pluck("cm.conversation_id", &conversationIds).
			Error; err != nil {
			return err
		}

		var ownedCommunityIDs []int64
		if err := tx.Model(&Community{}).
			Where("owner_id = ?", userId).
			Pluck("community_id", &ownedCommunityIDs).Error; err != nil {
			return err
		}

		if len(ownedCommunityIDs) > 0 {
			result := tx.Model(&Conversation{}).
				Where("type = ? AND status IN ? AND community_id IN ?",
					1, []int{
						utils.ConversationStatusNormal,
						utils.ConversationStatusMuted,
					},
					ownedCommunityIDs,
				).Update("status", utils.ConversationStatusDissolved)
			if result.Error != nil {
				return result.Error
			}
			// 将加群申请和群邀请申请都设置成已失效
			if err := tx.Model(&RelationRequest{}).
				Where("type IN ? AND target_id IN ? AND status = ?",
					[]int{2, 3}, ownedCommunityIDs, 1,
				).Update("status", 4).Error; err != nil {
				return err
			}
			// 解散对应的群聊
			if err := tx.Where("community_id IN ? AND owner_id = ?", ownedCommunityIDs, userId).
				Delete(&Community{}).Error; err != nil {
				return err
			}
		}

		if len(conversationIds) > 0 {
			result := tx.Model(&Conversation{}).
				Where("type = ? AND status = ? AND conversation_id IN ?",
					2, utils.ConversationStatusNormal, conversationIds).
				Update("status", utils.ConversationStatusDissolved)

			if result.Error != nil {
				return result.Error
			}
		}

		// 删除对应的关系表
		if err := tx.Where("owen_id = ? OR (type = 1 AND target_id = ?)", userId, userId).
			Delete(&Contact{}).Error; err != nil {
			return err
		}

		// 未通过的申请记录设置成失效
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

func UpdateUserAvatar(user_id int64, file_path string) (error, string) {
	_, err := os.Stat(file_path)

	if os.IsNotExist(err) {
		return err, "图片上传失败！"
	}

	result := utils.DB.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Model(&UserBasic{}).Where("user_id = ? ", user_id).
		Update("avatar = ?", file_path)
	if result.Error != nil {
		return result.Error, "更新用户头像失败"
	}
	if result.RowsAffected != 1 {
		return errors.New("更新失败，未搜素到对应用户！"), "更新用户头像失败"
	}

	return nil, "修改成功！"
}

type UserInfoResult struct {
	UserId       int64  `json:"user_id"`
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	LoginTime    uint64 `json:"login_time"`
	LoginOutTime uint64 `json:"login_out_time"`
	Avatar       string `json:"avatar"`
}

func GetUserInfo(user_id int64) (error, UserInfoResult) {
	var result UserInfoResult
	if err := utils.DB.Model(&UserBasic{}).
		Where("user_id = ?", user_id).
		Select("user_id,name,phone,email,login_time,login_out_time,avatar").
		Take(&result).Error; err != nil {
		return err, UserInfoResult{}
	}
}
