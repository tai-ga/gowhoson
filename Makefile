NAME      := gowhoson
SRCS      := $(shell git ls-files '*.go')
PWD       := $(shell pwd)
PKGS      := ./cmd/gowhoson ./whoson
DOCS      := README.md
VERSION   := $(shell git describe --tags --abbrev=0)
REVISION  := $(shell git rev-parse --short HEAD)
GOVERSION := $(shell go version | cut -d ' ' -f3 | sed 's/^go//')
LDFLAGS   := -s -X 'main.gVersion=$(VERSION)' \
                -X 'main.gGitcommit=$(REVISION)' \
                -X 'main.gGoversion=$(GOVERSION)'

all: deps test build

setup:
	go get -u github.com/golang/dep/...
	go get -u github.com/golang/lint/golint
	go get -u github.com/client9/misspell/cmd/misspell
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/fzipp/gocyclo

deps: setup
	dep ensure

lint:
	@$(foreach file,$(SRCS),golint --set_exit_status $(file) || exit;)

misspell:
	@$(foreach file,$(DOCS),misspell -error $(file) || exit;)
	@$(foreach file,$(SRCS),misspell -error $(file) || exit;)

ineffassign:
	@$(foreach file,$(SRCS),ineffassign $(file) || exit;)

gocyclo:
	@$(foreach pkg,$(PKGS),gocyclo -over 15 $(pkg) || exit;)

dep: ## dep ensure
	dep ensure

depup: ## dep -update
	dep ensure -update

vet:
	@$(foreach pkg,$(PKGS),go vet $(pkg) || exit;)

fmt:
	@$(foreach file,$(SRCS),go fmt $(file) || exit;)

test: lint misspell ineffassign gocyclo vet fmt ## Test
	$(foreach pkg,$(PKGS),go test -cover -v $(pkg) || exit;)

build: ## Build program
	go build -ldflags "$(LDFLAGS)" -o $(NAME) $<

clean: ## Clean up
	@rm -f $(NAME)
	@rm -rf vendor
	@rm -rf _coverage.out coverage.out

cover: ## Update coverage.out
	@$(foreach pkg,$(PKGS),cd $(pkg); go test -coverprofile=coverage.out;cd $(PWD) || exit;)
	@$(foreach pkg,$(PKGS),cat $(pkg)/coverage.out >> _coverage.out; rm -f $(pkg)/coverage.out || exit;)
	@cat _coverage.out | sort -r | uniq > coverage.out
	@rm -f _coverage.out

coverview: ## Coverage view
	@go tool cover -html=coverage.out

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all setup deps lint misspell ineffassign gocyclo dep depup vet fmt test build clean cover coverview help
