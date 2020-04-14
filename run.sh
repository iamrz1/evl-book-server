#! /bin/sh
#docker-compose up -d

echo "building server ..."
export GO111MODULE=on
go mod vendor
CGO_ENABLED=0 GOFLAGS=-mod=vendor go build -o server

echo "building running ..."

./server serve
