package utils

const (
	ConversationStatusNormal    = 1 // 正常
	ConversationStatusMuted     = 2 // 群禁言
	ConversationStatusDissolved = 3 // 会话已解散
	ConversationStatusLeft      = 4 // 当前用户已退出群聊，仅详情返回
	ConversationStatusNotFriend = 5 // 好友关系失效，仅详情返回
)
