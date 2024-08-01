
# Parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=poll-service

# Targets
all: clean deps test build_linux

deps:
	$(GOMOD) tidy

build: 
	$(GOBUILD) -o $(BINARY_NAME) -ldflags "-s -w" ./cmd/poll-service

test:
	$(GOTEST) -v ./...

build_linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)_linux_amd64 -ldflags "-s -w" ./cmd/poll-service

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)_linux_amd64

up:
	docker-compose up --build --force-recreate -d
