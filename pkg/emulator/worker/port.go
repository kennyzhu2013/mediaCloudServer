package worker

import (
	"errors"
	"sync"
)

func init() {
	defaultSuiteHelper = &PortSuiteHelper{
		startPort:     10000,
		capacity:      (20000 - 10000) / 2,
		portsUsed:     map[int]bool{},
		lastAllocated: -1,
	}
}

// module init here, called by external if not default.
func Init(start, end int) {
	defaultSuiteHelper.startPort = start
	defaultSuiteHelper.capacity = (end - start) / 2
}

// one port for audio or video, and for rtcp.
type PortSuite struct {
	PortRtp  int
	PortRtcp int
}

type PortSuiteHelper struct {
	startPort     int
	capacity      int
	lastAllocated int
	portsUsed     map[int]bool
	mu            sync.Mutex
}

var defaultSuiteHelper *PortSuiteHelper

func GetPortSuiteHelper() *PortSuiteHelper {
	return defaultSuiteHelper
}

// 4 ports for one session
func (psHelper *PortSuiteHelper) generatePortSuiteByID(id int) *PortSuite {
	base := psHelper.startPort + id*2
	return &PortSuite{
		PortRtp:  base,
		PortRtcp: base + 1,
	}
}

func (psHelper *PortSuiteHelper) AllotPort() (*PortSuite, error) {
	psHelper.mu.Lock()
	defer psHelper.mu.Unlock()

	if len(psHelper.portsUsed) >= psHelper.capacity {
		return nil, errors.New("insufficient capacity")
	}

	for {
		psHelper.lastAllocated = (psHelper.lastAllocated + 1) % psHelper.capacity
		if _, ok := psHelper.portsUsed[psHelper.lastAllocated]; !ok {
			break
		}
	}
	psHelper.portsUsed[psHelper.lastAllocated] = true
	return psHelper.generatePortSuiteByID(psHelper.lastAllocated), nil
}

func (psHelper *PortSuiteHelper) ReleasePort(portRtp int) {
	psHelper.mu.Lock()
	delete(psHelper.portsUsed, (portRtp-psHelper.startPort)/4)
	psHelper.mu.Unlock()
}
