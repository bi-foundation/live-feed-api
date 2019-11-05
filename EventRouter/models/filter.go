package models

// EventType of the subscribed filter
type EventType string

// Different event types
const (
	ChainCommit          EventType = "CHAIN_COMMIT"
	EntryCommit          EventType = "ENTRY_COMMIT"
	EntryReveal          EventType = "ENTRY_REVEAL"
	DirectoryBlockCommit EventType = "DIRECTORY_BLOCK_COMMIT"
	StateChange          EventType = "STATE_CHANGE"
	ProcessListEvent     EventType = "PROCESS_LIST_EVENT"
	NodeMessage          EventType = "NODE_MESSAGE"
)

// Filter for filtering an event type
type Filter struct {
	// Define a Filter on an EventType to filter the event. This allows to reduce the network traffic. The filtering is done with GraphQL
	Filtering string `json:"filtering" example:"{ identityChainID { hashValue } value { ... on NodeMessage { messageCode messageText } } }"`
}
