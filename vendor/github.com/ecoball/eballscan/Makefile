# Copyright QuakerChain All Rights Reserved.

BASE_VERSION = 1.1.1

all: eballscan

.PHONY: eballscan
eballscan:
	@echo "\033[;32mbuild eballscan \033[0m"
	mkdir -p build/
	go build -v -o eballscan main.go
	mv eballscan build/


.PHONY: clean
clean:
	@echo "\033[;31mclean project \033[0m"
	-rm -rf build/

.PHONY: test
test:
	@echo "\033[;31mhello world \033[0m"
