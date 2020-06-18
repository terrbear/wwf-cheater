.PHONY: clean build

init:
	rm words.db
	go run main.go -init

build:
	mkdir -p bin
	go build -o ./bin/wwf main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

lint:
	golangci-lint run ./...

install: build
	cp ./bin/wwf ~theath/bin
