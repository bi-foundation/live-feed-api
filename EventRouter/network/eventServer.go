package network

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/events/eventmessages"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"github.com/gogo/protobuf/proto"
	"log"
	"net"
)

var (
	StandardChannelSize = 5000
)

const (
	defaultConnectionHost     = "127.0.0.1"
	defaultConnectionPort     = "8040"
	defaultConnectionProtocol = "tcp"
)

type EventServer struct {
	eventsInQueue chan *eventmessages.FactomEvent
	state         runstate.RunState
	listener      net.Listener
	network       string
	address       string
}

func NewEventServer() EventServer {
	server := EventServer{
		eventsInQueue: make(chan *eventmessages.FactomEvent, StandardChannelSize),
		state:         runstate.New,
		network:       defaultConnectionProtocol,
		address:       fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort),
	}
	return server
}

func (ep *EventServer) Start() {
	go ep.listenIncomingConnections(ep.network, ep.address)
	ep.state = runstate.Running
}

func (ep *EventServer) Stop() {
	ep.state = runstate.Stopping
	err := ep.listener.Close()
	if err != nil {
		log.Fatalf("failed to close listener: %v", err)
	}
	ep.state = runstate.Stopped
}

func (ep *EventServer) listenIncomingConnections(network string, address string) {
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

func (ep *EventServer) handleConnection(conn net.Conn) error {
	defer closeConnection(conn)

	var bufferSize int32
	reader := bufio.NewReader(conn)

	for {
		// read an event from a client
		binary.Read(reader, binary.LittleEndian, &bufferSize)
		data := make([]byte, bufferSize)
		bytesRead, err := reader.Read(data)
		if err != nil {
			errorMsg := fmt.Sprintf("An error occurred while reading network data from remote address %v:, %v", getRemoteAddress(conn), err)
			if "EOF" == err.Error() {
				panic(errorMsg)
			} else {
				log.Println(errorMsg)
			}
		}

		factomEvent := &eventmessages.FactomEvent{}
		err = proto.Unmarshal(data[0:bytesRead], factomEvent)
		ep.eventsInQueue <- factomEvent
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

func (ep *EventServer) GetState() runstate.RunState {
	return ep.state
}

func (ep *EventServer) GetEventsInQueue() (chan *eventmessages.FactomEvent) {
	return ep.eventsInQueue
}

func (ep *EventServer) SetNetwork(network string) {
	ep.network = network
}

func (ep *EventServer) SetAddress(address string) {
	ep.address = address
}
