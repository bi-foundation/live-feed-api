package network

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/events/messages"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"github.com/gogo/protobuf/proto"
	"log"
	"net"
)

var (
	StandardChannelSize = 5000
)

const (
	defaultConnectionHost     = ""
	defaultConnectionPort     = "3333"
	defaultConnectionProtocol = "tcp"
)

type eventServer struct {
	eventsInQueue chan messages.Event
	state         runstate.RunState
	listener      net.Listener
	network       string
	address       string
}

func New() eventServer {
	server := eventServer{
		eventsInQueue: make(chan messages.Event, StandardChannelSize),
		state:         runstate.New,
		network:       defaultConnectionProtocol,
		address:       fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort),
	}
	return server
}

func (ep *eventServer) Start() {
	go ep.listenIncomingConnections(ep.network, ep.address)
	ep.state = runstate.Running
}

func (ep *eventServer) Stop() {
	ep.state = runstate.Stopping
	err := ep.listener.Close()
	if err != nil {
		log.Fatalf("failed to close listener: %v", err)
	}
	ep.state = runstate.Stopped
}

func (ep *eventServer) listenIncomingConnections(network string, address string) {
	var err error
	ep.listener, err = net.Listen(network, address)
	if err != nil {
		log.Fatalf("failed to listen to %s on %s: %v", network, address, err)
	}

	for {
		conn, err := ep.listener.Accept()
		if err != nil {
			log.Printf("failed to connect to factomd: %v", err)
		}
		go ep.handleConnection(conn)
	}
}

func (ep *eventServer) handleConnection(conn net.Conn) error {
	defer closeConnection(conn)

	var bufferSize int32
	reader := bufio.NewReader(conn)

	for {
		// read an event from a client
		binary.Read(reader, binary.LittleEndian, &bufferSize)
		data := make([]byte, bufferSize)
		bytesRead, err := reader.Read(data)
		if err != nil {
			log.Printf("An error occurred while reading network data from remote address %v:, %v", getRemoteAddress(conn), err)
		}

		anchoredEvent := &messages.AnchoredEvent{}
		err = proto.Unmarshal(data[0:bytesRead], anchoredEvent)

		log.Println("Received AnchoredEvent", anchoredEvent.DirectoryBlock.Header.DBHeight)
		ep.eventsInQueue <- anchoredEvent
	}
}

func closeConnection(conn net.Conn) {
	if r := recover(); r != nil {
		log.Printf("Connection error: %v\n", r)
	}
	conn.Close()
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

func (ep *eventServer) GetState() runstate.RunState {
	return ep.state
}

func (ep *eventServer) SetNetwork(network string) {
	ep.network = network
}

func (ep *eventServer) SetAddress(address string) {
	ep.address = address
}
