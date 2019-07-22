package events

import (
	"github.com/FactomProject/live-api/EventRouter/events/eventmessages"
	"log"
)

type EventRouter struct {
	eventsInQueue chan *eventmessages.FactomEvent
}

func NewEventRouter(queue chan *eventmessages.FactomEvent) EventRouter {
	return EventRouter{eventsInQueue: queue}
}

func (evr *EventRouter) Start() {
	go evr.handleEvents()
}

func (evr *EventRouter) handleEvents() {
	for factomEvent := range evr.eventsInQueue {
		switch factomEvent.Value.(type) {
		case *eventmessages.FactomEvent_AnchorEvent:
			log.Println("Received AnchoredEvent", factomEvent.GetAnchorEvent())
		case *eventmessages.FactomEvent_CommitChain:
			log.Println("Received CommitChain", factomEvent.GetCommitChain())
		case *eventmessages.FactomEvent_CommitEntry:
			log.Println("Received CommitEntry", factomEvent.GetCommitEntry())
		case *eventmessages.FactomEvent_RevealEntry:
			log.Println("Received FactomEvent_RevealEntry", factomEvent.GetRevealEntry())
		}
	}
}
