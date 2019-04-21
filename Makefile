all: build

run:
	go run main.go

build:
	go build

install: build
	sudo install -m 755 main -T  /usr/local/bin/easyssh
