#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
BINDIR ?= $(GOPATH)/bin
APP = ./app

# export VERSION := $(shell echo $(shell git describe --tags --always --match "v*") | sed 's/^v//')
# export VERSION := v0.0.16

export COMMIT := $(shell git log -1 --format='%H')

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build
export GO111MODULE = on

# process build tags
build_tags = netgo osusergo

ifeq ($(LEDGER_ENABLED),true)
  # Add ledger build tags if necessary
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=pax \
          -X github.com/cosmos/cosmos-sdk/version.AppName=paxd \
          -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
          -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
          -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ifeq (,$(findstring nostrip,$(UNIGRID_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
ifeq (,$(findstring nostrip,$(UNIGRID_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/paxd

build:
	go build $(BUILD_FLAGS) -o bin/paxd ./cmd/paxd

# Add other targets like test, lint, clean, etc., as per your project's needs

.PHONY: all install build
