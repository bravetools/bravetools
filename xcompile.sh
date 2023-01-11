#!/bin/bash
set -e
# you may need to go install github.com/mitchellh/gox@v1.0.1 first

GOFLAGS="-trimpath" CGO_ENABLED=0 gox -ldflags "-s -w" -output="bin/brave_{{.OS}}_{{.Arch}}" --osarch="darwin/amd64 darwin/arm64 linux/amd64 linux/arm linux/arm64 windows/amd64"
