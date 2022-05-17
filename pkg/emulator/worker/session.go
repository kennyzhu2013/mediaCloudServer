package worker

import (
	"math/rand"
	"strconv"
	"strings"
	"xmediaEmu/pkg/emulator/rtpua"
)

type Session struct {
	ID             string // session id
	peerconnection *rtpua.RtpUa

	// Should I make direct reference
	room *Room
}

// Close close a session
func (s *Session) Close() {
	// TODO: Use event base
	s.peerconnection.StopClient()
}

const separator = "___"

// getGameNameFromRoomID parse roomID to get roomID and gameName
func GetGameNameFromRoomID(roomID string) string {
	parts := strings.Split(roomID, separator)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// generateRoomID generate a unique room ID containing 16 digits
func GenerateRoomID(gameName string) string {
	// RoomID contains random number + gameName
	// Next time when we only get roomID, we can launch game based on gameName
	roomID := strconv.FormatInt(rand.Int63(), 16) + separator + gameName
	return roomID
}
