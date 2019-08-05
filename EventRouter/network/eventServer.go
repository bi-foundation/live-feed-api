package network

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/events/eventmessages"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/common/constants/runstate"
	"github.com/gogo/protobuf/proto"
	"io"
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

type EventServer interface {
	Start()
	Stop()
	GetState() runstate.RunState
	GetEventQueue() chan *eventmessages.FactomEvent
	GetAddress() string
}

type Server struct {
	eventQueue chan *eventmessages.FactomEvent
	state      runstate.RunState
	listener   net.Listener
	protocol   string
	address    string
}

func NewServer(protocol string, address string) EventServer {
	return &Server{
		eventQueue: make(chan *eventmessages.FactomEvent, StandardChannelSize),
		state:      runstate.New,
		protocol:   protocol,
		address:    address,
	}
}

func NewDefaultServer() EventServer {
	return NewServer(defaultConnectionProtocol, fmt.Sprintf("%s:%s", defaultConnectionHost, defaultConnectionPort))
}

func (server *Server) Start() {
	go server.listenIncomingConnections()
	server.state = runstate.Running
}

func (server *Server) Stop() {
	server.state = runstate.Stopping
	err := server.listener.Close()
	if err != nil {
		log.Error("failed to close listener: %v", err)
	}
	server.state = runstate.Stopped
}

func (server *Server) listenIncomingConnections() {
	listener, err := net.Listen(server.protocol, server.address)
	log.Info(" event server listening: '%s' at %s", server.protocol, server.address)
	if err != nil {
		log.Error("failed to listen to %s on %s: %v", server.protocol, server.address, err)
		return
	}
	server.listener = listener

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Error("connection from factomd failed: %v", err)
		}

		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer finalizeConnection(conn)
	if err := server.readEvents(conn); err != nil {
		log.Error("failed to read events: %v", err)
	}
}

func (server *Server) readEvents(conn net.Conn) (err error) {
	log.Debug("read events from: %s", getRemoteAddress(conn))

	var dataSize int32
	reader := bufio.NewReader(conn)

	// continuously read the stream of events from connection
	for {
		// read the size of the factom event
		err = binary.Read(reader, binary.LittleEndian, &dataSize)
		if err != nil {
			return fmt.Errorf("failed to data size from %s:, %v", getRemoteAddress(conn), err)
		}

		// read the factom event
		data := make([]byte, dataSize)
		bytesRead, err := io.ReadFull(reader, data)
		if err != nil {
			return fmt.Errorf("failed to data from %s:, %v", getRemoteAddress(conn), err)
		}

		factomEvent := &eventmessages.FactomEvent{}
		err = proto.Unmarshal(data[0:bytesRead], factomEvent)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event from %s: %v", getRemoteAddress(conn), err)
		}
		log.Debug("read factom event... %v", factomEvent)
		server.eventQueue <- factomEvent
	}
}

func finalizeConnection(conn net.Conn) {
	log.Info("connection closed unexpectedly to: %s", getRemoteAddress(conn))
	if r := recover(); r != nil {
		log.Error("recovered during handling connection: %v\n", r)
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

func (server *Server) GetState() runstate.RunState {
	return server.state
}

func (server *Server) GetAddress() string {
	if server.listener == nil {
		return server.address
	}
	return server.listener.Addr().String()
}

func (server *Server) GetEventQueue() chan *eventmessages.FactomEvent {
	return server.eventQueue
}
