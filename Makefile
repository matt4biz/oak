version=$(shell git describe --tags --long --dirty 2>/dev/null)

## NOTE: we can't use go install because it
## doesn't have the -o option to name the file

oak:
	go build -ldflags "-X main.version=$(version)" -o $@ ./cmd && mv $@ $(GOPATH)/bin

lint:
	golangci-lint run

test:
	go test -v ./... -coverprofile=c.out -covermode=count

clean:
	rm -f oak
	rm -f $(GOPATH)/bin/oak
	go clean -testcache
