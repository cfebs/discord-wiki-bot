SRC_GO_FILES := $(shell find . -iname "*.go")

discord-wiki-bot: $(SRC_GO_FILES) go.sum Makefile
	go test ./... && go build -o discord-wiki-bot ./cli/main.go

.PHONY: test
test:
	go test -v ./...
