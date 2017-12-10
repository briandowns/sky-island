# This file has been written making the assumption that the user will be on a 
# BSD based system or one using BSD Make rather than GNU Make.

BINARY = si-server

build:
	GOOS=freebsd GOARCH=amd64 go build -v -o bin/${BINARY}

mocks:
	mockery -dir=jail/ -all
	mockery -dir=filesystem/ -all
	mockery -dir=utils/ -all

test:
	go test -v -cover ./...

clean:
	go clean
	rm -f bin/${BINARY}*
