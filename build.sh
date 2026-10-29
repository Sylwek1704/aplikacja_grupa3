#!/bin/bash
set -e

cd "$(dirname "$0")"
export PATH=$PATH:/usr/local/go/bin

go mod tidy
go vet ./...

mkdir -p bin
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/czat .
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/czat.exe .

echo "zbudowano bin/czat i bin/czat.exe"
