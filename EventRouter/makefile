generate:
	protoc \
	    --gogoopsee_out=plugins=grpc+graphql+bitflags,Mgoogle/protobuf/timestamp.proto=github.com/bi-foundation/protobuf-graphql-extension/graphqlproto/typesre,Mopsee/protobuf/opsee.proto=github.com/opsee/protobuf/opseeproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./eventmessages/generated \
	    --proto_path=$(GOPATH)/src:. eventmessages/*.proto
clean:
	rm -r -f eventmessages/generated/*

.PHONY:
	clean
	generate
