package models

// EventType
//
// A filtering is linked to event type: [COMMIT_CHAIN, COMMIT_ENTRY, ANCHOR_EVENT, REVEAL_ENTRY, NODE_MESSAGE]
// swagger:model EventType
type EventType string

const (
	COMMIT_CHAIN    EventType = "COMMIT_CHAIN"
	COMMIT_ENTRY    EventType = "COMMIT_ENTRY"
	ANCHOR_EVENT    EventType = "ANCHOR_EVENT"
	REVEAL_ENTRY    EventType = "REVEAL_ENTRY"
	PROCESS_MESSAGE EventType = "PROCESS_MESSAGE"
	NODE_MESSAGE    EventType = "NODE_MESSAGE"
)

// Filter
//
// Define a Filter on an EventType to filter the event. This allows to reduce the network traffic.
// swagger:model Filter
type Filter struct {
	// Filtering with GraphQL
	// required: false
	Filtering string `json:"filtering"`
}
