GOCMD:=$(shell which go)
GOLINT:=$(shell which golint)
GOIMPORT:=$(shell which goimports)
GOFMT:=$(shell which gofmt)
GOBUILD:=$(GOCMD) build
GOINSTALL:=$(GOCMD) install
GOCLEAN:=$(GOCMD) clean
GOTEST:=$(GOCMD) test
GOGET:=$(GOCMD) get
GOLIST:=$(GOCMD) list
GOVET:=$(GOCMD) vet
u := $(if $(update),-u)

PACKAGES:=$(shell $(GOLIST) ./EventRouter/...)
GOFILES:=$(shell find . -path ./vendor -prune -o -name "*.go" -type f)

GOBIN:=$(GOPATH)/bin
export PATH := $(GOBIN):$(PATH)
export GO111MODULEENV := on
export GO111MODULE := on

all: clean build test

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME);

.PHONY: deps
deps:
	$(GOGET) ${u} -d

.PHONY: build
build: deps
	$(GOBUILD) ./factom-live-feed-api.go;

.PHONY: install
install: deps
	$(GOINSTALL) ./factom-live-feed-api.go;

.PHONY: test
test:
	echo "mode: count" > coverage.out
	for PKG in $(PACKAGES); do \
		$(GOCMD) test -v -covermode=count -coverprofile=profile.out $$PKG > tmp.out; \
		cat tmp.out; \
		if grep -q "^--- FAIL" tmp.out; then \
			rm -f tmp.out; \
			exit 1; \
		elif grep -q "build failed" tmp.out; then \
			rm -f tmp.out; \
			exit; \
		fi; \
		if [ -f profile.out ]; then \
			cat profile.out | grep -v "mode:" >> coverage.out; \
			rm profile.out; \
		fi; \
	done; \
	rm -f tmp.out;

.PHONY: run
run:
	$(GOCMD) run ./factom-live-feed-api.go

.PHONY: generate
generate:
	$(GOCMD) generate;

.PHONY: dev-deps
dev-deps:
	GO111MODULE=off $(GOGET) -v ${u} \
		golang.org/x/lint/golint \
		github.com/swaggo/swag/cmd/swag	\
		github.com/swaggo/swag/gen	\
		github.com/golang/protobuf/protoc-gen-go	\
		github.com/bi-foundation/protobuf-graphql-extension/protoc-gen-gogoopsee

.PHONY: lint
lint: dev-deps
	for PKG in $(PACKAGES); do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: vet
vet: deps dev-deps
	$(GOVET) $(PACKAGES)

.PHONY: fmt
fmt:
	$(GOFMT) -s -w $(GOFILES)

.PHONY: fmt-check
fmt-check:
	@diff=$$($(GOFMT) -s -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

.PHONY: view-covered
view-covered: test
	$(GOCMD) tool cover -html=coverage.out
