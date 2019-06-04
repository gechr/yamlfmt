TMPDIR = $(shell mktemp -d)

all: lint install

install:
	go install

lint:
	golangci-lint run

test:
	go test -race ./...

update:
	go get -u
	go mod tidy
	@go build -o $(TMPDIR)/main
