GO111MODULEBACKUP:=$(GO111MODULE)

generate:
# get protobuf-graphql-extension with modules turned off as protoc doesn't like @v1.0.0 tags in the include path
	GO111MODULEBACKUP=${GO111MODULE}; \
	export GO111MODULE=off; \
	go get -u github.com/bi-foundation/protobuf-graphql-extension; \
	protoc \
	    --gogoopsee_out=plugins=grpc+graphql+bitflags,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./eventmessages/generated \
		--proto_path=$(GOPATH)/src:. eventmessages/*.proto; \
	export GO111MODULE=${GO111MODULEBACKUP};
clean:
	rm -r -f eventmessages/generated/*
setup:
	cd /tmp; \
	wget https://github.com/bi-foundation/protobuf-graphql-extension/releases/download/v1.0.19/protobuf-graphql-extension_1.0.19_Linux_x86_64.tar.gz; \
	tar -xf protobuf-graphql-extension_1.0.19_Linux_x86_64.tar.gz; \
	mv ./protoc-gen-gogoopsee $(GOPATH)/bin; \

.PHONY:
	clean
	generate
