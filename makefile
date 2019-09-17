all: clean build test run
build:
	export GO111MODULE=on;\
	go get -u ...;\
	go install;\
	export GO111MODULE=auto;\
	cp live-feed.conf ${GOPATH}/bin
test:
	go test -v ./EventRouter/...
run:
	cd ${GOPATH}/bin;\
	./live-feed-api
clean:
	rm -f ${GOPATH}/bin/live-feed-api
