NAME       := gowhoson
SRCS       := $(shell git ls-files '*.go' | grep -v '.pb.go' | grep -v 'tools/tools.go')
PWD        := $(shell pwd)
PKGS       := ./internal/gowhoson ./pkg/whoson
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
                 -X 'main.gGitcommit=$(REVISION)'

INSTCMD             := golint misspell ineffassign gocyclo goviz \
                       protoc-gen-go protoc-gen-go-grpc staticcheck
INSTCMD_golint      := golang.org/x/lint/golint@v0.0.0-20210508222113-6edffad5e616
INSTCMD_misspell    := github.com/client9/misspell/cmd/misspell@v0.3.4
INSTCMD_ineffassign := github.com/gordonklaus/ineffassign@v0.0.0-20210914165742-4cc7213b9bc8
INSTCMD_gocyclo     := github.com/fzipp/gocyclo/cmd/gocyclo@v0.4.0
INSTCMD_goviz       := github.com/trawler/goviz@v0.0.0-20181113143047-634081648655
INSTCMD_protoc-gen-go := google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
INSTCMD_protoc-gen-go-grpc := google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
INSTCMD_staticcheck := honnef.co/go/tools/cmd/staticcheck@v0.3.0-0.dev.0.20220306074811-23e1086441d2

TOOLS_DIR := $(abspath ./.tools)

#
# Install commands
#
define instcmd
.PHONY: _instcmd_$(1)
_instcmd_$(1):
	@if [ ! -f $(TOOLS_DIR)/$1 ]; then \
		echo "install $1" && \
		GOBIN=$(TOOLS_DIR) go install $2; \
	fi
endef

all: instcmd test build

instcmd: $(addprefix _instcmd_,$(INSTCMD))
$(foreach p,$(INSTCMD),$(eval $(call instcmd,$(p),$(INSTCMD_$(p)))))

pb:
	protoc \
		--plugin=$(TOOLS_DIR)/protoc-gen-go \
		--plugin=$(TOOLS_DIR)/protoc-gen-go-grpc \
		--go_out=. \
		--go-grpc_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		./pkg/whoson/sync.proto

lint:
	@$(foreach file,$(SRCS),$(TOOLS_DIR)/golint --set_exit_status $(file) || exit;)

staticcheck:
	@$(TOOLS_DIR)/staticcheck ./...

misspell:
	@$(foreach file,$(DOCS),$(TOOLS_DIR)/misspell -error $(file) || exit;)
	@$(foreach file,$(SRCS),$(TOOLS_DIR)/misspell -error $(file) || exit;)

ineffassign:
	@$(foreach pkg,$(PKGS),$(TOOLS_DIR)/ineffassign $(pkg) || exit;)

gocyclo:
	@$(foreach pkg,$(PKGS),$(TOOLS_DIR)/gocyclo -over 15 $(pkg) || exit;)

vet:
	@$(foreach pkg,$(PKGS),go vet $(pkg) || exit;)

fmt:
	@$(foreach file,$(SRCS),go fmt $(file) || exit;)

test: instcmd lint staticcheck misspell ineffassign gocyclo vet fmt ## Test
	$(foreach pkg,$(PKGS),go test -cover -v $(pkg) || exit;)

build: ## Build program
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(NAME) $< cmd/gowhoson/main.go

cover: ## Update coverage.out
	@$(foreach pkg,$(PKGS),cd $(pkg); go test -coverprofile=coverage.out;cd $(PWD) || exit;)
	@$(foreach pkg,$(PKGS),cat $(pkg)/coverage.out >> _coverage.out; rm -f $(pkg)/coverage.out || exit;)
	@cat _coverage.out | sort -r | uniq > coverage.out
	@rm -f _coverage.out

coverview: ## Coverage view
	@go tool cover -html=coverage.out

goviz: ## Create struct map
	$(TOOLS_DIR)/goviz -i cmd/gowhoson | dot -Tpng -o goviz.png

$(SRCDIR)/$(NAME):
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(SRCDIR)/$(NAME) cmd/gowhoson/main.go
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
	@rm -f _coverage.out coverage.out coverage.html
	@rm -f goviz.png
	@rm -f rpmbuild/SOURCES/$(NAME)
	@rm -rf vendor
	@rm -rf rpm.bin
	@rm -rf dist
	@rm -rf $(TOOLS_DIR)
	@docker images | grep -q $(IMAGE_NAME) && docker rmi $(IMAGE_NAME) || true;
	@docker images | grep -q $(IMAGE_NAME)-login && docker rmi $(IMAGE_NAME)-login || true;

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all instcmd pb lint staticcheck misspell ineffassign gocyclo vet fmt test build cover coverview goviz clean help
