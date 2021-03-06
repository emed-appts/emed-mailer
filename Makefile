NAME := emed-mailer
IMPORT := github.com/emed-appts/$(NAME)
DIST := dist

ifeq ($(OS), Windows_NT)
	EXECUTABLE := $(NAME).exe
	HAS_RETOOL := $(shell where retool ;)
	HAS_PACKR := $(shell where packr ;)
	HAS_MOCKERY := $(shell where mockery ;)
else
	EXECUTABLE := $(NAME)
	HAS_RETOOL := $(shell command -v retool ;)
	HAS_PACKR := $(shell command -v packr ;)
	HAS_MOCKERY := $(shell command -v mockery ;)
endif

PACKAGES ?= $(shell go list ./... | grep -v /vendor/ | grep -v /_tools/)
SOURCES ?= $(shell find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./_tools/*" -not -path "./test/*" -not -name "mock_*")

TAGS ?=

ifndef VERSION
	ifneq ($(DRONE_TAG),)
		VERSION ?= $(subst v,,$(DRONE_TAG))
	else
		ifneq ($(DRONE_BRANCH),)
			VERSION ?= 0.0.0-$(subst /,,$(DRONE_BRANCH))
		else
			VERSION ?= 0.0.0-master
		endif
	endif
endif

ifndef SHA
	SHA := $(shell git rev-parse --short HEAD)
endif

ifndef DATE
	DATE := $(shell date -u '+%Y%m%d')
endif

LDFLAGS += -s -w -X "$(IMPORT)/internal/version.VersionString=$(VERSION)" -X "$(IMPORT)/internal/version.VersionDev=$(SHA)" -X "$(IMPORT)/internal/version.VersionDate=$(DATE)"

.PHONY: all
all: build

.PHONY: update
update:
	retool do dep ensure -update

.PHONY: sync
sync:
	retool do dep ensure

.PHONY: clean
clean:
	go clean -i ./...
	retool do packr clean
	rm -rf bin/ $(DIST)/binaries $(DIST)/release

.PHONY: fmt
fmt:
	gofmt -s -w $(SOURCES)

.PHONY: vet
vet:
	go vet $(PACKAGES)

.PHONY: megacheck
megacheck:
	retool do megacheck -tags '$(TAGS)' $(PACKAGES)

.PHONY: lint
lint:
	for PKG in $(PACKAGES); do retool do golint -set_exit_status $$PKG || exit 1; done;

.PHONY: generate
generate:
	retool do go generate $(PACKAGES)

.PHONY: test
test:
	retool do goverage -v -coverprofile coverage.out $(PACKAGES)

.PHONY: install
install: $(SOURCES)
	go install -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' ./cmd/$(NAME)

.PHONY: build
build: bin/$(EXECUTABLE)

bin/$(EXECUTABLE): $(SOURCES)
	go build -i -v -tags '$(TAGS)' -ldflags '$(LDFLAGS)' -o $@ ./cmd/$(NAME)

.PHONY: release
release: release-dirs release-windows release-linux release-darwin release-copy release-check

.PHONY: release-dirs
release-dirs:
	mkdir -p $(DIST)/binaries $(DIST)/release

.PHONY: release-windows
release-windows:
ifeq ($(CI),drone)
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'windows/*' -out $(EXECUTABLE)-$(VERSION)  ./cmd/$(NAME)
	mv /build/* $(DIST)/binaries
else
	retool do xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'windows/*' -out $(EXECUTABLE)-$(VERSION)  ./cmd/$(NAME)
endif

.PHONY: release-linux
release-linux:
ifeq ($(CI),drone)
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'linux/*' -out $(EXECUTABLE)-$(VERSION)  ./cmd/$(NAME)
	mv /build/* $(DIST)/binaries
else
	retool do xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '-linkmode external -extldflags "-static" $(LDFLAGS)' -targets 'linux/*' -out $(EXECUTABLE)-$(VERSION)  ./cmd/$(NAME)
endif

.PHONY: release-darwin
release-darwin:
ifeq ($(CI),drone)
	xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '$(LDFLAGS)' -targets 'darwin/*' -out $(EXECUTABLE)-$(VERSION)  ./cmd/$(NAME)
	mv /build/* $(DIST)/binaries
else
	retool do xgo -dest $(DIST)/binaries -tags 'netgo $(TAGS)' -ldflags '$(LDFLAGS)' -targets 'darwin/*' -out $(EXECUTABLE)-$(VERSION)  ./cmd/$(NAME)
endif

.PHONY: release-copy
release-copy:
	$(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),cp $(file) $(DIST)/release/$(notdir $(file));)

.PHONY: release-check
release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),sha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

.PHONY: publish
publish: release

.PHONY: install-tools
install-tools: retool packr mockery

.PHONY: retool
retool:
ifndef HAS_RETOOL
	go get -u github.com/twitchtv/retool
endif
	retool sync
	retool build

.PHONY: packr
packr:
ifndef HAS_PACKR
	go get -u github.com/gobuffalo/packr/packr
endif

.PHONY: mockery
mockery:
ifndef HAS_MOCKERY
	go get -u github.com/vektra/mockery/...
endif
