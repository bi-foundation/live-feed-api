package events

import (
	"live-api/EventRouter/events/eventmessages"
	"live-api/EventRouter/log"
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
		log.Debug("handle event: %v", factomEvent)
		switch factomEvent.Value.(type) {
		case *eventmessages.FactomEvent_AnchorEvent:
			log.Info("Received AnchoredEvent: %v", factomEvent.GetAnchorEvent())
		case *eventmessages.FactomEvent_CommitChain:
			log.Info("Received CommitChain: %v", factomEvent.GetCommitChain())
		case *eventmessages.FactomEvent_CommitEntry:
			log.Info("Received CommitEntry: %v", factomEvent.GetCommitEntry())
		case *eventmessages.FactomEvent_RevealEntry:
			log.Info("Received FactomEvent_RevealEntry: %v", factomEvent.GetRevealEntry())
		case *eventmessages.FactomEvent_NodeMessage:
			log.Info("Received FactomEvent_NodeMessage: %v", factomEvent.GetNodeMessage())
		}
	}
}
