#===============================================================================
#  release information
#===============================================================================
PACKAGE := $(shell go list)
BINARY := $(notdir $(PACKAGE))

TOOL_DIR := _tool
RELEASE_DIR := _release
PKG_DEST_DIR := $(RELEASE_DIR)/.pkg

ALL_OS := linux
ALL_ARCH := amd64

LATEST_LOCAL_BRANCH := $(subst * ,,$(shell git branch --sort='-committerdate' |\
	head --lines=1))
NEW_TAG := $(shell echo "$(LATEST_LOCAL_BRANCH)"      |\
	grep --only-matching -E '[0-9]+\.[0-9]+\.[0-9]+')


#===============================================================================
#  version information embedding
#===============================================================================
# バージョンタグは `git tag -a 'x.y.z'` と注釈付きタグであることが前提。
VERSION := $(shell git describe --always --dirty 2>/dev/null || echo 'no git tag')
VERSION_PACKAGE := main
BUILD_REVISION := $(shell git rev-parse --short HEAD)
BUILD_WITH := $(shell go version)
LD_FLAGS := -s -w -X '$(VERSION_PACKAGE).buildVersion=$(VERSION)' \
	-X '$(VERSION_PACKAGE).buildRevision=$(BUILD_REVISION)'       \
	-X '$(VERSION_PACKAGE).buildWith=$(BUILD_WITH)'               \
	-extldflags -static


#===============================================================================
#  targets
#    `make [help]` shows tasks what you should execute.
#    The other are helper targets.
#===============================================================================
SHELL := /bin/bash
.DEFAULT_GOAL := help

# [Add a help target to a Makefile that will allow all targets to be self documenting]
# https://gist.github.com/prwhite/8168133
.PHONY: help
help: ## show help
	@echo 'USAGE: make [target]'
	@echo
	@echo 'TARGETS:'
	@grep -E '^[-_: a-zA-Z0-9]+##' $(MAKEFILE_LIST)  |\
		sed -e 's/:[-_ a-zA-Z0-9]\+/:/'              |\
		column -t -s ':#'

# install development tools
.PHONY: setup
setup:
	type -a glide &>/dev/null || curl https://glide.sh/get | sh
	go get -v -u github.com/alecthomas/gometalinter
	go get -v -u github.com/tcnksm/ghr
	gometalinter --install

.PHONY: deps-install
deps-install: setup ## install vendor packages based on glide.lock or glide.yaml
	glide install --strip-vendor

.PHONY: install
install:
	CGO_ENABLED=0 go install -a -tags netgo -installsuffix netgo -ldflags "$(LD_FLAGS)"

.PHONY: lint
lint: ## lint go sources and check whether only LICENSE file has copyright sentence
	glide list || glide install
	gometalinter --errors --enable-all --deadline=60s $(shell glide novendor)
	$(TOOL_DIR)/copyright-check.sh

.PHONY: push-release
push-release: lint test ## update CHANGELOG and push all of the your development works
	$(TOOL_DIR)/add-changelog.sh "$(NEW_TAG)"
	git checkout master
	git merge --ff "$(LATEST_LOCAL_BRANCH)"
	git push
	$(TOOL_DIR)/add-release-tag.sh "$(NEW_TAG)"

.PHONY: test
test: ## go test
	go test -v $(shell glide novendor)

.PHONY: all-build
all-build: lint test
	$(TOOL_DIR)/build-static-bins.sh "$(ALL_OS)" "$(ALL_ARCH)" "$(LD_FLAGS)" "$(PKG_DEST_DIR)" "$(BINARY)"

.PHONY: all-archive
all-archive:
	$(TOOL_DIR)/archive.sh "$(ALL_OS)" "$(ALL_ARCH)" "$(PKG_DEST_DIR)"

.PHONY: release
release: all-build all-archive ## build binaries for all platforms and upload them to GitHub
	ghr "$(VERSION)" "$(RELEASE_DIR)"

.PHONY: clean
clean: ## uninstall the binary and remove $(RELEASE_DIR) directory
	go clean -i .
	rm -rf $(RELEASE_DIR)
