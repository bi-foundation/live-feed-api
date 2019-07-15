package network

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/eventMessages"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"github.com/gogo/protobuf/proto"
	"net"
)

var (
	StandardChannelSize = 5000
)

type EventProxyServer struct {
	eventsInQueue chan eventMessages.Event
	RunState      runstate.RunState
	listener      net.Listener
}

func (ep *EventProxyServer) Init() *EventProxyServer {
	ep.eventsInQueue = make(chan eventMessages.Event, StandardChannelSize)
	return ep
}

func (ep *EventProxyServer) StartProxy() *EventProxyServer {
	go ep.tcpListener()
	return ep
}

func (ep *EventProxyServer) tcpListener() {
	ep.listener, _ = net.Listen("tcp", ":8040")

	for {
		conn, err := ep.listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go ep.handleConnection(conn)
	}
}

func (ep *EventProxyServer) handleConnection(conn net.Conn) {
	var bufferSize int32
	reader := bufio.NewReader(conn)

	for {
		binary.Read(reader, binary.LittleEndian, &bufferSize)
		data := make([]byte, bufferSize)
		bytesRead, err := reader.Read(data)
		if err != nil {
			fmt.Println("Could not read incoming data", err)
		}

		anchoredEvent := &eventMessages.AnchoredEvent{}
		err = proto.Unmarshal(data[0:bytesRead], anchoredEvent)

		fmt.Println("Received AnchoredEvent", anchoredEvent.DirectoryBlock.Header.DBHeight)
		ep.eventsInQueue <- anchoredEvent
	}
}

func (ep *EventProxyServer) Stop() {
	ep.listener.Close()
}
