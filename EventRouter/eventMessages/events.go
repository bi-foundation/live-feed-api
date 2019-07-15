package eventMessages

import (
	"github.com/golang/protobuf/proto"
)

type Event interface {
	Reset()
	String() string
	ProtoMessage()
	XXX_Unmarshal(b []byte) error
	XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
	XXX_Merge(src proto.Message)
	XXX_Size() int
	XXX_DiscardUnknown()
}
