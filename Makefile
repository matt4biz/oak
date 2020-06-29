version=$(shell git describe --tags --long --dirty 2>/dev/null)

oak:
	go install -ldflags "-X main.version=$(version)" ./cmd/oak.go

lint:
	golangci-lint run

test:
	go test -v ./... -coverprofile=c.out -covermode=count

clean:
	rm -f oak
	rm -f $(GOPATH)/bin/oak
