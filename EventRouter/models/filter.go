package models

// EventType
//
// A filtering is linked to event type: [CHAIN_REGISTRATION, ENTRY_REGISTRATION, BLOCK_COMMIT, ENTRY_CONTENT_REGISTRATION, NODE_MESSAGE]
// swagger:model EventType
type EventType string

const (
	CHAIN_REGISTRATION         EventType = "CHAIN_REGISTRATION"
	ENTRY_REGISTRATION         EventType = "ENTRY_REGISTRATION"
	ENTRY_CONTENT_REGISTRATION EventType = "ENTRY_CONTENT_REGISTRATION"
	BLOCK_COMMIT               EventType = "BLOCK_COMMIT"
	PROCESS_MESSAGE            EventType = "PROCESS_MESSAGE"
	NODE_MESSAGE               EventType = "NODE_MESSAGE"
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
