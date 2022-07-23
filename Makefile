VERSION = 0.2.0
RELEASE = 1
DIST_DIR = $(shell pwd)/dist
#OS :=$(shell awk -F= '/^ID/{print $$2}' /etc/os-release)
#BUILDROOT ?=

# Docker command to use, can be podman
DOCKER ?= docker

##@ Build

build: fmt lint vet
	go build -mod=vendor -o bin/flotta main.go

LINT_IMAGE=golangci/golangci-lint:v1.45.0
lint: ## Check if the go code is properly written, rules are in .golangci.yml
	$(DOCKER) run --rm -v $(CURDIR):$(CURDIR) -w="$(CURDIR)" $(LINT_IMAGE) sh -c 'golangci-lint run'

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

##@ Development

cobra: ## Download cobra locally if necessary.
ifeq (, $(shell which cobra))
	@(cd /tmp && go get github.com/mitchellh/go-homedir@v1.1.0)
	@(cd /tmp/ && go get github.com/spf13/viper@v1.12.0)
	$(call go-install-tool,$(COBRA),github.com/spf13/cobra@v1.5.0)
endif

vendor:
	go mod tidy -go=1.17
	go mod vendor

rpm-tarball:
	 (git archive --prefix flotta-dev-cli-$(VERSION)/ HEAD ) \
	    | gzip > flotta-dev-cli-$(VERSION).tar.gz

rpm-src: rpm-tarball
	install -D -m644 flotta-dev-cli-$(VERSION).tar.gz --target-directory `rpmbuild -E %_sourcedir`
	rpmbuild -bs \
		-D "VERSION $(VERSION)" \
		-D "RELEASE $(RELEASE)" \
		--buildroot $(DIST_DIR) ./flotta-dev-cli.spec

rpm-copr-testing: rpm-src
	copr build project-flotta/flotta-testing $(HOME)/rpmbuild/SRPMS/flotta-dev-cli-$(VERSION)-$(RELEASE).*.src.rpm

rpm-copr: rpm-src
	copr build project-flotta/flotta $(HOME)/rpmbuild/SRPMS/flotta-dev-cli-$(VERSION)-$(RELEASE).*.src.rpm

rpm-build: rpm-src
	rpmbuild $(RPMBUILD_OPTS) --rebuild $(HOME)/rpmbuild/SRPMS/flotta-dev-cli-$(VERSION)-$(RELEASE).*.src.rpm

rpm: ## Create rpm build
	RPMBUILD_OPTS=--target=x86_64 $(MAKE) rpm-build

# go-install-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef