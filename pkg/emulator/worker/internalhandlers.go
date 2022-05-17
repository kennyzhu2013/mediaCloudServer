package worker

import (
	"xmediaEmu/pkg/cws/entity"
	"xmediaEmu/pkg/emulator/rtpua"
	"xmediaEmu/pkg/log"
)

// startGameHandler starts a game if roomID is given, if not create new room
func (h *Handler) startGameHandler(gameName string, bUseUnixSocket bool, existedRoomID string, playerIndex int, peerconnection *rtpua.RtpUa) *Room {
	log.Logger.Infof("Loading game: %v\n", gameName)
	// If we are connecting to coordinator, request corresponding serverID based on roomID
	// TODO: check if existedRoomID is in the current server
	room := h.getRoom(existedRoomID)
	// If room is not running
	if room == nil {
		log.Logger.Info("Got Room from local ", room, " ID: ", existedRoomID)
		// Create new room and update player index
		room = h.createNewRoom(gameName, bUseUnixSocket, existedRoomID)
		room.UpdatePlayerIndex(peerconnection, playerIndex)

		// Wait for done signal from room
		go func() {
			// room没有client时关闭，通知client端关闭了.
			<-room.Done
			h.detachRoom(room.ID)
			// send signal to coordinator that the room is closed, then client will remove that room
			// no session left, remove it.
			h.oClient.Send(entity.CloseRoomPacket(room.ID), nil)
		}()
	}

	// Attach peerconnection to room. If PC is already in room, don't detach
	log.Logger.Infof("startGameHandler Is PC in room:%v", room.IsPCInRoom(peerconnection))
	if !room.IsPCInRoom(peerconnection) {
		h.detachPeerConn(peerconnection)
		room.AddConnectionToRoom(peerconnection)
	}

	// Register room to coordinator if we are connecting to coordinator
	if room != nil && h.oClient != nil {
		// 直接在response里一条消息返回给client.
		// h.oClient.Send(entity.RegisterRoomPacket(room.ID), nil)
	}

	return room
}

// detachPeerConn detaches a peerconnection from the current room.
func (h *Handler) detachPeerConn(pc *rtpua.RtpUa) {
	log.Logger.Info("[worker] closing peer connection")
	gameRoom := h.getRoom(pc.RoomID)
	if gameRoom == nil || gameRoom.IsEmpty() {
		return
	}
	gameRoom.RemoveSession(pc)
	if gameRoom.IsEmpty() {
		log.Logger.Info("[worker] closing an empty room")
		gameRoom.Close()
		pc.InputChannel <- []byte{0xFF, 0xFF}
		close(pc.InputChannel)
	}
}
