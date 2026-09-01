package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"github.com/hxysj/im-system/utils"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	MessageId      int64 `gorm:"not null;uniqueIndex:idx_message_conversation,priority:2" json:"message_id"`
	FromId         int64 //发送者
	ConversationId int64 `gorm:"not null;index:idx_message_conversation,priority:1"` //会话id
	// TargetId       int64  // 接收者
	// Type           int    // 消息类型 1群聊 2私聊 3广播
	Media   int    //消息类型 - 1文字 2表情包 3图片 4音频
	Content string //消息内容
	Pic     string
	Url     string
	Desc    string
	Amount  int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
	Done      chan struct{}
}

// 客户端的映射表
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// websocket.Upgrader 是 Gorilla WebSocket 库中的一个结构体，它的核心作用是将普通的 HTTP 连接升级为 WebSocket 连接。
func Chat(userId int64, writer http.ResponseWriter, request *http.Request) {
	isValid := true

	// coon 是websocket的连接
	coon, err := (&websocket.Upgrader{ //创建一个 Upgrader 实例（取地址）
		// token 校验 配置跨域检查
		CheckOrigin: func(r *http.Request) bool {
			// 是否允许跨域
			return isValid
		},
	}).Upgrade(writer, request, nil) // 调用 Upgrade 方法，升级连接

	if err != nil {
		fmt.Println("chat is error", err)
		return
	}

	// 构建一个客户端节点
	node := &Node{
		Conn:      coon,
		DataQueue: make(chan []byte, 50),   //消息队列（缓冲区 50）	异步发送消息的通道
		GroupSets: set.New(set.ThreadSafe), //群组集合（线程安全）	记录该用户加入了哪些群聊
		Done:      make(chan struct{}),
	}
	// 设置心跳机制的超时时间
	pongWait := 60 * time.Second
	// 收到pong响应后重置超时计时器
	node.Conn.SetReadDeadline(time.Now().Add(pongWait))
	// 实现心跳检测：确保连接任然活跃
	node.Conn.SetPongHandler(func(string) error {
		node.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// userId 和 node 绑定

	// 查找是否有旧连接存在
	rwLocker.Lock()

	oldNode := clientMap[int64(userId)]

	clientMap[int64(userId)] = node

	rwLocker.Unlock()

	if oldNode != nil {
		oldNode.Conn.Close()
	}

	// 发送消息的逻辑
	go sendProc(int64(userId), node)
	// 接收消息的逻辑
	go recvProc(int64(userId), node)

	sendMsg(int64(userId), []byte("欢迎进入im-system系统的聊天室！"))
}

// 发送消息 - 负责将客户端获取的消息发送到对应的目标上去
func sendProc(userId int64, node *Node) {
	// 心跳检测的逻辑
	// 每30秒发送一次Ping帧
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	// 无限循环，支持处理消息
	for {
		select {
		case <-ticker.C:
			// 发送Ping消息
			err := node.Conn.WriteControl(
				websocket.PingMessage,
				nil,
				time.Now().Add(10*time.Second),
			)
			// 心跳机制检测失败了 - 关闭连接
			if err != nil {
				_ = node.Conn.Close()
				return
			}
		// 监听到消息队列的数据发送变化的时候触发取出
		case data := <-node.DataQueue:
			fmt.Println("[send]<<<<<  ", string(data))
			// 通过websocket发送消息，websocket.TextMessage 知道消息为文本消息，data是从消息队列中获取的
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				node.Conn.Close()
				return
			}
		case <-node.Done:
			// 收到连接被关闭的时候退出这个节点的监听
			return

		}
	}
}

// 接收消息 -负责从 WebSocket 连接中读取客户端发来的消息
func recvProc(userId int64, node *Node) {
	// 监听客户端消息退出的时候触发
	defer func() {
		// 触发关闭websocket连接
		node.Conn.Close()
		close(node.Done)
		// 上锁
		rwLocker.Lock()
		// 将clientMap中对应的数据删除
		if current, ok := clientMap[userId]; ok && current == node {
			delete(clientMap, userId)
		}
		// 解锁
		rwLocker.Unlock()
	}()

	for {
		// websocket读取接收到的消息
		_, data, err := node.Conn.ReadMessage() //忽略消息类型，只取数据内容和错误
		if err != nil {
			// 客户端发送 ws.Close()的时候会触发
			//IsUnexpectedCloseError 判断Websocket关闭错误是否属于意外情况
			if websocket.IsUnexpectedCloseError(
				err,
				//状态码 1000 正常关闭
				websocket.CloseNormalClosure,
				//状态码  1001  浏览器或客户端离开
				websocket.CloseGoingAway,
			) {
				fmt.Println("WebSocket异常断开：", err)
			} else {
				fmt.Println("WebSocket连接关闭", err)
			}
			return
		}
		broadMsg(data) //广播消息
		fmt.Println("[ws]<<<<<<  ", string(data))
	}
}

var udpSendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpSendChan <- data //将收到的消息传入到广播的通道中去
}

// 初始化执行的函数
func init() {
	// 创建广播发送消息的进程
	go udpSendProc()
	// 船舰广播接收消息的进程
	go udpRecvProc()
	fmt.Println("init goruntine")
}

// UDP服务发送消息
func udpSendProc() {
	// 建立 UDP 连接
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 3000,
	})

	// 函数结束之前会触发执行，关闭UDP的服务
	defer con.Close()

	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		// 监听UDP通道的消息
		case data := <-udpSendChan:
			fmt.Println("[udpSendProc] >>> ", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// UDP接收消息
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero, // 0.0.0.0 监听所有网卡
		Port: 3000,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	defer con.Close()
	for {
		var buf [512]byte           //创建 512 字节的缓冲区
		n, err := con.Read(buf[0:]) //读取 UDP 数据到缓冲区
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[0:n]) //将收到的数据分发处理
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 要从conversation表中获取对应的成员
	var conversationMemberList []ConversationMember
	if err := utils.
		DB.
		Model(&ConversationMember{}).
		Where("conversation_id = ? AND user_id != ? AND left_at IS NULL", msg.ConversationId, msg.FromId).
		Find(&conversationMemberList).Error; err != nil {
		fmt.Println("获取会话信息表失败:", err)
		return
	}

	fmt.Println("[dispatch] >>> ", msg.FromId, string(data))
	// 将消息存储到数据库中

	msg.MessageId = utils.NextId()
	tx := utils.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&msg).Error; err != nil {
		tx.Rollback()
		fmt.Print("保存消息失败：", err)
		return
	}

	var conversation Conversation
	if err := tx.Model(&Conversation{}).Where("conversation_id = ?", msg.ConversationId).First(&conversation).Error; err != nil {
		tx.Rollback()
		fmt.Println("获取会话失败:", err)
		return
	}

	if err := tx.Model(&Conversation{}).
		Where("conversation_id = ?", msg.ConversationId).
		Updates(map[string]interface{}{
			"last_message_id": msg.MessageId,
			"last_message_at": msg.CreatedAt,
		}).Error; err != nil {
		tx.Rollback()
		fmt.Println("更新会话最后消息失败：", err)
		return
	}

	// 私聊或者群聊都是通过会话表获取目标用户信息，然后将数据发送给他们
	for _, conversationMember := range conversationMemberList {
		if err := tx.Model(&ConversationMember{}).
			Where("conversation_id = ? AND user_id = ? AND left_at IS NULL", msg.ConversationId, conversationMember.UserId).
			Updates(map[string]interface{}{
				"unread_count": gorm.Expr("unread_count + ?", 1),
				"visible_at":   msg.CreatedAt,
			}).Error; err != nil {
			tx.Rollback()
			return
		}

		sendMsg(conversationMember.UserId, data)
	}

	if err := tx.Model(&ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", msg.ConversationId, msg.FromId).
		Update("visible_at", msg.CreatedAt).Error; err != nil {
		tx.Rollback()
		fmt.Println("更新信息失败:", err)
		return
	}

	tx.Commit()

	for _, conversationMember := range conversationMemberList {
		sendMsg(conversationMember.UserId, data)
	}
}

func sendMsg(userId int64, msg []byte) {
	fmt.Println("[sendMsg] >>> ", userId, string(msg))
	rwLocker.Lock()
	// 获取接收者的实例
	node, ok := clientMap[userId]
	rwLocker.Unlock()
	if ok {
		// 将消息存入接收者实例的消息队列中
		node.DataQueue <- msg
	}
}

type MessageUserInfo struct {
	UserId int64  `json:"user_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type MessageListResult struct {
	MessageId int64           `json:"message_id"`
	FromInfo  MessageUserInfo `json:"from_info,omitempty"`
	Media     int             `json:"media"`
	Content   string          `json:"content"`
	Pic       string          `json:"pic"`
	Url       string          `json:"url"`
	Desc      string          `json:"desc"`
	CreatedAt time.Time       `json:"created_at"`
	IsSelf    bool            `json:"is_self"`
}

// 获取用户消息
// func GetMessageList(userId int64, targetId int64, msgType int, limit int, page int) ([]MessageListResult, int64, error) {
// 	if msgType == 1 {
// 		var community Community
// 		if err := utils.DB.Model(&Community{}).Where("community_id = ?", targetId).First(&community).Error; err != nil {
// 			return nil, 0, err
// 		}
// 		var contact Contact
// 		if err := utils.DB.Model(&Contact{}).Where("owen_id = ? AND target_id = ? AND type = 2", userId, targetId).First(&contact).Error; err != nil {
// 			return nil, 0, err
// 		}
// 	} else if msgType == 2 {
// 		var user UserBasic
// 		if err := utils.DB.Model(&UserBasic{}).Where("user_id = ?", targetId).First(&user).Error; err != nil {
// 			return nil, 0, err
// 		}
// 	}
// 	var total int64
// 	var messageList []Message
// 	// 私聊需要获取自己发送的还需要获取对方发送的
// 	query := utils.DB.Model(&Message{}).
// 		Where("(from_id = ? AND target_id = ? OR from_id = ? AND target_id = ?) AND type = ?",
// 			userId,
// 			targetId,
// 			targetId,
// 			userId,
// 			msgType)
// 	// 群聊则是获取所有人发送的
// 	if msgType == 1 {
// 		query = utils.DB.Model(&Message{}).
// 			Where("target_id = ? AND type = ?", targetId, msgType)
// 	}

// 	if err := query.Count(&total).Error; err != nil {
// 		return nil, 0, err
// 	}
// 	if err := query.Order("created_at DESC").
// 		Limit(limit).
// 		Offset((page - 1) * limit).
// 		Find(&messageList).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	userIdSet := make(map[int64]struct{}, 0)

// 	for _, message := range messageList {
// 		if msgType == 1 {
// 			if _, ok := userIdSet[message.FromId]; !ok {
// 				userIdSet[message.FromId] = struct{}{}
// 			}
// 		} else if msgType == 2 {
// 			if _, ok := userIdSet[message.FromId]; !ok {
// 				userIdSet[message.FromId] = struct{}{}
// 			}
// 			if _, ok := userIdSet[message.TargetId]; !ok {
// 				userIdSet[message.TargetId] = struct{}{}
// 			}
// 		}
// 	}

// 	userIdList := make([]int64, 0, len(userIdSet))

// 	for key, _ := range userIdSet {
// 		userIdList = append(userIdList, key)
// 	}

// 	var userInfoList []UserBasic

// 	if err := utils.DB.Model(&UserBasic{}).Where("user_id IN ?", userIdList).Find(&userInfoList).Error; err != nil {
// 		return nil, 0, err
// 	}

// 	userInfoMap := make(map[int64]UserBasic, len(userInfoList))

// 	for _, user := range userInfoList {
// 		userInfoMap[user.UserId] = user
// 	}

// 	result := make([]MessageListResult, 0, len(messageList))

// 	for _, message := range messageList {
// 		res := MessageListResult{
// 			MessageId: message.MessageId,
// 			Type:      message.Type,
// 			Media:     message.Media,
// 			Content:   message.Content,
// 			Pic:       message.Pic,
// 			Url:       message.Url,
// 			Desc:      message.Desc,
// 			CreatedAt: message.CreatedAt,
// 			IsSelf:    false,
// 		}
// 		if message.Type == 1 {
// 			// 群聊消息
// 			fromUser, _ := userInfoMap[message.FromId]
// 			res.FromInfo = MessageUserInfo{
// 				UserId: fromUser.UserId,
// 				Name:   fromUser.Name,
// 				Avatar: fromUser.Avatar,
// 			}
// 			if fromUser.UserId == userId {
// 				res.IsSelf = true
// 			}
// 		} else if message.Type == 2 {
// 			// 私聊消息
// 			fromUser, _ := userInfoMap[message.FromId]
// 			targetUser, _ := userInfoMap[message.TargetId]
// 			res.FromInfo = MessageUserInfo{
// 				UserId: fromUser.UserId,
// 				Name:   fromUser.Name,
// 				Avatar: fromUser.Avatar,
// 			}
// 			res.TargetInfo = MessageUserInfo{
// 				UserId: targetUser.UserId,
// 				Name:   targetUser.Name,
// 				Avatar: targetUser.Avatar,
// 			}
// 			if fromUser.UserId == userId {
// 				res.IsSelf = true
// 			}
// 		}

// 		result = append(result, res)
// 	}

// 	return result, total, nil
// }

// 根据会话id获取消息列表
func GetMessageList(userId int64, conversation_id int64, limit int, page int) ([]MessageListResult, int64, error) {
	var conversationMember ConversationMember

	if err := utils.DB.Model(&ConversationMember{}).
		Where("user_id = ? AND conversation_id = ?", userId, conversation_id).
		First(&conversationMember).
		Error; err != nil {
		return nil, 0, err
	}

	var total int64
	var messageList []Message
	query := utils.DB.Model(&Message{}).
		Where("conversation_id = ? AND message_id > ?", conversation_id, conversationMember.ClearBeforeMessageId)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("message_id DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&messageList).Error; err != nil {
		return nil, 0, err
	}

	userIdSet := make(map[int64]struct{}, 0)

	for _, message := range messageList {
		userIdSet[message.FromId] = struct{}{}
	}

	userIdList := make([]int64, 0, len(userIdSet))

	for key, _ := range userIdSet {
		userIdList = append(userIdList, key)
	}

	var userInfoList []UserBasic

	if err := utils.DB.Model(&UserBasic{}).
		Where("user_id IN ?", userIdList).
		Find(&userInfoList).Error; err != nil {
		return nil, 0, err
	}

	userInfoSet := make(map[int64]UserBasic, len(userInfoList))

	for _, userInfo := range userInfoList {
		userInfoSet[userInfo.UserId] = userInfo
	}

	result := make([]MessageListResult, 0, len(messageList))
	for _, message := range messageList {
		result = append(result, MessageListResult{
			MessageId: message.MessageId,
			FromInfo: MessageUserInfo{
				UserId: userInfoSet[message.FromId].UserId,
				Name:   userInfoSet[message.FromId].Name,
				Avatar: userInfoSet[message.FromId].Avatar,
			},
			Media:     message.Media,
			Content:   message.Content,
			Pic:       message.Pic,
			Url:       message.Url,
			Desc:      message.Desc,
			CreatedAt: message.CreatedAt,
			IsSelf:    message.FromId == userId,
		})
	}

	return result, total, nil
}

func ReadMessage(userId int64, conversation_id int64) error {

	// var conversation Conversation

	// if err := utils.DB.Model(&Conversation{}).Where("conversation_id = ?", conversation_id).First(&conversation).Error; err != nil {
	// 	fmt.Println("获取会话失败！", err)
	// 	return err
	// }

	// var conversationMember ConversationMember

	// if err := utils.DB.Model(&ConversationMember{}).Where("user_id = ? AND conversation_id = ? AND left_at is NULL", userId, conversation_id).First(&conversationMember).Error; err != nil {
	// 	fmt.Println("获取会话关系失败！", err)
	// 	return err
	// }

	type readConversationRow struct {
		LastMessageId int `gorm:"column:last_message_id"`
	}

	var row readConversationRow
	// Take(&row) 查询一条匹配记录，并把结果放到 row 中
	err := utils.DB.Table("conversation AS C").
		Select("c.last_message_id").
		Joins(`INNER JOIN conversation_member AS cm ON cm.conversation_id = c.conversation_id`).
		Where(`c.conversation_id = ? AND cm.user_id = ? AND cm.left_at IS NULL`, conversation_id, userId).Take(&row).Error

	if err != nil {
		fmt.Println("获取会话或会话成员失败", err)
		return err
	}

	if err := utils.DB.Model(&ConversationMember{}).
		Where("user_id = ? AND conversation_id = ? AND left_at IS NULL", userId, conversation_id).
		Updates(map[string]interface{}{
			// "last_read_message_id": conversation.LastMessageId,
			"last_read_message_id": row.LastMessageId,
			"unread_count":         0,
		}).Error; err != nil {
		fmt.Println("更换会话信息失败:", err)
		return err
	}
	return nil
}
