package cws

import (
	"common/util/process"
	"common/web"
	"encoding/json"
	"github.com/pterm/pterm"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
	"xmediaEmu/pkg/log"
)

// 定义通用消息和ws接口.
type (
	//
	WSPacket struct {
		ID   string `json:"id"`   // 命令id.
		Data string `json:"data"` // 自定义任何数据.
		//RoomID      string `json:"room_id"`
		//PlayerIndex int    `json:"player_index"`

		PacketID string `json:"packet_id"` // packet包id.
		// Globally ID of a sip session
		SessionID string `json:"session_id"`

		// sessions n->1 with one room
		RoomID string `json:"room_id"`
	}

	// 一个client代表一段连接，全双工.
	Client struct {
		id   string // = sessionid
		conn *web.WSocket

		// sendCallback is callback based on packetID
		// 根据消息类型进行分发.
		sendCallback     map[string]func(req WSPacket)
		sendCallbackLock sync.Mutex
		// recvCallback is callback when receive based on ID of the packet
		recvCallback map[string]func(req WSPacket)
		//Done chan struct{}
	}

	PacketHandler func(resp WSPacket) (req WSPacket)
)

var (
	EmptyPacket = WSPacket{}
	// HeartbeatPacket = WSPacket{ID: "heartbeat"} // 心跳后续加.
)

const WSWait = 20 * time.Second

// 基于gin框架
func NewClient(serverId string, conn *web.WSocket) *Client {
	id := serverId
	sendCallback := map[string]func(WSPacket){}
	recvCallback := map[string]func(WSPacket){}

	return &Client{
		id:           id,
		conn:         conn,
		sendCallback: sendCallback,
		recvCallback: recvCallback,
		//Done: make(chan struct{}),
	}
}

// 设置注册接收处理.
// Receive receive and response back
func (c *Client) Receive(id string, f PacketHandler) {
	c.recvCallback[id] = func(response WSPacket) {
		defer func() {
			if v := recover(); v != nil {
				process.DefaultPanicReport.RecoverFromPanic(c.id, "CWS.Client.Receive", v)
			}
		}()

		req := f(response)
		// Add Meta data
		req.PacketID = response.PacketID
		req.SessionID = response.SessionID

		// Skip response if it is EmptyPacket
		if response == EmptyPacket {
			return
		}
		resp, err := json.Marshal(req)
		if err != nil {
			log.Logger.Errorf("[!] json marshal error:%v", err)
		}

		// 发送
		c.conn.SendTextWithDeadLine(string(resp), WSWait)
	}
}

// Send sends a packet and trigger callback when the packet comes back
func (c *Client) Send(request WSPacket, callback func(response WSPacket)) {
	request.PacketID = uuid.NewV4().String()
	data, err := json.Marshal(request)
	if err != nil {
		return
	}

	// TODO: Consider using lock free
	// Wrap callback with sessionID and packetID
	if callback != nil {
		wrapperCallback := func(resp WSPacket) {
			defer func() {
				if v := recover(); v != nil {
					process.DefaultPanicReport.RecoverFromPanic(c.id, "CWS.Client.Send", v)
				}
			}()

			resp.PacketID = request.PacketID
			resp.SessionID = request.SessionID
			callback(resp)
		}
		c.sendCallbackLock.Lock()
		c.sendCallback[request.PacketID] = wrapperCallback
		c.sendCallbackLock.Unlock()
	}

	c.conn.SendTextWithDeadLine(string(data), WSWait)
}

// 没注册成功?..
func (c *Client) Listen() {
	c.conn.OnTextMessage = c.OnTextMessage
	c.conn.OnBinaryMessage = c.OnBinaryMessage
	c.conn.OnDisconnected = c.Dispose
	c.conn.OnConnectError = c.Dispose

	c.conn.Start()
}

// 没注册成功?..
func (c *Client) Connect() error {
	c.conn.OnTextMessage = c.OnTextMessage
	c.conn.OnBinaryMessage = c.OnBinaryMessage
	// c.conn.OnDisconnected = c.Dispose
	c.conn.OnConnectError = c.Dispose

	return c.conn.Connect()
}

func (c *Client) OnTextMessage(message string, socket web.WSocket) {
	c.OnBinaryMessage([]byte(message), socket)
}

// 二进制流也当做string类型处理.
func (c *Client) OnBinaryMessage(data []byte, socket web.WSocket) {
	pterm.FgWhite.Println("OnBinaryMessage enter")
	wspacket := WSPacket{}
	err := json.Unmarshal(data, &wspacket)
	if err != nil {
		log.Logger.Error("Warn: error decoding", string(data))
		return
	}
	pterm.FgWhite.Printfln("OnBinaryMessage message is:%v c.recvCallback:%v", wspacket, c.recvCallback)

	callback, ok := c.sendCallback[wspacket.PacketID]
	//c.sendCallbackLock.Unlock()
	if ok {
		pterm.FgWhite.Println("OnBinaryMessage sendCallback")
		go callback(wspacket)
		//c.sendCallbackLock.Lock()
		delete(c.sendCallback, wspacket.PacketID)
		//c.sendCallbackLock.Unlock()
		// Skip receiveCallback to avoid duplication
		return
	}

	// Check if some receiver with the ID is registered
	if callback, ok := c.recvCallback[wspacket.ID]; ok {
		go callback(wspacket)
	}
}

func (c *Client) Dispose(err error, socket web.WSocket) {
	log.Logger.Infof("CloseSocket from:%v %v", socket.Url, err)
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// 同步调用知道有响应.
func (c *Client) SyncSend(request WSPacket) (response WSPacket) {
	res := make(chan WSPacket)
	f := func(resp WSPacket) {
		res <- resp
	}
	c.Send(request, f)
	return <-res
}
