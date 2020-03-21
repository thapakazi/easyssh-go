all: build

run:
	go run main.go

build:
	go build -o main

install: build
	sudo install -m 755 main -T  /usr/local/bin/easyssh
