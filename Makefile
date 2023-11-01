SHELL:=/bin/bash

all: build

build:
	go build -o bin/server server/main.go
	go build -o bin/client client/main.go