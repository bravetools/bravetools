VERSION := $(shell cat ./VERSION)
BINARY_NAME := brave
GO_LDFLAGS="-s -X github.com/beringresearch/bravetools/shared.braveVersion=$(VERSION)"

ubuntu:
	@echo "Building Bravetools ..."
	go clean
	go get
	rm -f install/ubuntu/$(BINARY_NAME)
	@GOOS=linux go build -ldflags=$(GO_LDFLAGS) -o install/ubuntu/$(BINARY_NAME) *.go
	@echo "Installing ..."
	sudo cp install/ubuntu/brave /usr/bin/
	@echo "version: "$(VERSION)
	@echo "Bravetools installed"

linux:
	@echo "Building Bravetools ..."
	go clean
	go get
	rm -f install/linux/$(BINARY_NAME)
	@GOOS=linux go build -ldflags=$(GO_LDFLAGS) -o install/linux/$(BINARY_NAME) *.go
	@echo "Installing ..."
	sudo cp install/linux/brave /usr/bin/
	@echo "version: "$(VERSION)
	@echo "Bravetools installed"

darwin:
	@echo "Building Bravetools ..."
	go clean
	go get
	rm -f install/darwin/$(BINARY_NAME)
	@GOOS=darwin go build -ldflags=$(GO_LDFLAGS) -o install/darwin/$(BINARY_NAME) *.go
	@echo "Installing ..."
	sudo cp -f install/darwin/brave /usr/local/bin/
	@echo "version: "$(VERSION)
	@echo "Bravetools installed"

deps:
	go get -u