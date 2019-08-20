protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf \
--gofast_out=\
Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,\
Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. \
gen/eventmessages/generalTypes.proto \
gen/eventmessages/factoidBlock.proto \
gen/eventmessages/adminBlock.proto \
gen/eventmessages/entryCredit.proto \
gen/eventmessages/factomEvents.proto