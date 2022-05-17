package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"xmediaEmu/pkg/emulator/config"
	"xmediaEmu/pkg/emulator/worker"
	"xmediaEmu/pkg/log"
)

var (
	upgrader = &websocket.Upgrader{} // use default options
)

//func init() {
//	upgrader = &websocket.Upgrader{
//		HandshakeTimeout: 15 * time.Second,
//		CheckOrigin: func(r *http.Request) bool {
//			if r.Method != http.MethodGet {
//				return false
//			}
//			if r.URL.Path != "/room" {
//				return false
//			}
//			return true
//		},
//	}
//}

// work as ws server. work with coordinator
func main() {
	log.Init("1", log.Config{Level: int32(0), Dir: "./", Path: "worker.log", FileNum: 10})
	http.HandleFunc("/room", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		go startWsHandle(conn)
	})

	_ = http.ListenAndServe(":9999", nil)
}

//func startMonitor(conn *websocket.Conn) {
//	defer conn.Close()
//
//	for {
//		_ = conn.SetReadDeadline(time.Now().Add(20 * time.Second))
//		messageType, message, err := conn.ReadMessage()
//		if err != nil {
//			fmt.Printf("Read message failed: %v\n", err)
//			return
//		}
//		err = conn.WriteMessage(messageType, message)
//		if err != nil {
//			fmt.Printf("Write message failed: %v\n", err)
//			return
//		}
//		fmt.Printf("The message is: %v\n", string(message))
//	}
//}
//
func startWsHandle(conn *websocket.Conn) {
	defer conn.Close()
	videoconfig := config.VideoConfig{Codec: "h264"}
	videoconfig.H264.Profile = "baseline"
	videoconfig.H264.Tune = "stillimage"
	videoconfig.H264.Crf = 20
	videoconfig.H264.Preset = "fast"
	encoder := config.EncoderConfig{Audio: config.AudioConfig{Codec: "g711", Channels: 1, Frame: 20, Frequency: 8000}, Video: videoconfig, BUseUnixSocket: false}
	mainHandler := worker.NewHandler(config.Config{Encoder: encoder, Width: 640, Height: 480, LocalMediaIp: "10.153.90.4"}, conn) // 媒体层统一ip.
	mainHandler.Run("testWorkServer")
}
