package app

import (
	"common/monitor"
	"common/util/process"
	"github.com/pterm/pterm"
	"os"
	"syscall"
	"time"
	"xmediaEmu/pkg/metric"
	"xmediaEmu/pkg/util"

	. "xmediaEmu/pkg/config"
	"xmediaEmu/pkg/log"
)

// init rtpProxy func.
func InitService() {
	InitConfig("conf/setting.conf")

	service := Service
	service.Name = AppConf.Api.SrvName
	var nodeSelf = service.Nodes[0]
	nodeSelf.Address = AppConf.Http.Ip
	nodeSelf.Port = AppConf.Http.Port

	// write id to metadata as server-id tag.
	if len(AppConf.Api.ServerId) != 0 {
		nodeSelf.Metadata["serverId"] = AppConf.Api.ServerId
	}
	nodeSelf.Id = AppConf.HttpAddress

	// init log here.
	log.Init(nodeSelf.Id, log.Config{ int32(AppConf.Logger.LogLevel), AppConf.Logger.LogDir, AppConf.Logger.LogPath, AppConf.Logger.LogFiles })
	log.Logger.Infof("logger init, log name:%v, LogLevel:%v, Timeout.TRecordMedia:%v", AppConf.Logger.LogPath, AppConf.Logger.LogLevel, AppConf.Timeout.TRecordMedia)


	// init recover handler, 远程告警注册, 远程告警服务器不支持微服务.
	process.InitPanicReport(AppConf.Api.SrvName, AppConf.HttpAddress, func(req *process.PanicReq) interface{} {
		code, _, _ := util.Post(AppConf.Logger.Advise, map[string]interface{}{
			"alerts": req,
			"email":  "panic",
		})
		return code
	})

	// init metric here
	metric.InitMetrics(nodeSelf.Id, service.Version)


	router.Init(service.Version)
	monitor.SetMaxCalls((AppConf.Ims.Ports.End - AppConf.Ims.Ports.Start) / 2)

	// set core unlimited
	var ulimitSize uint64
	ulimitSize = 0x7FFFFFFFFF
	rlimit := &syscall.Rlimit{ulimitSize, ulimitSize}
	if err := syscall.Setrlimit(syscall.RLIMIT_CORE, rlimit); err != nil {
		log.Logger.Errorf("MediaServer StartFail, syscall.Setrlimit failed:%v", err)
		pterm.FgLightRed.Printf("MediaServer StartFail, syscall.Setrlimit failed:%s\n", err.Error())
		os.Exit(-1)
	}

	if err := syscall.Getrlimit(syscall.RLIMIT_CORE, rlimit); err != nil {
		log.Logger.Errorf("MediaServer StartFail, syscall.Getrlimit failed:%v", err)
		pterm.FgLightRed.Printf("MediaServer StartFail, syscall.Getrlimit failed:%s\n", err.Error())
		os.Exit(-1)
	}
	log.Logger.Infof("After set rlimit CORE dump current is:%v, max is:%v", rlimit.Cur, rlimit.Max)
	pterm.FgYellow.Printf("After set rlimit CORE dump current is:%d, max is:%d\n", rlimit.Cur, rlimit.Max)

	// open files
	rlimit.Cur = 0x7FFFF
	rlimit.Max = 0x7FFFF
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, rlimit); err != nil {
		log.Logger.Errorf("XMedia StartFail, syscall.Setrlimit failed:%v", err)
		os.Exit(-1)
	}
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, rlimit); err != nil {
		log.Logger.Errorf("XMedia StartFail, syscall.Getrlimit failed:%v", err)
		os.Exit(-1)
	}
	log.Logger.Infof("After set rlimit RLIMIT_NOFILE current is:%v, max is:%v", rlimit.Cur, rlimit.Max)
	pterm.FgYellow.Printf("After set rlimit RLIMIT_NOFILE current is:%d, max is:%d\n", rlimit.Cur, rlimit.Max)

	tempEnv := os.Getenv("GOTRACEBACK")
	log.Logger.Infof("os.GetEnv GOTRACEBACK is:%v", tempEnv)
}

// modify service info to servers..
// not used any more
func StartMonitor(exit chan bool) {
	t := time.NewTicker(monitor.HeartBeatCheck)
	sessionManage, _ := rtpproxy.GetSessionMgr()
	defer func() {
		// here must move to the record file function
		if v := recover(); v != nil {
			process.DefaultPanicReport.RecoverFromPanic("", "startMonitor", v)
		}
	}()

	// to add ping prometheus monitor goroutines.
	for {
		select {
		case <-t.C:
			if IsPause() {
				continue
			}
			monitor.UpdateServiceWithoutSync(Service, sessionManage.Size())
		case <-exit:
			t.Stop()

			return
		}
	}
}

