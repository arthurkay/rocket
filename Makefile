REGISTRY?=nerdygeek/rocket
APP_VERSION?=latest
.PHONY: default server client deps fmt clean all release-all assets client-assets server-assets contributors

default: deps build

deps: 
	go mod download

compress:
	upx rocketd
	upx rocket

server: deps
	go build -o rocketd ./cmd/rocketd

fmt:
	go fmt ./...

client: deps
	go build -o rocket ./cmd/rocket

compile-all:
	GOOS=linux GOARCH=386 go build -o rocket_linux_i386 ./cmd/rocket
	GOOS=windows GOARCH=386 go build -o rocket_windows_i386 ./cmd/rocket
	GOOS=linux GOARCH=arm64 go build -o rocket_linux_arm64 ./cmd/rocket
	GOOS=windows GOARCH=arm64 go build -o rocket_windows_arm64 ./cmd/rocket
	GOOS=linux GOARCH=amd64 go build -o rocket_linux_amd64 ./cmd/rocket
	GOOS=windows GOARCH=amd64 go build -o rocket_windows_amd64 ./cmd/rocket

	GOOS=linux GOARCH=386 go build -o rocketd_linux_i386 ./cmd/rocketd
	GOOS=windows GOARCH=386 go build -o rocketd_windows_i386 ./cmd/rocketd
	GOOS=linux GOARCH=arm64 go build -o rocketd_linux_arm64 ./cmd/rocketd
	GOOS=windows GOARCH=arm64 go build -o rocketd_windows_arm64 ./cmd/rocketd
	GOOS=linux GOARCH=amd64 go build -o rocketd_linux_amd64 ./cmd/rocketd
	GOOS=windows GOARCH=amd64 go build -o rocketd_windows_amd64 ./cmd/rocketd

build: client server

clean:
	go clean -i -r ./...

contributors:
	echo "Contributors to rocket, both large and small:\n" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS

registry: registry-build registry-push

registry-build:
	docker build --pull -t $(REGISTRY):$(APP_VERSION) .

registry-push:
	docker push $(REGISTRY):$(APP_VERSION)
