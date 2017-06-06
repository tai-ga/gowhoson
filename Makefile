NAME      := gowhoson
SRCS      := $(shell git ls-files '*.go')
PKGS      := ./cmd/gowhoson ./whoson
VERSION   := $(shell git describe --tags --abbrev=0)
REVISION  := $(shell git rev-parse --short HEAD)
GOVERSION := $(shell go version | cut -d ' ' -f3 | sed 's/^go//')
LDFLAGS   := -s -X 'main.gVersion=$(VERSION)' \
                -X 'main.gGitcommit=$(REVISION)' \
                -X 'main.gGoversion=$(GOVERSION)'

all: deps test build

setup:
	go get -u github.com/golang/dep/...
#	go get -u github.com/golang/lint/golint

deps: setup
	dep ensure

#lint:
#	$(foreach file,$(SRCS),golint $(file) || exit;)

dep: ## dep ensure
	dep ensure

depup: ## dep -update
	dep ensure -update

vet:
	$(foreach pkg,$(PKGS),go vet $(pkg);)

fmt:
	$(foreach file,$(SRCS),go fmt $(file);)

#test: lint vet  fmt ## Test
test: vet  fmt ## Test
	$(foreach pkg,$(PKGS),go test -cover -v $(pkg) || exit;)

build: ## Build program
	go build -ldflags "$(LDFLAGS)" -o $(NAME) $<

clean: ## clean up
	@rm -f $(NAME)
	@rm -rf vendor

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

#.PHONY: all setup deps lint dep depup vet fmt test build clean help
.PHONY: all setup deps dep depup vet fmt test build clean help
