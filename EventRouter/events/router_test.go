package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/live-api/EventRouter/events/eventmessages"
	"github.com/FactomProject/live-api/EventRouter/log"
	"github.com/FactomProject/live-api/EventRouter/models"
	"github.com/FactomProject/live-api/EventRouter/repository"
	"github.com/gogo/protobuf/types"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestEventRouter_Start(t *testing.T) {
	log.SetLevel(log.D)

	var eventsReceived int32 = 0
	event := mockAnchorEvent()

	queue := make(chan *eventmessages.FactomEvent)
	router := NewEventRouter(queue)
	router.Start()

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		defer atomic.AddInt32(&eventsReceived, 1)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var prettyJSON bytes.Buffer
		error := json.Indent(&prettyJSON, body, "", "\t")
		if error != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fmt.Printf("> %s\n", string(prettyJSON.Bytes()))
		// TODO assertion
		w.WriteHeader(http.StatusOK)
	})
	go http.ListenAndServe(":23232", nil)

	subscription := &models.Subscription{
		Callback: "http://localhost:23232/callback",
		Filters: map[models.EventType]models.Filter{
			// NO hardcode this
			models.COMMIT_EVENT: models.Filter{Filtering: ""},
		},
	}
	subscription, _ = repository.SubscriptionRepository.CreateSubscription(subscription)

	n := 1
	queue <- event

	//
	// TODO receive events on endpoints

	timeLimit := 1 * time.Minute
	deadline := time.Now().Add(timeLimit)
	for int(atomic.LoadInt32(&eventsReceived)) != n && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}
}

func mockAnchorEvent() *eventmessages.FactomEvent {
	now := time.Now()
	testHash := []byte("12345678901234567890123456789012")
	return &eventmessages.FactomEvent{
		EventSource: 0,
		Value: &eventmessages.FactomEvent_AnchorEvent{
			AnchorEvent: &eventmessages.AnchoredEvent{
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
