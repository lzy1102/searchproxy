all: build

build:
	rm -rf bin/
	mkdir -p bin
	cp -rf config.json bin/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/task task.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/scanproxy plugin/scanproxy/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/pushmsg plugin/pushmsg/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/save plugin/save/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/config plugin/config/main.go