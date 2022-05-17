package worker

import "xmediaEmu/pkg/cws/entity"

// 命令字路由.
func (h *Handler) routes() {
	if h.oClient == nil {
		return
	}

	// TODO: Start带对端地址启动，stop停止完成两个基本功能.
	h.oClient.Receive(entity.RoomStart, h.handleRoomStart())
	h.oClient.Receive(entity.RoomQuit, h.handleRoomQuit())
}
