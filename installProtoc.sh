cd /tmp
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.9.0/protoc-3.9.0-linux-x86_64.zip
unzip protoc-3.9.0-linux-x86_64.zip.1 -d protoc3
sudo mv protoc3/bin/* /usr/local/bin/
sudo mv protoc3/include/* /usr/local/include/
sudo chown $USER:root /usr/local/bin/proto
sudo chown $USER:root /usr/local/include/google/
go get -u github.com/golang/protobuf/protoc-gen-go
go get github.com/gogo/protobuf/protoc-gen-gofast
go get github.com/gogo/protobuf/protoc-gen-gogofast
