package libretro

import (
	"image"
	"net"
)

// GameFrame contains image and timeframe
type GameFrame struct {
	Image     *image.RGBA
	Timestamp uint32
}

// VideoExporter produces image frame to unix socket, h.264流.
type VideoExporter struct {
	sock         net.Conn
	imageChannel chan<- GameFrame
}
