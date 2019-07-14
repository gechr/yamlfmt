TMPDIR = $(shell mktemp -d)

all: lint install

install:
	go install

lint:
	golangci-lint run

test:
	go test -race ./...

update:
	go mod tidy
	go get -u
	@go build -o $(TMPDIR)/main
	@git diff -- go.mod go.sum || :
