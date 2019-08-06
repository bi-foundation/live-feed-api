package models

type EventType int

const (
	COMMIT_EVENT EventType = iota
	COMMIT_ENTRY
	ANCHOR_EVENT
	REVEAL_ENTRY
	NODE_MESSAGE
)

type Filter struct {
	EventType EventType
}
