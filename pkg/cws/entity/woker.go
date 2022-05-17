package entity

import "xmediaEmu/pkg/cws"

const (
	// 视频增强处理.
	GetRoom   = "get_room"
	CloseRoom = "close_room"
	// RegisterRoom = "register_room"

	RoomStart   = "start"
	RoomStarted = "started"
	RoomQuit    = "quit"
)

// RoomStart对应的命令.
type RoomStartCall struct {
	Name string `json:"name"`
	// UnixSocket bool   `json:"unixsocket"`
	AudioPayloadType int `json:"audioPayload"`
	VideoPayloadType int `json:"videoPayload"`

	// TODO: 定义ip 和port直接传送.
	Zone string `json:"zone,omitempty"` // default: udp
	Addr string `json:"addr,omitempty"` // ip:port string
}

func (packet *RoomStartCall) From(data string) error { return from(packet, data) }
func (packet *RoomStartCall) To() (string, error)    { return to(packet) }

// RoomStart对应的响应.
type RoomStartRsp struct {
	RoomId string `json:"room"`
	Sdp    string `json:"sdp,omitempty"` // SdpOffer结构体.
}

func (packet *RoomStartRsp) From(data string) error { return from(packet, data) }
func (packet *RoomStartRsp) To() (string, error)    { return to(packet) }

type ConnectionRequest struct {
	Zone string `json:"zone,omitempty"` // default: udp
	Addr string `json:"addr,omitempty"`
	// IsHTTPS  bool   `json:"is_https,omitempty"`
}

// close room.data只有roomId.
func CloseRoomPacket(data string) cws.WSPacket { return cws.WSPacket{ID: CloseRoom, Data: data} }

// func RegisterRoomPacket(data string) cws.WSPacket { return cws.WSPacket{ID: RegisterRoom, Data: data} }
