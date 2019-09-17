all: deps clean build test run
deps:
	export GO111MODULE=on;\
	go get -u ...;\
	export GO111MODULE=auto;\
build:
	go install;\
	cp live-feed.conf ${GOPATH}/bin
test:
	go test -v ./EventRouter/...
run:
	cd ${GOPATH}/bin;\
	./live-feed-api
clean:
	rm -f ${GOPATH}/bin/live-feed-api
