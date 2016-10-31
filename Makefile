#===============================================================================
#  release information
#===============================================================================
PACKAGE := $(shell go list)
BINARY := $(basename $(PACKAGE))

TOOL_DIR := _tool
RELEASE_DIR := _release
PKG_DEST_DIR := $(RELEASE_DIR)/.pkg

ALL_OS := linux
ALL_ARCH := amd64
OUTPUT := $(PKG_DEST_DIR)/{{.OS}}_{{.Arch}}/{{.Dir}}
# ALL_OS * ALL_ARCH の組み合わせで 'OS_ARCH OS_ARCH OS_ARCH ...' という文字列を作る。
ALL_OS_ARCH := $(foreach OS,$(ALL_OS),$(foreach ARCH,$(ALL_ARCH),$(OS)_$(ARCH)))


#===============================================================================
#  version information embedding
#===============================================================================
# git tag は `git tag -a 'x.y.z'` と -a オプションで明示的に
# 注釈としてバージョン番号を付けること。
VERSION := $(shell git describe --always --dirty 2>/dev/null || echo 'no git tag')
VERSION_PACKAGE := main
BUILD_DATE := $(shell date '+%F %T %Z')
BUILD_WITH := $(shell go version)
LD_FLAGS := -X '$(VERSION_PACKAGE).buildVersion=$(VERSION)' \
	-X '$(VERSION_PACKAGE).buildDate=$(BUILD_DATE)'         \
	-X '$(VERSION_PACKAGE).buildWith=$(BUILD_WITH)'


#===============================================================================
#  targets
#    `make [help]` shows tasks what you should execute.
#    The other are helper targets.
#===============================================================================
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
	(which glide &>/dev/null) || curl https://glide.sh/get | sh
	go get -v github.com/alecthomas/gometalinter
	go get -v github.com/mitchellh/gox
	go get -v github.com/tcnksm/ghr
	gometalinter --install

.PHONY: deps-install
deps-install: setup ## install vendor packages based on glide.lock or glide.yaml
	glide install

.PHONY: install
install: deps-install ## install the binary
	go install -ldflags "$(LD_FLAGS) -linkmode external -extldflags -static"

.PHONY: lint
lint: ## lint go sources and check whether only LICENSE file has copyright sentence
	gometalinter --deadline=60s --exclude=cryptographic $(shell glide novendor)
	$(TOOL_DIR)/copyright-check.sh

.PHONY: test
test: ## go test
	go test -v $(shell glide novendor)

.PHONY: all-build
all-build:
	gox -os='$(ALL_OS)' -arch='$(ALL_ARCH)' -ldflags="$(LD_FLAGS)" -output='$(OUTPUT)'

.PHONY: all-archive
all-archive:
	$(TOOL_DIR)/archive.sh "$(ALL_OS_ARCH)" "$(PKG_DEST_DIR)"

.PHONY: release
release: test all-build all-archive ## build binaries for all platforms and upload them to GitHub
	ghr "$(VERSION)" "$(RELEASE_DIR)"

.PHONY: clean
clean: ## uninstall the binary and remove $(RELEASE_DIR) directory
	go clean -i .
	rm -rf $(RELEASE_DIR)
