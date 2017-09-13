NAME       := gowhoson
SRCS       := $(shell git ls-files '*.go' | grep -v '.pb.go')
PWD        := $(shell pwd)
PKGS       := ./cmd/gowhoson ./whoson
DOCS       := README.md
VERSION    := $(shell git describe --tags --abbrev=0)
REVISION   := $(shell git rev-parse --short HEAD)
GOVERSION  := $(shell go version | cut -d ' ' -f3 | sed 's/^go//')
SRCDIR     := rpmbuild/SOURCES
RELEASE    := 1
IMAGE_NAME := $(NAME)-build
TARGZ_FILE := $(NAME).tar.gz
UID        := $(shell id -u)
LDFLAGS    := -s -X 'main.gVersion=$(VERSION)' \
                 -X 'main.gGitcommit=$(REVISION)' \
                 -X 'main.gGoversion=$(GOVERSION)'

all: deps test build

setup:
	go get -u github.com/golang/dep/...
	go get -u github.com/golang/lint/golint
	go get -u github.com/client9/misspell/cmd/misspell
	go get -u github.com/gordonklaus/ineffassign
	go get -u github.com/fzipp/gocyclo

pb:
	protoc --go_out=plugins=grpc:. whoson/sync.proto

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

cover: ## Update coverage.out
	@$(foreach pkg,$(PKGS),cd $(pkg); go test -coverprofile=coverage.out;cd $(PWD) || exit;)
	@$(foreach pkg,$(PKGS),cat $(pkg)/coverage.out >> _coverage.out; rm -f $(pkg)/coverage.out || exit;)
	@cat _coverage.out | sort -r | uniq > coverage.out
	@rm -f _coverage.out

coverview: ## Coverage view
	@go tool cover -html=coverage.out

goviz: ## Create struct map
	@goviz -i github.com/tai-ga/gowhoson | dot -Tpng -o goviz.png

$(SRCDIR)/$(NAME):
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(SRCDIR)/$(NAME)
	@docker images | grep -q $(IMAGE_NAME) && docker rmi $(IMAGE_NAME) || true;
	@docker images | grep -q $(IMAGE_NAME)-login && docker rmi $(IMAGE_NAME)-login || true;

%.bin: rpmbuild/SPECS/$(NAME).spec $(SRCDIR)/$(NAME)
	@docker build --build-arg UID=$(UID) --build-arg NAME=$(NAME) --build-arg VERSION=$(VERSION) \
		--build-arg RELEASE=$(RELEASE) -t $(IMAGE_NAME) -f Dockerfile.build .
	@docker build -t $(IMAGE_NAME)-login -f Dockerfile.login .
	@docker run --name $(IMAGE_NAME)-tmp $(IMAGE_NAME)
	@docker wait $(IMAGE_NAME)-tmp
	@docker cp $(IMAGE_NAME)-tmp:/tmp/$(TARGZ_FILE) /tmp
	@docker rm $(IMAGE_NAME)-tmp
	@[ ! -d $@ ] && mkdir $@ || :
	@tar zxf /tmp/$(TARGZ_FILE) -C $@
	@[ -f /tmp/$(TARGZ_FILE) ] && rm -f /tmp/$(TARGZ_FILE) || :

rpm: rpm.bin ## Build rpms for CentOS6
rpm-login: rpm ## Login build environment for CentOS6
	docker run --rm  -v $(PWD)/rpmbuild/SOURCES:/rpmbuild/SOURCES \
	-v $(PWD)/rpmbuild/SPECS:/rpmbuild/SPECS \
	-v $(PWD)/rpm.bin/RPMS:/rpmbuild/RPMS \
	-v $(PWD)/rpm.bin/SRPMS:/rpmbuild/SRPMS \
	-it $(IMAGE_NAME)-login /bin/bash

clean: ## Clean up
	@rm -f $(NAME)
	@rm -f _coverage.out coverage.out
	@rm -f goviz.png
	@rm -f rpmbuild/SOURCES/$(NAME)
	@rm -rf vendor
	@rm -rf rpm.bin
	@docker images | grep -q $(IMAGE_NAME) && docker rmi $(IMAGE_NAME) || true;
	@docker images | grep -q $(IMAGE_NAME)-login && docker rmi $(IMAGE_NAME)-login || true;

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all setup deps lint misspell ineffassign gocyclo dep depup vet fmt test build clean cover coverview goviz help
