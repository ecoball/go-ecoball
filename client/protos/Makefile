# Makefile for protobuf
default: all

PB = $(wildcard *.proto)
GO = $(PB:.proto=.pb.go)

all: $(GO)

.PHONY: target

%.pb.go: %.proto
	protoc -I=. -I=$(GOPATH)/src -I=$(GOPATH)/src/github.com/gogo/protobuf/protobuf -I../../../../../../src --gofast_out=plugins=grpc:. $<

.PHONY: clean
clean:
	-rm *.pb.go

