package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"simple-chatroom/utils"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

// 消息
type Message struct {
	gorm.Model
	UserId     int64  `json:"UserId"`     //发送者
	TargetId   int64  `json:"TargetId"`   //接受者
	Type       int    `json:"Type"`       //发送类型  1私聊  2群聊  3心跳
	Media      int    `json:"Media"`      //消息类型  1文字 2表情包 3语音 4图片 /表情包
	Content    string `json:"Content"`    //消息内容
	CreateTime uint64 `json:"CreateTime"` //创建时间
	ReadTime   uint64 `json:"ReadTime"`   //读取时间
	Pic        string `json:"Pic"`
	Url        string `json:"Url"`
	Desc       string `json:"Desc"`
	Amount     int    `json:"Amount"` //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

// const (
// 	HeartbeatMaxTime = 1 * 60
// )

type Node struct {
	Conn          *websocket.Conn //连接
	Addr          string          //客户端地址
	FirstTime     uint64          //首次连接时间
	HeartbeatTime uint64          //心跳时间
	LoginTime     uint64          //登录时间
	DataQueue     chan []byte     //消息
	GroupSets     set.Interface   //好友 / 群
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// 需要 ：发送者ID ，接受者ID ，消息类型，发送的内容，发送类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.  获取参数 并 检验 token 等合法性
	query := request.URL.Query()
	Id := query.Get("userId")
	token := query.Get("token")
	userId, _ := strconv.ParseInt(Id, 10, 64)

	// 直接使用 ParseJwt 进行JWT token验证
	if token == "" {
		fmt.Printf("WebSocket连接失败: 用户%d token为空\n", userId)
		http.Error(writer, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	claims, err := ParseJwt(token, "secretKey")
	if err != nil {
		fmt.Printf("WebSocket连接失败: 用户%d JWT解析失败: %v\n", userId, err)
		http.Error(writer, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	// 验证用户ID是否匹配
	if int64(claims.UserID) != userId {
		fmt.Printf("WebSocket连接失败: 用户ID不匹配, token中的ID=%d, 请求的ID=%d\n", claims.UserID, userId)
		http.Error(writer, "Unauthorized: User ID mismatch", http.StatusUnauthorized)
		return
	}

	fmt.Printf("WebSocket连接成功: 用户=%s, ID=%d\n", claims.Username, claims.UserID)

	conn, err := (&websocket.Upgrader{
		//token 校验
		CheckOrigin: func(r *http.Request) bool {
			return true // token已在上面验证过
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//2.获取conn
	currentTime := uint64(time.Now().Unix())
	node := &Node{
		Conn:          conn,
		Addr:          conn.RemoteAddr().String(), //客户端地址
		HeartbeatTime: currentTime,                //心跳时间
		LoginTime:     currentTime,                //登录时间
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe),
	}
	//3. 用户关系
	//4. userid 跟 node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()
	//5.完成发送逻辑
	go sendProc(node)
	//6.完成接受逻辑
	go recvProc(node)
	//7.加入在线用户到缓存
	SetUserOnlineInfo("online_"+Id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)

	//sendMsg(userId, []byte("欢迎进入聊天系统"))

}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws]sendProc >>>> msg :", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		msg := Message{}
		err = json.Unmarshal(data, &msg)
		if err != nil {
			fmt.Println(err)
		}
		//心跳检测 msg.Media == -1 || msg.Type == 3
		if msg.Type == 3 {
			currentTime := uint64(time.Now().Unix())
			node.Heartbeat(currentTime)
		} else {
			dispatch(data)
			broadMsg(data) //todo 将消息广播到局域网
			fmt.Println("[ws] recvProc <<<<< ", string(data))
		}

	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine ")
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255),
		Port: viper.GetInt("port.udp"),
	})

	if err != nil {
		fmt.Println(err)
	}
	if con == nil {
		fmt.Println("con is nil")
		return
	}
	defer con.Close()
	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udpSendProc  data :", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

// 完成udp数据接收协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: viper.GetInt("port.udp"),
	})
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc  data :", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{}
	msg.CreateTime = uint64(time.Now().Unix())
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: //私信
		fmt.Println("dispatch  data :", string(data))
		sendMsg(msg.TargetId, data)
	case 2: //群发
		sendGroupMsg(msg.TargetId, data)
		// case 4: // 心跳
		// node.Heartbeat()
		//case 4:
		//
	}
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息")
	userIds := SearchUserByGroupId(uint(targetId))
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)

	// 保存群聊消息到Redis
	ctx := context.Background()
	groupKey := "group_msg_" + strconv.Itoa(int(targetId))
	res, err := utils.Red.ZRevRange(ctx, groupKey, 0, -1).Result()
	if err != nil {
		fmt.Println("Redis ZRevRange error:", err)
	}
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, groupKey, &redis.Z{Score: score, Member: msg}).Result()
	if e != nil {
		fmt.Println("Redis ZAdd error:", e)
	} else {
		fmt.Println("群聊消息已保存到Redis:", ress)
		// 设置4小时过期时间
		utils.Red.Expire(ctx, groupKey, 4*time.Hour)
	}

	// 发送给所有群成员（包括发送者，用于确认消息发送成功）
	for i := 0; i < len(userIds); i++ {
		sendMsgToUser(int64(userIds[i]), msg)
	}
}

// 新增：单独发送消息给用户的函数
func sendMsgToUser(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()

	if ok {
		fmt.Println("sendMsgToUser >>> userID: ", userId, "  msg:", string(msg))
		select {
		case node.DataQueue <- msg:
			// 消息发送成功
		default:
			fmt.Println("用户消息队列已满，消息丢弃")
		}
	} else {
		fmt.Println("用户不在线: ", userId)
	}
}

func JoinGroup(userId uint, comId string) (int, string) {
	contact := Contact{}
	contact.OwnerId = userId
	//contact.TargetId = comId
	contact.Type = 2
	community := Community{}

	utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
	if community.Name == "" {
		return -1, "没有找到群"
	}
	utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, comId).Find(&contact)
	if !contact.CreatedAt.IsZero() {
		return -1, "已加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}
}

func sendMsg(userId int64, msg []byte) {

	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> userID: ", userId, "  msg:", string(msg))
			node.DataQueue <- msg
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg
	//res, e := utils.Red.Do(ctx, "zadd", key, 1, jsonMsg).Result() //备用 后续拓展 记录完整msg
	if e != nil {
		fmt.Println(e)
	} else {
		// 设置4小时过期时间
		utils.Red.Expire(ctx, key, 4*time.Hour)
	}
	fmt.Println(ress)
}

// 需要重写此方法才能完整的msg转byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

// 获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	//node, ok := clientMap[userIdA]
	rwLocker.RUnlock()
	//jsonMsg := Message{}
	//json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	//key = "msg_" + userIdStr + "_" + targetIdStr
	//rels, err := utils.Red.ZRevRange(ctx, key, 0, 10).Result()  //根据score倒叙

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) //没有找到
	}
	// 发送推送消息
	/**
	// 后台通过websoket 推送消息
	for _, val := range rels {
		fmt.Println("sendMsg >>> userID: ", userIdA, "  msg:", val)
		node.DataQueue <- []byte(val)
	}**/
	return rels
}

// 更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

// 清理超时连接
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("cleanConnection err", r)
		}
	}()
	//fmt.Println("定时任务,清理超时连接 ", param)
	//node.IsHeartbeatTimeOut()
	currentTime := uint64(time.Now().Unix())
	for i := range clientMap {
		node := clientMap[i]
		if node.IsHeartbeatTimeOut(currentTime) {
			fmt.Println("心跳超时..... 关闭连接：", node)
			node.Conn.Close()
		}
	}
	return result
}

// 用户心跳是否超时
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		fmt.Println("心跳超时。。。自动下线", node)
		timeout = true
	}
	return
}

// 获取群聊缓存消息
func RedisGroupMsg(groupId int64, start int64, end int64, isRev bool) []string {
	ctx := context.Background()
	groupKey := "group_msg_" + strconv.Itoa(int(groupId))

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, groupKey, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, groupKey, start, end).Result()
	}
	if err != nil {
		fmt.Println("获取群聊历史消息失败:", err)
	} else {
		fmt.Println("获取群聊历史消息成功，消息数量:", len(rels))
	}
	return rels
}
