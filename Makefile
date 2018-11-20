SHELL=/bin/bash

# Release info.
VERSION?=$(shell ./version.sh)
PRERELEASE=$(shell cat PRERELEASE)
GIT_COMMIT=$(shell git rev-parse HEAD)

# Cross-compile for these architecture when using `make build`.
NIX_OS_ARCHS?=darwin-amd64 linux-amd64
WIN_OS_ARCHS?=windows-amd64

# Parent directory of the Go entry point for each commands.
COMMAND_DIR=cmd

# Get the command names for list of directories.
COMMANDS=$(shell ls $(COMMAND_DIR))

# Parent directory of the output binaries.
DIST_DIR?=dist

# Compute the paths of the binaries for the specified architectures.
NIX_EXECS=$(foreach command, $(COMMANDS), $(foreach os-arch, $(NIX_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(command)))
WIN_EXECS=$(foreach command, $(COMMANDS), $(foreach os-arch, $(WIN_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(command).exe))
EXECS=$(NIX_EXECS) $(WIN_EXECS)

# Go build flags and commands (static by default).
GO_CMD=go
GO_LD_FLAGS?=-extldflags "-static"
CGO_ENABLED?=0
GO_BUILD=$(GO_CMD) build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -ldflags '-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) $(GO_LD_FLAGS)'


GIT_PATH=$(shell git rev-parse --show-toplevel)
GITHUB_REPO=$(shell basename $(GIT_PATH))
GITHUB_USER=$(shell basename $(shell dirname $(GIT_PATH)))
GIT_TAG=v$(VERSION)
RELEASE_NAME=$(GIT_TAG)
RELEASE_NOTES_FILE=RELEASE_NOTES.md
TEXT_FILES=LICENSE RELEASE_NOTES.md CHANGE_LOG.md
DOCKER_USER?=stratumn
DOCKER_FILE_TEMPLATE=Dockerfile.tpl
COVERAGE_FILE=coverage.txt
COVERHTML_FILE=coverhtml.txt
CLEAN_PATHS=$(DIST_DIR) $(COVERAGE_FILE) $(COVERHTML_FILE)
TMP_DIR:=$(shell mktemp -d)

GO_LINT_CMD=golangci-lint
KEYBASE_CMD=keybase
GITHUB_RELEASE_COMMAND=github-release
DOCKER_CMD=docker

DOCKER_GOLANG_IMAGE?=golang:1.11.2

GITHUB_RELEASE_FLAGS=--user '$(GITHUB_USER)' --repo '$(GITHUB_REPO)' --tag '$(GIT_TAG)'
GITHUB_RELEASE_RELEASE_FLAGS=$(GITHUB_RELEASE_FLAGS) --name '$(RELEASE_NAME)' --description "$$(cat $(RELEASE_NOTES_FILE))"


GO_LIST=$(GO_CMD) list
GO_TEST=$(GO_CMD) test
GO_BENCHMARK=$(GO_TEST) -bench .
GO_LINT=$(GO_LINT_CMD) run --build-tags="lint" --disable="gas" --deadline=4m --tests=false --skip-dirs="testutils" --skip-dirs="storetestcases"
KEYBASE_SIGN=$(KEYBASE_CMD) pgp sign
GITHUB_RELEASE_RELEASE=$(GITHUB_RELEASE_COMMAND) release $(GITHUB_RELEASE_RELEASE_FLAGS)
GITHUB_RELEASE_UPLOAD=$(GITHUB_RELEASE_COMMAND) upload $(GITHUB_RELEASE_FLAGS)
GITHUB_RELEASE_EDIT=$(GITHUB_RELEASE_COMMAND) edit $(GITHUB_RELEASE_RELEASE_FLAGS)
DOCKER_BUILD=$(DOCKER_CMD) build
DOCKER_PUSH=$(DOCKER_CMD) push

PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor)
TEST_PACKAGES=$(shell $(GO_LIST) ./... | grep -v vendor | grep -v testutil | grep -v testcases)
COVERAGE_SOURCES=$(shell find * -name '*.go' -not -path "testutil/*" -not -path "*testcases/*" | grep -v 'doc.go')
BUILD_SOURCES=$(shell find * -name '*.go' -not -path "testutil/*" -not -path "*testcases/*" | grep -v '_test.go' | grep -v 'doc.go')

SIGNATURES=$(foreach exec, $(EXECS), $(exec).sig)
NIX_ZIP_FILES=$(foreach command, $(COMMANDS), $(foreach os-arch, $(NIX_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(command).zip))
WIN_ZIP_FILES=$(foreach command, $(COMMANDS), $(foreach os-arch, $(WIN_OS_ARCHS), $(DIST_DIR)/$(os-arch)/$(command).zip))
ZIP_FILES=$(NIX_ZIP_FILES) $(WIN_ZIP_FILES)
DOCKER_FILES=$(foreach command, $(COMMANDS), $(DIST_DIR)/$(command).Dockerfile)
LICENSED_FILES=$(shell find * -name '*.go' -not -path "vendor/*" | grep -v mock | grep -v '^\./\.')

TEST_LIST=$(foreach package, $(TEST_PACKAGES), test_$(package))
BENCHMARK_LIST=$(foreach package, $(TEST_PACKAGES), benchmark_$(package))
GITHUB_UPLOAD_LIST=$(foreach file, $(ZIP_FILES), github_upload_$(firstword $(subst ., ,$(file))))
DOCKER_IMAGE_LIST=$(foreach command, $(COMMANDS), docker_image_$(command))
DOCKER_PUSH_LIST=$(foreach command, $(COMMANDS), docker_push_$(command))
CLEAN_LIST=$(foreach path, $(CLEAN_PATHS), clean_$(path))

# == .PHONY ===================================================================
.PHONY: test coverage benchmark lint build git_tag github_draft github_upload github_publish docker_images docker_push clean test_headers $(TEST_LIST) $(BENCHMARK_LIST) $(GITHUB_UPLOAD_LIST) $(DOCKER_IMAGE_LIST) $(DOCKER_PUSH_LIST) $(CLEAN_LIST)

# == all ======================================================================
all: build

# == release ==================================================================
release: test lint clean build git_tag github_draft github_upload github_publish docker_images docker_push

# == test =====================================================================
test: $(TEST_LIST)

$(TEST_LIST): test_%:
	@$(GO_TEST) $* $(GO_TEST_OPTS)

test_wo_cache: GO_TEST_OPTS=-count=1
test_wo_cache: $(TEST_LIST)

test_integration: GO_TEST_OPTS=-integration
test_integration: $(TEST_LIST)

# == coverage =================================================================
coverage: $(COVERAGE_FILE)

$(COVERAGE_FILE): $(COVERAGE_SOURCES)
	@for d in $(TEST_PACKAGES); do \
	    $(GO_TEST) -coverprofile=profile.out -covermode=atomic $$d || exit 1; \
	    if [ -f profile.out ]; then \
	        cat profile.out >> $(COVERAGE_FILE); \
	        rm profile.out; \
	    fi \
	done

coverhtml:
	echo 'mode: set' > $(COVERHTML_FILE)
	@for d in $(TEST_PACKAGES); do \
	    $(GO_TEST) -coverprofile=profile.out $$d || exit 1; \
	    if [ -f profile.out ]; then \
	        tail -n +2 profile.out >> $(COVERHTML_FILE); \
	        rm profile.out; \
	    fi \
	done
	$(GO_CMD) tool cover -html $(COVERHTML_FILE)


# == benchmark ================================================================
benchmark: $(BENCHMARK_LIST)

$(BENCHMARK_LIST): benchmark_%:
	@$(GO_BENCHMARK) -benchmem $*

# == list =====================================================================
lint:
	@$(GO_LINT) ./...

# == build ====================================================================
build: $(EXECS)

BUILD_OS_ARCH=$(word 2, $(subst /, ,$@))
BUILD_OS=$(firstword $(subst -, ,$(BUILD_OS_ARCH)))
BUILD_ARCH=$(lastword $(subst -, ,$(BUILD_OS_ARCH)))
BUILD_COMMAND=$(firstword $(word 1, $(subst ., ,$(lastword $(subst /, ,$@)))))
BUILD_PACKAGE=$(shell $(GO_LIST) ./$(COMMAND_DIR)/$(BUILD_COMMAND))

$(EXECS): $(BUILD_SOURCES)
	GOOS=$(BUILD_OS) GOARCH=$(BUILD_ARCH) CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) -o $@ $(BUILD_PACKAGE)

# == build_docker =============================================================
# Builds dynamically linked binaries from within a Docker container.
# Warning: this is really slow on a Mac.
build_docker: $(BUILD_SOURCES)
	docker run \
		--rm \
		--volume ${PWD}:/go/src/github.com/$(GITHUB_USER)/$(GITHUB_REPO) \
		--workdir /go/src/github.com/$(GITHUB_USER)/$(GITHUB_REPO) \
		--env WIN_OS_ARCHS= \
		--env NIX_OS_ARCHS=linux-amd64 \
		--env CGO_ENABLED=1 \
		--env GO_LD_FLAGS= \
		$(DOCKER_GOLANG_IMAGE) make build

# == sign =====================================================================
sign: $(SIGNATURES)

%.sig: %
	$(KEYBASE_SIGN) -d -i $* -o $@

# == zip ======================================================================
zip: $(ZIP_FILES)

ZIP_TMP_OS_ARCH_DIR=$(TMP_DIR)/$(BUILD_OS_ARCH)
ZIP_TMP_CMD_DIR=$(ZIP_TMP_OS_ARCH_DIR)/$(BUILD_COMMAND)

%.zip: %.exe %.exe.sig
	mkdir -p $(ZIP_TMP_CMD_DIR)
	cp $*.exe $(ZIP_TMP_CMD_DIR)
	cp $*.exe.sig $(ZIP_TMP_CMD_DIR)
	cp $(TEXT_FILES) $(ZIP_TMP_CMD_DIR)
	mv $(ZIP_TMP_CMD_DIR)/LICENSE $(ZIP_TMP_CMD_DIR)/LICENSE.txt
	cd $(ZIP_TMP_OS_ARCH_DIR) && zip -r $(BUILD_COMMAND){.zip,} 1>/dev/null
	cp $(ZIP_TMP_CMD_DIR).zip $@

%.zip: % %.sig
	mkdir -p $(ZIP_TMP_CMD_DIR)
	cp $* $(ZIP_TMP_CMD_DIR)
	cp $*.sig $(ZIP_TMP_CMD_DIR)
	cp $(TEXT_FILES) $(ZIP_TMP_CMD_DIR)
	cd $(ZIP_TMP_OS_ARCH_DIR) && zip -r $(BUILD_COMMAND){.zip,} 1>/dev/null
	cp $(ZIP_TMP_CMD_DIR).zip $@

# == git_tag ==================================================================
git_tag:
	git tag $(GIT_TAG)
	git push origin --tags

# == github_draft =============================================================
github_draft:
	@if [[ "$(PRERELEASE)" != "false" ]]; then \
		echo $(GITHUB_RELEASE_RELEASE) --draft --pre-release; \
		$(GITHUB_RELEASE_RELEASE) --draft --pre-release; \
	else \
		echo $(GITHUB_RELEASE_RELEASE) --draft; \
		$(GITHUB_RELEASE_RELEASE) --draft; \
	fi

# == github_upload ============================================================
github_upload: $(GITHUB_UPLOAD_LIST)

$(GITHUB_UPLOAD_LIST): github_upload_%: %.zip
	$(GITHUB_RELEASE_UPLOAD) --file $*.zip --name $(BUILD_COMMAND)-$(BUILD_OS_ARCH).zip

# == github_publish ===========================================================
github_publish:
	@if [[ "$(PRERELEASE)" != "false" ]]; then \
		echo $(GITHUB_RELEASE_EDIT) --pre-release; \
		$(GITHUB_RELEASE_EDIT) --pre-release; \
	else \
		echo $(GITHUB_RELEASE_EDIT); \
		$(GITHUB_RELEASE_EDIT); \
	fi

# == docker_files =============================================================
docker_files: $(DOCKER_FILES)

DOCKER_EXTRA=./$(COMMAND_DIR)/$*/Docker

$(DIST_DIR)/%.Dockerfile: $(DOCKER_FILE_TEMPLATE)
	@mkdir -p $(DIST_DIR)
	@sed 's/{{CMD}}/$*/g' $(DOCKER_FILE_TEMPLATE) > $@
	@if [[ -f $(DOCKER_EXTRA) ]]; then \
		echo cat $(DOCKER_EXTRA) \>\> $@; \
		cat $(DOCKER_EXTRA) >> $@; \
	fi

# == docker_images ============================================================
docker_images: $(DOCKER_IMAGE_LIST)

DOCKER_IMAGE=$(DOCKER_USER)/$*

$(DOCKER_IMAGE_LIST): docker_image_%: $(DIST_DIR)/%.Dockerfile $(DIST_DIR)/linux-amd64/%
	@echo $(DOCKER_BUILD) -f $*.Dockerfile -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest .
	@# docker context is built in a temporary directory to improve performances. \
	# - Copy mandatory files or directories in the context directory \
	# - Rename copied files without project hierarchy in Dockerfile to fit with context directory
	@mkdir -p $(DIST_DIR)/$* && \
		cd $(DIST_DIR)/$* && \
		ln ../linux-amd64/$* && \
		for i in `sed -n 's/^COPY *\([^ ]*\).*/\1/p' ../$*.Dockerfile`; do test `basename $$i` = $* || cp -r ../../$$i .; done && \
		sed 's!^COPY *.*/\([^/]*\) \(.*\)!COPY \1 \2!' ../$*.Dockerfile > $*.Dockerfile && \
		$(DOCKER_BUILD) -f $*.Dockerfile -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest . ; \
		cd - ; rm -fr $(DIST_DIR)/$*

# == docker_push ==============================================================
docker_push: $(DOCKER_PUSH_LIST)

$(DOCKER_PUSH_LIST): docker_push_%:
	$(DOCKER_PUSH) $(DOCKER_IMAGE):$(VERSION)
	@if [[ "$(VERSION)" = "master" ]]; then \
	  $(DOCKER_PUSH) $(DOCKER_IMAGE):latest; \
	fi

# == license_headers ==========================================================
license_headers: $(LICENSED_FILES)

$(LICENSED_FILES): LICENSE_HEADER
	perl -i -0pe 's/\/\/ Copyright \d* Stratumn.*\n(\/\/.*\n)*/`cat LICENSE_HEADER`/ge' $@

# == clean ====================================================================
clean: $(CLEAN_LIST)

$(CLEAN_LIST): clean_%:
	rm -rf $*
 
test_headers:
	@ ./test_headers.sh $(LICENSED_FILES)
