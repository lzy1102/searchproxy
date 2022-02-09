all: build

build:
	rm -rf bin/
	mkdir -p bin
	cp -rf config.json bin/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/checkproxy app/plugin/checkproxy/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/restful app/plugin/api/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/task app/task.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/scanproxy app/plugin/scanproxy/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/pushmsg app/plugin/pushmsg/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/save app/plugin/save/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/config app/plugin/config/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/scanport app/plugin/scanport/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/proxyscan app/plugin/proxyscan/main.go