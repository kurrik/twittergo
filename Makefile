.PHONY: build run usage

usage:
	@echo '   build         Build the executable'
	@echo '   run           Execute the built program'

build:
	gd src -o build/main

run: build
	build/main
