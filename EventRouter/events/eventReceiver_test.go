package events

import (
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/live-feed-api/EventRouter/events/eventmessages"
	"github.com/gogo/protobuf/proto"
	"github.com/opsee/protobuf/opseeproto/types"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net"
	"sync/atomic"
	"testing"
	"time"
)

var eventsQueue chan *eventmessages.FactomEvent
var address string

func init() {
	// Start the new server at random port
	server := NewReceiver("tcp", ":0")
	server.Start()
	time.Sleep(10 * time.Millisecond) // sleep to allow the server to start before making a connection
	address = server.GetAddress()
	fmt.Printf("start server at: '%s'\n", address)
	eventsQueue = server.GetEventQueue()
}

func TestWriteEvents(t *testing.T) {
	n := 10
	data := mockData(t)
	dataSize := int32(len(data))

	conn := connect(t)
	defer conn.Close()

	for i := 1; i <= n; i++ {
		err := binary.Write(conn, binary.LittleEndian, dataSize)
		if err != nil {
			t.Fatal(err)
		}

		status, err := conn.Write(data)
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("written %d, %v", dataSize, status)
	}

	correctSendEvents := listenForEvents("WRITE", n, 20*time.Second)
	assert.EqualValues(t, n, correctSendEvents, "failed to receive the correct number of events %d != %d", n, correctSendEvents)
}

func TestEOFConnection(t *testing.T) {
	n := 10
	data := mockData(t)
	dataSize := int32(len(data))

	// test in parallel
	for i := 0; i < n; i++ {
		go func() {
			// prevent every thread making connection at the same time
			r := rand.Intn(10)
			time.Sleep(time.Duration(r) * time.Millisecond)

			conn := connect(t)
			defer conn.Close()

			// send one event correctly
			err := binary.Write(conn, binary.LittleEndian, dataSize)
			if err != nil {
				t.Fatal(err)
			}
			_, err = conn.Write(data)
			if err != nil {
				t.Fatal(err)
			}
		}()
	}

	correctSendEvents := listenForEvents("EOF", n, 20*time.Second)
	assert.EqualValues(t, n, correctSendEvents, "failed to receive the correct number of events %d != %d", n, correctSendEvents)
}

func connect(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func listenForEvents(testId string, n int, timeLimit time.Duration) int {
	var correctSendEvents int32 = 0
	quit := make(chan bool)
	go func() {
		for {
			select {
			case <-eventsQueue:
				atomic.AddInt32(&correctSendEvents, 1)
				fmt.Printf("[%s] > received event in queue\n", testId)
			case <-quit:
				return
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	deadline := time.Now().Add(timeLimit)
	for int(atomic.LoadInt32(&correctSendEvents)) != n && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
	quit <- true
	return int(correctSendEvents)
}

func mockData(t *testing.T) []byte {
	event := mockAnchorEvent()
	data, err := proto.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mockAnchorEvent() *eventmessages.FactomEvent {
	now := time.Now()
	testHash := []byte("12345678901234567890123456789012")
	return &eventmessages.FactomEvent{
		IdentityChainID: &eventmessages.Hash{
			HashValue: []byte("value"),
		},
		Value: &eventmessages.FactomEvent_BlockCommit{
			BlockCommit: &eventmessages.BlockCommit{
				DirectoryBlock: &eventmessages.DirectoryBlock{
					Header: &eventmessages.DirectoryBlockHeader{
						BodyMerkleRoot: &eventmessages.Hash{
							HashValue: testHash,
						},
						PreviousKeyMerkleRoot: &eventmessages.Hash{
							HashValue: testHash,
						},
						PreviousFullHash: &eventmessages.Hash{
							HashValue: testHash,
						},
						Timestamp:  &types.Timestamp{Seconds: int64(now.Second()), Nanos: int32(now.Nanosecond())},
						BlockCount: 456,
					},
					Entries: []*eventmessages.Entry{
						{
							ChainID: &eventmessages.Hash{
								HashValue: testHash,
							},
							KeyMerkleRoot: &eventmessages.Hash{
								HashValue: testHash,
							},
						}, {
							ChainID: &eventmessages.Hash{
								HashValue: testHash,
							},
							KeyMerkleRoot: &eventmessages.Hash{
								HashValue: testHash,
							},
						},
					},
				},
			},
		},
	}
}
