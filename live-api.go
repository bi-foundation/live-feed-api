package main

import (
	"github.com/FactomProject/live-api/EventRouter/network"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"time"
)

var (
	eventProxyServer *network.EventProxyServer = &network.EventProxyServer{}
)

func main() {

	eventProxyServer.Init()
	eventProxyServer.StartProxy()

	for eventProxyServer.RunState < runstate.Stopping {
		time.Sleep(time.Second)
	}

	eventProxyServer.Stop()
}
