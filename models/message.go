package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/fatih/set"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	FromId int64 //发送者
	TargetId int64 // 接收者
	Type int  // 消息类型 1群聊 2私聊 3广播
	Media int //消息类型 - 1文字 2表情包 3图片 4音频
	Content string //消息内容
	Pic string
	Url string
	Desc string
	Amount int //其他数字统计
}

func (table *Message) TableName() string{
	return "message"
}

type Node struct{
	Conn *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 客户端的映射表
var clientMap map[int64]*Node = make(map[int64]*Node,0)

// 读写锁
var rwLocker sync.RWMutex

// websocket.Upgrader 是 Gorilla WebSocket 库中的一个结构体，它的核心作用是将普通的 HTTP 连接升级为 WebSocket 连接。
func Chat(writer http.ResponseWriter,request *http.Request){
	// 获取请求携带的参数
	query := request.URL.Query()
	// 校验token
	// token := query.Get("token")
	userId,_ := strconv.ParseInt(query.Get("userId"),10,36)
	// 第一个参数 s：要解析的字符串（query.Get("userId") 返回的字符串）
	// 第二个参数 base：进制数（这里是 0，表示自动检测进制）
	// 第三个参数 bitSize：结果类型（36 表示要解析成 int64 类型）
	// msgType := query.Get("type")
	// targetId := query.Get("targetId")
	// context := query.Get("context")

	isValid := true

	// coon 是websocket的连接
	coon,err := (&websocket.Upgrader{  //创建一个 Upgrader 实例（取地址）
		// token 校验 配置跨域检查
		CheckOrigin: func(r *http.Request) bool{
			return isValid
		},
	}).Upgrade(writer,request,nil) // 调用 Upgrade 方法，升级连接

	if err != nil{
		fmt.Println("chat is error",err)
		return
	}

	// 构建一个客户端节点
	node := &Node{
		Conn: coon,
		DataQueue: make(chan []byte,50),   //消息队列（缓冲区 50）	异步发送消息的通道
		GroupSets: set.New(set.ThreadSafe), //群组集合（线程安全）	记录该用户加入了哪些群聊

	}

	// 用户关系

	// userId 和 node 绑定
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	// 发送消息的逻辑
	go sendProc(node)
	// 接收消息的逻辑
	go recvProc(node)

	sendMsg(userId,[]byte("欢迎进入im-system系统的聊天室！"))
}


// 发送消息
func sendProc(node *Node){
	// 无限循环，支持处理消息
	for {
		select{
		// 监听到消息队列的数据发送变化的时候触发取出
		case data:= <-node.DataQueue:
			fmt.Println("[send]<<<<<  ",string(data))
			// 通过websocket发送消息，websocket.TextMessage 知道消息为文本消息，data是从消息队列中获取的
			err := node.Conn.WriteMessage(websocket.TextMessage,data)
			if err !=nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 接收消息 -负责从 WebSocket 连接中读取客户端发来的消息
func recvProc(node *Node){
	for{
		// websocket读取接收到的消息
		_,data,err := node.Conn.ReadMessage() //忽略消息类型，只取数据内容和错误
		if err != nil{
			fmt.Println(err)
			return
		}
		broadMsg(data)  //广播消息
		fmt.Println("[ws]<<<<<<  ",string(data))
	}
}

var udpSendChan chan []byte = make(chan []byte,1024)

func broadMsg(data []byte){
	udpSendChan <- data  //将收到的消息传入到广播的通道中去
}

// 初始化执行的函数
func init(){
	// 创建广播发送消息的进程
	go udpSendProc()
	// 船舰广播接收消息的进程
	go udpRecvProc()
	fmt.Println("inti goruntine")
}

// UDP服务发送消息
func udpSendProc(){
	// 建立 UDP 连接
	con,err := net.DialUDP("udp",nil,&net.UDPAddr{
		IP: net.IPv4(127,0,0,1),
		Port:3000,
	})

	// 函数结束之前会触发执行，关闭UDP的服务
	defer con.Close()

	if err != nil{
		fmt.Println(err)
	}

	for{
		select {
		// 监听UDP通道的消息
		case data := <-udpSendChan:
			fmt.Println("[udpSendProc] >>> ",string(data))
			_,err := con.Write(data)
			if err != nil{
				fmt.Println(err)
				return
			}
		}
	}
}

// UDP接收消息
func udpRecvProc(){
	con,err := net.ListenUDP("udp", &net.UDPAddr{
		IP: net.IPv4zero,  // 0.0.0.0 监听所有网卡
		Port: 3000,
	})

	if err != nil{
		fmt.Println(err)
		return
	}

	defer con.Close()
	for {
		var buf [512]byte  //创建 512 字节的缓冲区
		n,err := con.Read(buf[0:])  //读取 UDP 数据到缓冲区
		if err !=nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[0:n]) //将收到的数据分发处理
	}
}

// 后端调度逻辑处理
func dispatch(data []byte){
	msg := Message{}
	err := json.Unmarshal(data,&msg)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println("[dispatch] >>> ",msg.TargetId,string(data))
	// 将消息存储到数据库中

	switch msg.Type{
	// 检测消息类型是1的
	case 2:
		sendMsg(msg.TargetId,data)
	}
}


func sendMsg(userId int64,msg []byte){
	fmt.Println("[sendMsg] >>> ",userId,string(msg))
	rwLocker.Lock()
	// 获取接收者的实例
	node,ok := clientMap[userId]
	rwLocker.Unlock()
	if ok{
		// 将消息存入接收者实例的消息队列中
		node.DataQueue <- msg
	}
}
