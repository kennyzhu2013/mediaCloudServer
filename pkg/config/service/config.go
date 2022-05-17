package service

type Http struct {
	Ip   string
	Port int

	WsUrl      string
	EarlyUrl   string
	EnableSwag bool
}

type Api struct {
	ClusterName string
	SrvName     string
	ServerId    string
}

// d对接媒体rtp的ip地址和端口.
type Ims struct {
	Ip    string
	Ports struct {
		Start int
		End   int
	}
}
