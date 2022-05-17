package config

import (
	"common/monitor"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strconv"
	"time"
	"xmediaEmu/pkg/config/middleware"
	"xmediaEmu/pkg/config/service"

	"common/registry"
)


// TODO: 拆分yaml配置文件..
var AppConf struct {
	Http    service.Http
	Api     service.Api
	Ims     service.Ims
	Timeout struct {
		T1           int
		T2           int
		TEarlyMedia  int64
		TRecordMedia int64
	}
	Registry middleware.Registry

	Logger struct {
		LogLevel int
		LogPath  string
		LogDir   string
		LogFiles int
		Rtpdebug bool
		Advise   string
	}

	Ping struct {
		Timeout  int64
		HostPing string
	}

	Name        string
	Version     string
	HttpAddress string `yaml:"omitempty"`
}

func InitConfig(filepath string) {
	if "" == filepath {
		filepath = "setting.yaml"
	}
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err.Error())
	}
	err = yaml.Unmarshal(yamlFile, &AppConf)
	if err != nil {
		panic(err.Error())
	}

	if "" == AppConf.Http.Ip {
		AppConf.Http.Ip = "localhost"
	}
	AppConf.HttpAddress = AppConf.Http.Ip + ":" + strconv.Itoa(AppConf.Http.Port)

	if 0 == AppConf.Timeout.TEarlyMedia {
		AppConf.Timeout.TEarlyMedia = 10
	}
	AppConf.Timeout.TEarlyMedia = AppConf.Timeout.TEarlyMedia * int64(time.Second)

	// get version file.
	versionFile, err2 := ioutil.ReadFile(filepath)
	if err2 == nil {
		AppConf.Version = string(versionFile[:])
	}
}

// registry service ip and port to the ET-CD.
// address and port must be re-written.
var (
	Service = &registry.Service{
		Name: "go.micro.media-emulator",
		Metadata: map[string]string{
			"serverDescription": "media emulator service", // server desc.
		},
		Nodes: []*registry.Node{
			{
				Id:      "go.micro.media-emulator",
				Address: "localhost",
				Port:    8800,
				Metadata: map[string]string{
					"serverTag":           "media-emulator", // server division.
					monitor.ServiceStatus: monitor.DeleteState,
				},
			},
		},
		Version: "1",
	}
)

func IsNormal() bool {
	node := Service.Nodes[0]
	return node.Metadata[monitor.ServiceStatus] == monitor.NormalState
}

func IsPause() bool {
	node := Service.Nodes[0]
	return node.Metadata[monitor.ServiceStatus] == monitor.DeleteState
}

func SetStatus(status string) {
	Service.Nodes[0].Metadata[monitor.ServiceStatus] = status
}
