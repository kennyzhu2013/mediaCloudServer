package asr

import "xmediaEmu/pkg/cws/entity"

func (h *Handler) routes() {
	if h.Socket == nil {
		return
	}

	//
	h.Socket.Receive(entity.InitAsr, h.handleInitAsr())
}
