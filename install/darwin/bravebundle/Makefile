all:

build:
	go get
	go clean
	rm -f bravebundle
	@GOOS=darwin go build -ldflags=-s -o bravebundle *.go
	@echo "building bravebundle done"
