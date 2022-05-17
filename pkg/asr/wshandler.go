package asr

import (
	"common/rtpengine"
	"github.com/gin-gonic/gin"
	"sync"
	"xmediaEmu/pkg/cws"
	"xmediaEmu/pkg/log"
)

// Client is a websocket client
type Handler struct {
	ID        string          // taskID
	// Socket    *websocket.Conn // xaudiobusiness
	// OutSocket *websocket.Conn // 出口的socket
	Socket    *cws.Client

	// xmedia方向进来的rtp流.
	xmediaTrack *rtpengine.TrackLocal

	OutSocket *cws.Client
	// exit      chan bool
	sync.Mutex

	// rtp
}

//
// WsAsrHandler负责处理asr相关,
func WsAsrHandler(c *gin.Context) {
	id  := c.GetHeader("id")
	log.Logger.Info("WsAsrHandler id:", id)
	handlers := &Handler{ID: id}

	// 处理逻辑.
}

// Asr建立连接请求.
func (h *Handler) handleInitAsr() cws.PacketHandler {

}