package events

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/config"
	"github.com/FactomProject/live-feed-api/EventRouter/eventmessages/generated/eventmessages"
	"github.com/FactomProject/live-feed-api/EventRouter/log"
	"github.com/FactomProject/live-feed-api/EventRouter/models"
	"github.com/gogo/protobuf/proto"
	"io"
	"net"
)

const (
	defaultStandardChannelSize = 5000
	supportedProtocolVersion   = byte(1)
)

// EventReceiver responsible to receive events from factomd
type EventReceiver interface {
	Start()
	Stop()
	GetState() models.RunState
	GetEventQueue() chan *eventmessages.FactomEvent
	GetAddress() string
}

type receiver struct {
	eventQueue chan *eventmessages.FactomEvent
	state      models.RunState
	listener   net.Listener
	protocol   string
	address    string
}

// NewReceiver creates a new receiver
func NewReceiver(eventListenerConfig *config.ReceiverConfig) EventReceiver {
	return &receiver{
		eventQueue: make(chan *eventmessages.FactomEvent, defaultStandardChannelSize),
		state:      models.New,
		protocol:   eventListenerConfig.Protocol,
		address:    fmt.Sprintf("%s:%d", eventListenerConfig.BindAddress, eventListenerConfig.Port),
	}
}

// Start the receiver with listening
func (receiver *receiver) Start() {
	go receiver.listenIncomingConnections()
	receiver.state = models.Running
}

// Stop the receiver
func (receiver *receiver) Stop() {
	receiver.state = models.Stopping
	err := receiver.listener.Close()
	if err != nil {
		log.Error("failed to close listener: %v", err)
	}
	receiver.state = models.Stopped
}

func (receiver *receiver) listenIncomingConnections() {
	listener, err := net.Listen(receiver.protocol, receiver.address)
	log.Info("start event receiver at: '%s' at %s", receiver.protocol, receiver.address)
	if err != nil {
		log.Error("failed to listen to %s on %s: %v", receiver.protocol, receiver.address, err)
		return
	}
	receiver.listener = listener

	for {
		conn, err := receiver.listener.Accept()
		if err != nil {
			log.Error("connection from factomd failed: %v", err)
			continue
		}

		go receiver.handleConnection(conn)
	}
}

func (receiver *receiver) handleConnection(conn net.Conn) {
	defer finalizeConnection(conn)
	if err := receiver.readEvents(conn); err != nil {
		log.Error("failed to read events: %v", err)
	}
}

func (receiver *receiver) readEvents(conn net.Conn) (err error) {
	log.Debug("read events from: %s", getRemoteAddress(conn))

	var dataSize int32
	reader := bufio.NewReader(conn)

	// continuously read the stream of events from connection
	for {
		// Read the protocol version, return an error on mismatch
		protocolVersion, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to protocol version from %s:, %v", getRemoteAddress(conn), err)
		}
		if protocolVersion != supportedProtocolVersion {
			return fmt.Errorf("invalid protocol version from %s:, the received version is %d while the supported version is %d",
				getRemoteAddress(conn), protocolVersion, supportedProtocolVersion)
		}

		// read the size of the factom event
		err = binary.Read(reader, binary.LittleEndian, &dataSize)
		if err != nil {
			if err == io.EOF {
				log.Warn("the client at %s disconnected", getRemoteAddress(conn))
				return nil
			}
			return fmt.Errorf("failed to data size from %s:, %v", getRemoteAddress(conn), err)
		}

		// read the factom event
		data := make([]byte, dataSize)
		bytesRead, err := io.ReadFull(reader, data)
		if err != nil {
			if err == io.EOF {
				log.Warn("the client at %s disconnected", getRemoteAddress(conn))
				return nil
			}
			return fmt.Errorf("failed to data from %s:, %v", getRemoteAddress(conn), err)
		}

		factomEvent := &eventmessages.FactomEvent{}
		err = proto.Unmarshal(data[0:bytesRead], factomEvent)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event from %s: %v", getRemoteAddress(conn), err)
		}
		log.Debug("read factom event... %v", factomEvent)
		receiver.eventQueue <- factomEvent
	}
}

func finalizeConnection(conn net.Conn) {
	log.Info("connection from %s closed unexpectedly", getRemoteAddress(conn))
	if r := recover(); r != nil {
		log.Error("recovered during handling connection: %v\n", r)
	}
	_ = conn.Close()
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

// GetState to get the state of the receiver
func (receiver *receiver) GetState() models.RunState {
	return receiver.state
}

// GetAddress to get the address where the receiver is listening to
func (receiver *receiver) GetAddress() string {
	if receiver.listener == nil {
		return receiver.address
	}
	return receiver.listener.Addr().String()
}

// GetEventQueue to get queue of new events
func (receiver *receiver) GetEventQueue() chan *eventmessages.FactomEvent {
	return receiver.eventQueue
}
