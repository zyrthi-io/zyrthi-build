.PHONY: build test clean install

BINARY := zyrthi-build

build:
	go build -o $(BINARY) .

test:
	go test -v -race -coverprofile=coverage.out ./...

clean:
	rm -f $(BINARY) coverage.out

install: build
	go install .
