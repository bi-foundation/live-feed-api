package messages

type Event interface {
	Reset()
	String() string
	ProtoMessage()
}
