package ops

import (
	"context"
	"github.com/gin-gonic/gin"
	"strconv"
	"xmediaEmu/pkg/config"
	"xmediaEmu/pkg/log"
)

type Ops struct {
	// for monitor service commands.
	ServiceCancel context.CancelFunc

	// IsPause       bool
	Version string // for handlers.
}

// @tags 系统操作
// @Summary 日志级别更新
// @Produce json
// @Param level query int true "日志级别，0：debug、1：info"
// @Success 200
// @Router /ops/log [get]
func LogHandle(ctx *gin.Context) {
	levelString := ctx.Query("level")
	level, _ := strconv.Atoi(levelString)
	if level > 0 {
		config.AppConf.Logger.Rtpdebug = false
	} else {
		config.AppConf.Logger.Rtpdebug = true
	}

	log.SetLevel(level)
	log.Logger.Infof("log level change to:%d", level)
}