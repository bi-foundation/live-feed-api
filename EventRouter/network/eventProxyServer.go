package network

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/events/messages"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"github.com/gogo/protobuf/proto"
	"github.com/joomcode/errorx"
	"net"
)

var (
	StandardChannelSize = 5000
)

type EventProxyServer struct {
	eventsInQueue chan messages.Event
	RunState      runstate.RunState
	listener      net.Listener
}

func (ep *EventProxyServer) Init() *EventProxyServer {
	ep.eventsInQueue = make(chan messages.Event, StandardChannelSize)
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

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("Connection error: %v", r)
			fmt.Println(err)
			conn.Close()
		}
	}()

	var bufferSize int32
	reader := bufio.NewReader(conn)

	for {
		binary.Read(reader, binary.LittleEndian, &bufferSize)
		data := make([]byte, bufferSize)
		bytesRead, err := reader.Read(data)
		if err != nil {
			errx := errorx.Decorate(err, "An error occurred while reading network data from remote address %v", getRemoteAddress(conn))
			panic(errx)
		}

		anchoredEvent := &messages.AnchoredEvent{}
		err = proto.Unmarshal(data[0:bytesRead], anchoredEvent)

		fmt.Println("Received AnchoredEvent", anchoredEvent.DirectoryBlock.Header.DBHeight)
		ep.eventsInQueue <- anchoredEvent
	}
}

func (ep *EventProxyServer) Stop() {
	ep.listener.Close()
}

func getRemoteAddress(conn net.Conn) string {
	var addrString string
	remoteAddr := conn.RemoteAddr()
	if addr, ok := remoteAddr.(*net.TCPAddr); ok {
		addrString = addr.IP.String()
	} else {
		addrString = remoteAddr.String()
	}
	return addrString
}
