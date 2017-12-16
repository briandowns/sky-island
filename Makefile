# This file has been written making the assumption that the user will be on a 
# BSD based system or one using BSD Make rather than GNU Make.

BINARY = sky-island

build:
	GOOS=freebsd GOARCH=amd64 go build -v -o bin/${BINARY}

install: clean build
	cp bin/${BINARY} /usr/local/bin
	cp contrib/rc.d/${BINARY} /usr/local/etc/rc.d
	echo 'sky-island_enable="YES"' >> /etc/rc.conf

mocks:
	mockery -dir=jail/ -all
	mockery -dir=filesystem/ -all
	mockery -dir=utils/ -all

test:
	go test -v -cover ./...

clean:
	go clean
	rm -f bin/*

release: clean 
	./release.sh
