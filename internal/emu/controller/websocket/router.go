package websocket

import (
	ginPlugin "common/util/app/gin-plugin"
	"xmediaEmu/pkg/asr"

	"github.com/gin-gonic/gin"
	"net/http"
	"xmediaEmu/internal/emu/controller/websocket/ops"
)

var Handlers struct {
	Router        *gin.Engine
	MetricHandler http.Handler
	ops.Ops
}

// for health beat..
func NoModules(ctx *gin.Context) {
	ctx.String(200, "Server Started! ")
}

// self init...
func init() {
	gin.SetMode(gin.ReleaseMode)
	Handlers.Router = gin.New()
	Handlers.Router.Use(ginPlugin.AccessLog())
	Handlers.Router.Use(ginPlugin.Recovery())
}

func Init(ver string) {
	// micro health go.micro.api.gin call this function.
	Handlers.Router.POST("/", NoModules)
	Handlers.Router.GET("/", NoModules)
	Handlers.Version = ver

	// 保留xmedia的运维基本命令.
	opsGroup := Handlers.Router.Group("/ops")
	{
		opsGroup.GET("/log", ops.LogHandle)
	}

	// Asr命令.
	Handlers.Router.POST("/asr", asr.WsAsrHandler)

	// TODO: emulator命令.

}