package worker

import (
	"common/web"
	"fmt"
	"github.com/gorilla/websocket"
	"xmediaEmu/pkg/cws"
	"xmediaEmu/pkg/cws/entity"
	"xmediaEmu/pkg/emulator/config"
	"xmediaEmu/pkg/emulator/rtpua"
	"xmediaEmu/pkg/log"
)

// one server for all users.
type Handler struct {
	// service.RunnableService
	wsConn *websocket.Conn
	// Client that connects to coordinator
	oClient *cws.Client //ws连接.
	cfg     config.Config

	// Rooms map : RoomID -> Room
	// rooms map[string]*Room
	// global ID of the current server
	// serverID string

	// sessions handles all sessions server is handler (key is sessionID)
	sessions map[string]*Session //保留map实例常量.
	// Rooms map : RoomID -> Room
	rooms map[string]*Room //支持启动多个room
}

// 支持ws复用，gin 框架.
// ws的命令server只要一个通道即可.
func NewHandler(conf config.Config, conn *websocket.Conn) *Handler {
	return &Handler{
		wsConn:   conn,
		cfg:      conf,
		rooms:    map[string]*Room{},
		sessions: map[string]*Session{},
	}
}

// Run starts a Handler running logic
func (h *Handler) Run(serverId string) {
	h.oClient = cws.NewClient(serverId, web.NewWSocket(h.wsConn))
	h.routes()
	log.Logger.Debugf("[worker] connected from: %v", h.wsConn.RemoteAddr())

	// block here.
	h.oClient.Listen()
}

// 根据roomId创建room，如果之前有room则直接返回EmptyPacket.
// client端收到响应后需要绑定对应ip和端口.
func (h *Handler) handleRoomStart() cws.PacketHandler {
	return func(resp cws.WSPacket) (req cws.WSPacket) {
		log.Logger.Info("Received a start request from coordinator")
		// TODO: Standardize for all types of packet. Make WSPacket generic
		rom := entity.RoomStartCall{}
		if err := rom.From(resp.Data); err != nil {
			return cws.EmptyPacket
		}

		session := h.getSession(resp.SessionID)
		if session != nil {
			log.Logger.Errorf("error: session already exist with id: %s", resp.SessionID)
			return cws.EmptyPacket
		}

		// 创建 session.
		if session = h.newSession(resp.SessionID, &rom); session == nil {
			log.Logger.Errorf("error: no session with id: %s", resp.SessionID)
			return cws.EmptyPacket
		}

		sdpAnswer, err := session.peerconnection.StartClient(rom.Addr)
		if err != nil {
			log.Logger.Errorf("error: StartClient failed: %v, peer:%s", err, rom.Addr)
			return cws.EmptyPacket
		}

		// game := games.GameMetadata{Name: rom.Name, Type: rom.Type, Base: rom.Base, Path: rom.Path}
		room := h.startGameHandler(rom.Name, h.cfg.Encoder.BUseUnixSocket, resp.RoomID, 0, session.peerconnection) // playerIndex暂时不用，多方通话时再填上.
		session.room = room
		// TODO: can data race (and it does)
		h.rooms[room.ID] = room
		rsp := entity.RoomStartRsp{RoomId: room.ID, Sdp: sdpAnswer}
		data, _ := rsp.To()
		return cws.WSPacket{ID: entity.RoomStarted, RoomID: room.ID, SessionID: resp.SessionID, PacketID: req.PacketID, Data: data}
	}
}

// 退出房间.
func (h *Handler) handleRoomQuit() cws.PacketHandler {
	return func(resp cws.WSPacket) (req cws.WSPacket) {
		log.Logger.Info("Received a quit request from client")
		session := h.getSession(resp.SessionID)

		if session != nil {
			room := h.getRoom(session.ID)
			// Defensive coding, check if the peerconnection is in room
			if room.IsPCInRoom(session.peerconnection) {
				h.detachPeerConn(session.peerconnection)
			}
		} else {
			log.Logger.Warnf("Error: No session for ID: %s\n", resp.SessionID)
		}

		return cws.EmptyPacket
	}
}

// TODO: 实例循环利用，不要临时创建.
func (h *Handler) newSession(sessionId string, startCall *entity.RoomStartCall) *Session {
	// rptua初始化.
	// one port for both audio and video.
	portSuit, err := GetPortSuiteHelper().AllotPort()
	if err != nil {
		log.Logger.Errorf("error: AllotPort failed: %v", err)
		return nil
	}
	udpAddr := fmt.Sprintf("%s:%d", h.cfg.LocalMediaIp, portSuit.PortRtp)
	peerConnection, err := rtpua.NewWebRTC(rtpua.Config{Encoder: h.cfg.Encoder, NetWork: startCall.Zone, UdpAddr: udpAddr, SessionId: sessionId}, startCall.AudioPayloadType, startCall.VideoPayloadType)
	if err != nil {
		log.Logger.Errorf("error: rtpua.NewWebRTC failed: %v", err)
		return nil
	}

	session := &Session{}
	session.peerconnection = peerConnection
	session.ID = sessionId
	return session
}

// getRoom returns session from sessionID
func (h *Handler) getSession(sessionID string) *Session {
	session, ok := h.sessions[sessionID]
	if !ok {
		return nil
	}

	return session
}

func (h *Handler) getRoom(roomID string) (r *Room) {
	r, ok := h.rooms[roomID]
	if !ok {
		return nil
	}
	return
}

// detachRoom detach room from Handler
func (h *Handler) detachRoom(roomID string) {
	delete(h.rooms, roomID)
}

// createNewRoom creates a new room
// Return nil in case of room is existed
func (h *Handler) createNewRoom(game string, bUseUnixSocket bool, roomID string) *Room {
	// If the roomID doesn't have any running sessions (room was closed)
	// we spawn a new room
	if !h.isRoomBusy(roomID) {
		newRoom := NewRoom(roomID, game, bUseUnixSocket, h.cfg)
		// TODO: Might have race condition (and it has (:)
		h.rooms[newRoom.ID] = newRoom
		return newRoom
	}
	return nil
}

// isRoomBusy check if there is any running sessions.
// TODO: If we remove sessions from room anytime a session is closed,
// we can check if the sessions list is empty or not.
func (h *Handler) isRoomBusy(roomID string) bool {
	if roomID == "" {
		return false
	}
	// If no roomID is registered
	r, ok := h.rooms[roomID]
	if !ok {
		return false
	}
	return r.IsRunningSessions()
}
