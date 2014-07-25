GOPATH=$(shell pwd)/.gopath

all:
	scripts/build.sh

dist:
	scripts/dist.sh

clean:
	rm -f bin/mqtt-bridgeify || true
	rm -rf .gopath || true

test:
	cd .gopath/src/github.com/ninjasphere/driver-go-gestic && go get -t ./...
	cd .gopath/src/github.com/ninjasphere/driver-go-gestic && go test ./...

vet:
	go vet ./...

.PHONY: all	dist clean test
