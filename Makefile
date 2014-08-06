GOPATH=$(shell pwd)/.gopath

debug:
	scripts/build.sh

clean:
	rm -f bin/driver-go-gestic || true
	rm -rf .gopath || true

test:
	cd .gopath/src/github.com/ninjasphere/driver-go-gestic && go get -t ./...
	cd .gopath/src/github.com/ninjasphere/driver-go-gestic && go test ./...

vet:
	go vet ./...

.PHONY: debug clean test vet
