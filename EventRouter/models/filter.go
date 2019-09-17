package models

type EventType string

const (
	CHAIN_REGISTRATION         EventType = "CHAIN_REGISTRATION"
	ENTRY_REGISTRATION         EventType = "ENTRY_REGISTRATION"
	ENTRY_CONTENT_REGISTRATION EventType = "ENTRY_CONTENT_REGISTRATION"
	BLOCK_COMMIT               EventType = "BLOCK_COMMIT"
	PROCESS_MESSAGE            EventType = "PROCESS_MESSAGE"
	NODE_MESSAGE               EventType = "NODE_MESSAGE"
)

type Filter struct {
	// Define a Filter on an EventType to filter the event. This allows to reduce the network traffic. The filtering is done with GraphQL
	Filtering string `json:"filtering" example:"{ identityChainID { hashValue } value { ... on NodeMessage { messageCode messageText } } }"`
}
