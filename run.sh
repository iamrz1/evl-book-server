#! /bin/sh
echo "building server ..."
export GO111MODULE=on
go mod vendor
CGO_ENABLED=0 GOFLAGS=-mod=vendor go build -o server

echo "server running ..."

./server serve
