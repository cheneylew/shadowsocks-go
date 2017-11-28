package main

import (
	ss "github.com/cheneylew/shadowsocks-go/shadowsocks"
	"sync"
)

var PORT_PASSWORD map[string]string
var DURATION_QUERY_PORT_PASSWORD int64 = 5
var DURATION_UPLOAD_FLOW int64 = 7
var SS_DEBUG ss.DebugLog = true
var SS_FlowCounterManager FlowCounterManager = FlowCounterManager{flowCounter:map[string]*FlowCounter{}}

func init() {
	PORT_PASSWORD = make(map[string]string)
	PORT_PASSWORD["10004"] = "11111111"

}

type FlowCounterManager struct {
	sync.Mutex
	flowCounter map[string]*FlowCounter
}

func (pm *FlowCounterManager) add(in, out uint64, port string) {
	flow, ok := pm.get(port)

	pm.Lock()
	if ok {
		flow.In += in
		flow.Out += out
	} else {
		pm.flowCounter[port] = &FlowCounter{In:in, Out:out, Port:port}
	}
	pm.Unlock()
}

func (pm *FlowCounterManager) get(port string) (fc *FlowCounter, ok bool) {
	pm.Lock()
	fc, ok = pm.flowCounter[port]
	pm.Unlock()
	return fc, ok
}

func (pm *FlowCounterManager) update(in, out uint64, port string) {
	flow, ok := pm.get(port)
	pm.Lock()
	if ok {
		flow.In = in
		flow.Out = out
	} else {
		pm.flowCounter[port] = &FlowCounter{In:in, Out:out, Port:port}
	}
	pm.Unlock()
}

func (pm *FlowCounterManager) del(port string) {
	_, ok := pm.get(port)
	if !ok {
		return
	}
	pm.Lock()
	delete(pm.flowCounter, port)
	pm.Unlock()
}
