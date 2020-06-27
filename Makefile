oak:
	go install ./cmd/oak.go

lint:
	golangci-lint run

test:
	go test ./... -coverprofile=c.out -covermode=count

clean:
	rm -f oak
	rm -f $(GOPATH)/bin/oak
