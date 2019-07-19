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
		case *eventmessages.FactomEvent_AnchoredEvent:
			log.Println("Received AnchoredEvent", factomEvent.GetAnchoredEvent())
		case *eventmessages.FactomEvent_IntermediateEvent:
			log.Println("Received IntermediateEvent", factomEvent.GetIntermediateEvent())
		}
	}
}
