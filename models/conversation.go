package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 唯一索引能够指定字段或者字段组合在数据库中不能重复
// 普通索引使用  index:索引名称
// gorm 给GORM的数据库映射配置  uniqueIndex:索引名称 创建唯一索引  un_conversation_private 唯一索引名称   priority 该字段在联合索引中的优先级
// json  JSON序列化配置  private_low_user_id 是返回给前端的JSON字段名  omitempty 值为空的时候不输出该字段

type Conversation struct {
	ConversationId    int64      `gorm:"primaryKey" json:"conversation_id"`                                                                                          // 会话id
	Type              int        `gorm:"not null;uniqueIndex:uk_conversation_private,priority:1;uniqueIndex:uk_conversation_group,priority:1;default:1" json:"type"` //2 私聊  1群聊
	PrivateLowUserId  *int64     `gorm:"uniqueIndex:uk_conversation_private,priority:2" json:"private_low_user_id,omitempty"`                                        //私聊中双方id较小的用户
	PrivateHighUserId *int64     `gorm:"uniqueIndex:uk_conversation_private,priority:3" json:"private_high_user_id,omitempty"`                                       //私聊中双方id较大的用户
	CommunityId       *int64     `gorm:"uniqueIndex:uk_conversation_group,priority:2" json:"community_id,omitempty"`                                                 //群会话所关联的群id
	LastMessageId     int64      `gorm:"not null;default:0" json:"last_message_id"`                                                                                  //会话的最后一条消息id
	LastMessageAt     *time.Time `gorm:"index:idx_conversation_last_message" json:"last_message_at"`                                                                 //会话最后一条消息的时间
	Status            int        `gorm:"not null;default:1" json:"status"`                                                                                           //会话状态，例如正常、解散、禁言
	CreatedAt         time.Time  `gorm:"not null;autoCreateTime" json:"created_at"`                                                                                  //会话的创建时间
	UpdatedAt         time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`                                                                                  //会话的更新时间
}

func (Conversation) TableName() string {
	return "conversation"
}

type ConversationMember struct {
	ConversationId       int64      `gorm:"primaryKey" json:"conversation_id"`                 //用户所属的会话
	UserId               int64      `gorm:"primaryKey" json:"user_id"`                         //会话的成员id
	Role                 int        `gorm:"not null;default:0" json:"role"`                    //成员的身份 1 普通用户  2管理员  3 创建者
	LastReadMessageId    int64      `gorm:"not null;default:0" json:"last_read_message_id"`    //用户最后读到的消息id
	UnreadCount          int        `gorm:"not null;default:0" json:"unread_count"`            //当前未读消息数量的缓存
	IsPinned             bool       `gorm:"not null;default:false" json:"is_pinned"`           //是否置顶会话
	IsMuted              bool       `gorm:"not null;default:false" json:"is_muted"`            //是否开启消息免打扰
	JoinedAt             time.Time  `gorm:"not null;autoCreateTime" json:"joined_at"`          //用户加入会话的时间
	LeftAt               *time.Time `json:"left_at,omitempty"`                                 //用户退出会话的时间
	UpdatedAt            time.Time  `gorm:"not null;autoUpdateTime" json:"updated_at"`         //状态最后的修改时间
	VisibleAt            *time.Time `json:"visible_at,omitempty"`                              //用户是否可见会话列表
	ClearBeforeMessageId int64      `gorm:"not null;default:0" json:"clear_before_message_id"` //用户只允许看到此消息之后的消息
}

func (ConversationMember) TableName() string {
	return "conversation_member"
}

// 检查是否存在好友关系
func IsMutualFriend(tx *gorm.DB, userId int64, targetUserId int64) (bool, error) {
	var contacts []Contact
	err := tx.Where("type = ? AND ((owen_id = ? AND target_id = ?) OR (owen_id = ? AND target_id = ?))", 1, userId, targetUserId, targetUserId, userId).Find(&contacts).Error

	if err != nil {
		return false, err
	}

	var forwardExists bool
	var reverseExists bool

	for _, contact := range contacts {
		switch {
		case int64(contact.OwenId) == userId && int64(contact.TargetId) == targetUserId:
			forwardExists = true
		case int64(contact.TargetId) == userId && int64(contact.OwenId) == targetUserId:
			reverseExists = true
		}
	}

	return forwardExists && reverseExists, nil
}

func CreateConversation(userId int64, target int64, cType int) (int64, string, error) {
	var conversationId int64
	var msg string

	err := utils.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		conversationId, msg, err = CreateConversationTx(tx, userId, target, cType)
		return err
	})

	return conversationId, msg, err
}

func CreateConversationTx(tx *gorm.DB, userId int64, target int64, cType int) (int64, string, error) {
	// 如果cType为1 的话是群聊  2是私聊
	if cType == 2 {
		isFriend, err := IsMutualFriend(tx, userId, target)

		if err != nil {
			return 0, "好友关系查询失败", err
		}

		if !isFriend {
			return 0, "好友关系异常", errors.New("对方还不是你的好友")
		}

		var conversation Conversation
		lowUserId := userId
		HighUserId := target

		if lowUserId > HighUserId {
			lowUserId, HighUserId = HighUserId, lowUserId
		}

		result := tx.Model(&Conversation{}).Where("private_low_user_id = ? AND private_high_user_id = ? AND type = 2", lowUserId, HighUserId).First(&conversation)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, "查询会话失败", result.Error
		}
		// 没找到报的错误
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// 会话存在的话需要检查用户是否可以看见，不可以的话要更新visibleAt
			var conversationMember ConversationMember

			result := tx.Model(&ConversationMember{}).Where("conversation_id = ? AND user_id = ?", conversation.ConversationId, userId).First(&conversationMember)

			// 需要判断用户是否关联对应的会话
			if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return 0, "查询会话失败", result.Error
			}
			// 没找到的需要新建关联
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				conversationMember = ConversationMember{
					ConversationId: conversation.ConversationId,
					UserId:         userId,
				}
				now := time.Now()
				conversationMember.VisibleAt = &now
				if err := tx.Model(&ConversationMember{}).Create(conversationMember).Error; err != nil {
					return 0, "查询会话失败", err
				}
				return conversation.ConversationId, "", nil
			}

			now := time.Now()
			if err := tx.Model(&ConversationMember{}).
				Where("conversation_id = ? AND user_id = ?", conversation.ConversationId, userId).
				Updates(map[string]interface{}{
					"visible_at": now,
					"left_at":    nil,
				}).Error; err != nil {
				return 0, "新建会话失败", err
			}

			return conversation.ConversationId, "", nil
		} else {
			conversation = Conversation{
				ConversationId:    utils.NextId(),
				Type:              2,
				PrivateLowUserId:  &lowUserId,
				PrivateHighUserId: &HighUserId,
			}
			if err := tx.Model(&Conversation{}).Create(conversation).Error; err != nil {
				return 0, "新建会话失败", err
			}
			now := time.Now()
			conversationMember := ConversationMember{
				ConversationId: conversation.ConversationId,
				UserId:         userId,
				VisibleAt:      &now,
				LeftAt:         nil,
			}

			if err := tx.Model(&ConversationMember{}).Create(conversationMember).Error; err != nil {
				return 0, "新建会话失败", err
			}

			targetConversationMember := ConversationMember{
				ConversationId: conversation.ConversationId,
				UserId:         target,
			}

			if err := tx.Model(&ConversationMember{}).Create(targetConversationMember).Error; err != nil {
				return 0, "新建会话失败", err
			}

			return conversation.ConversationId, "", nil
		}

	} else if cType == 1 {
		// 判断群聊是否存在,判断用户是否在群聊中
		var community Community
		if err := tx.Model(&Community{}).Where("community_id = ?", target).First(&community).Error; err != nil {
			return 0, "查询会话失败", err
		}

		var contact Contact
		if err := tx.Model(&Contact{}).Where("owen_id = ? AND target_id = ? AND type = 2", userId, target).First(&contact).Error; err != nil {
			return 0, "查询会话失败", err
		}

		var conversation Conversation

		result := tx.Model(&Conversation{}).Where("type = 1 AND community_id = ?", community.CommunityId).First(&conversation)

		if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, "查询会话失败", result.Error
		}

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			conversation = Conversation{
				ConversationId: utils.NextId(),
				Type:           1,
				CommunityId:    &community.CommunityId,
			}

			if err := tx.Model(&Conversation{}).Create(conversation).Error; err != nil {
				return 0, "获取会话信息失败", err
			}

			conversationMember := ConversationMember{
				ConversationId: conversation.ConversationId,
				UserId:         userId,
			}

			if community.OwnerId == uint(userId) {
				conversationMember.Role = 3
			}

			nowTime := time.Now()
			conversationMember.VisibleAt = &nowTime

			if err := tx.Model(&ConversationMember{}).Create(conversationMember).Error; err != nil {
				return 0, "获取会话信息失败", err
			}
			return conversation.ConversationId, "", nil

		} else {

			var conversationMember ConversationMember

			result := tx.Model(&ConversationMember{}).Where("conversation_id = ? AND user_id = ?", conversation.ConversationId, userId).First(&conversationMember)

			if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return 0, "获取会话信息失败", result.Error
			}

			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				conversationMember = ConversationMember{
					ConversationId: conversation.ConversationId,
					UserId:         userId,
				}
				nowTime := time.Now()
				conversationMember.VisibleAt = &nowTime
				if community.OwnerId == uint(userId) {
					conversationMember.Role = 3
				}
				if err := tx.Model(&ConversationMember{}).Create(conversationMember).Error; err != nil {
					return 0, "获取会话信息失败", err
				}
			} else {
				nowTime := time.Now()
				if err := tx.Model(&ConversationMember{}).Where("conversation_id = ? AND user_id = ?", conversation.ConversationId, userId).Updates(map[string]interface{}{
					"visible_at": nowTime,
					"left_at":    nil,
				}).Error; err != nil {
					return 0, "获取会话信息失败", err
				}
			}

			return conversation.ConversationId, "", nil
		}
	}

	return 0, "", fmt.Errorf("参数有误")
}

type LastMessageInfo struct {
	MessageId int64     `json:"message_id"`
	Content   string    `json:"content"`
	Media     int       `json:"media"`
	Pic       string    `json:"pic"`
	Url       string    `json:"url"`
	Desc      string    `json:"desc"`
	CreateAt  time.Time `json:"created_at"`
	IsSelf    bool      `json:"is_self"`
}

type TargetInfo struct {
	TargetId int64  `json:"target_id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

type ConversationListResult struct {
	ConversationId  int64           `json:"conversation_id"`
	IsPinned        bool            `json:"is_pinned"`
	IsMuted         bool            `json:"is_muted"`
	UnreadCount     int             `json:"unread_count"`
	LastMessageInfo LastMessageInfo `json:"last_message_info"`
	TargetInfo      TargetInfo      `json:"target_info"`
	Type            int             `json:"type"`
}

// 获取会话列表
func LoadConversationList(userId int64) ([]ConversationListResult, error) {

	type conversationListRow struct {
		ConversationId     int64  `gorm:"column:conversation_id"`
		Type               int    `gorm:"column:type"`
		CommunityId        *int64 `gorm:"column:community_id"`
		PrivateHightUserId *int64 `gorm:"column:private_high_user_id"`
		PrivateLowUserId   *int64 `gorm:"column:private_low_user_id"`

		IsPinned    bool `gorm:"column:is_pinned"`
		IsMuted     bool `gorm:"column:is_muted"`
		UnreadCount int  `gorm:"column:unread_count"`

		LastMessageId *int64     `gorm:"column:last_message_id"`
		LastFromId    *int64     `gorm:"column:last_from_id"`
		LastMedia     *int       `gorm:"column:last_media"`
		LastContent   *string    `gorm:"column:last_content"`
		LastPic       *string    `gorm:"column:last_pic"`
		LastUrl       *string    `gorm:"column:last_url"`
		LastDesc      *string    `gorm:"column:last_desc"`
		LastCreatedAt *time.Time `gorm:"column:last_created_at"`

		CommunityName   *string `gorm:"column:community_name"`
		CommunityAvatar *string `gorm:"column:community_avatar"`

		LowName    *string `gorm:"column:low_name"`
		LowAvatar  *string `gorm:"column:low_avatar"`
		HighName   *string `gorm:"column:high_name"`
		HighAvatar *string `gorm:"column:high_avatar"`
	}

	var rows []conversationListRow

	err := utils.DB.Table("conversation AS c").
		Select(`
			c.conversation_id,
			c.type,
			c.community_id,
			c.private_low_user_id,
			c.private_high_user_id,

			cm.is_pinned,
			cm.is_muted,
			cm.unread_count,

			m.message_id AS last_message_id,
			m.from_id AS last_from_id,
			m.media AS last_media,
			m.content AS last_content,
			m.pic AS last_pic,
			m.url AS last_url,
			m.desc AS last_desc,
			m.created_at AS last_created_at,

			g.name AS community_name,
			g.img AS community_avatar,

			u_low.name AS low_name,
			u_low.avatar AS low_avatar,
			u_high.name AS high_name,
			u_high.avatar AS high_avatar
		`).
		Joins(`
			JOIN conversation_member AS cm ON cm.conversation_id = c.conversation_id
		`).
		Joins(`
			LEFT JOIN message AS m ON m.message_id = c.last_message_id AND m.conversation_id = c.conversation_id
		`).
		Joins(`
			LEFT JOIN community AS g ON c.type = 1 AND g.community_id = c.community_id
		`).
		Joins(`
			LEFT JOIN user_basic AS u_low ON c.type = 2 AND u_low.user_id = c.private_low_user_id
		`).
		Joins(`
			LEFT JOIN user_basic AS u_high ON c.type = 2 AND u_high.user_id = c.private_high_user_id
		`).
		Where(`
			cm.user_id = ? AND cm.visible_at IS NOT NULL AND cm.left_at IS NULL AND c.status != 0
		`, userId).Order("cm.is_pinned DESC").Order("c.last_message_at IS NULL ASC").Order("c.last_message_at DESC").Order("c.conversation_id DESC").Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	result := make([]ConversationListResult, 0, len(rows))

	for _, row := range rows {
		item := ConversationListResult{
			ConversationId: row.ConversationId,
			IsPinned:       row.IsPinned,
			IsMuted:        row.IsMuted,
			UnreadCount:    row.UnreadCount,
			Type:           row.Type,
		}

		if row.Type == 1 && row.CommunityId != nil {
			item.TargetInfo = TargetInfo{
				TargetId: *row.CommunityId,
				Name:     *row.CommunityName,
				Avatar:   *row.CommunityAvatar,
			}
		}

		if row.Type == 2 {
			targetId := row.PrivateHightUserId
			targetName := row.HighName
			targetAvatar := row.HighAvatar

			if row.PrivateHightUserId != nil && *row.PrivateHightUserId == userId {
				targetId = row.PrivateLowUserId
				targetName = row.LowName
				targetAvatar = row.LowAvatar
			}

			if targetId != nil {
				item.TargetInfo = TargetInfo{
					TargetId: *targetId,
					Name:     *targetName,
					Avatar:   *targetAvatar,
				}
			}
		}

		if row.LastMessageId != nil {
			item.LastMessageInfo = LastMessageInfo{
				MessageId: *row.LastMessageId,
				Content:   *row.LastContent,
				Media:     *row.LastMedia,
				Pic:       *row.LastPic,
				Url:       *row.LastUrl,
				Desc:      *row.LastDesc,
				CreateAt:  *row.LastCreatedAt,
				IsSelf:    row.LastFromId != nil && *row.LastFromId == userId,
			}
		}

		result = append(result, item)
	}
	return result, nil
}

// 删除会话
func DeleteConversation(userId int64, conversationId int64) error {
	return utils.DB.Transaction(func(tx *gorm.DB) error {
		var conversation Conversation
		// 查询当前会话，锁住这条conversation记录, 防止删除和新增同时发生
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("conversation_id", "last_message_id").
			Where("conversation_id = ?", conversationId).
			Take(&conversation).Error; err != nil {
			return err
		}

		result := tx.Model(&ConversationMember{}).
			Where("user_id = ? AND conversation_id = ? AND left_at IS NULL", userId, conversationId).
			Updates(map[string]interface{}{
				"visible_at":              nil,
				"unread_count":            0,
				"clear_before_message_id": conversation.LastMessageId,
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}
