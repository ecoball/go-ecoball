# Copyright QuakerChain All Rights Reserved.

BASE_VERSION = 1.1.1

all: ecoball ecoclient proto

.PHONY: proto ecoball ecoclient
ecoball: proto
	@echo "\033[;32mbuild ecoball \033[0m"
	mkdir -p build/
	go build -v -o ecoball node/*.go
	mv ecoball build/

ecoclient: 
	@echo "\033[;32mbuild ecoclient \033[0m"
	mkdir -p build/
	go build -v -o ecoclient client/client.go
	mv ecoclient build/

proto:
	@echo "\033[;32mbuild protobuf file \033[0m"
	make -C core/pb
	make -C client/protos
	make -C net/message/pb

.PHONY: clean
clean:
	@echo "\033[;31mclean project \033[0m"
	-rm -rf build/
	make -C core/pb/ clean
	make -C client/protos clean

.PHONY: test
test:
	@echo "\033[;31mhello world \033[0m"
