VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)

LDFLAGS := -gcflags=all=-trimpath=${PWD} -asmflags=all=-trimpath=${PWD} -ldflags=-extldflags=-zrelro -ldflags=-extldflags=-znow
ifdef VERSION
	LDFLAGS += -ldflags '-s -w -X main.commit=$(COMMIT) -X main.branch=$(BRANCH) -X main.version=$(VERSION)'
else
	VERSION := $(BRANCH)-$(COMMIT)
	LDFLAGS += -ldflags '-s -w -X main.commit=$(COMMIT) -X main.branch=$(BRANCH)'
endif
MOD := -mod=vendor
export G111MODULE=on
PKGS := $(shell go list ./... | grep -v /vendor)
GOCC := $(shell go version)

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(shell go env GOPATH))/bin:$(PATH)
endif

ifneq (,$(findstring gccgo,$(GOCC)))
	export GOPATH=$(shell pwd)/.go
	LDFLAGS := -gccgoflags '-s -w'
	MOD :=
endif

default: build

.PHONY: all
all: lint test build

.PHONY: test
test:
	go test $(PKGS)

.PHONY: lint
lint:
	golangci-lint run ./
	for p in ${GOPLUGINS}; do golangci-lint run $$p; done

.PHONY: install
install:
	install -d -m 644 $(DESTDIR)/var/log/gogios
	install -d -m 664 $(DESTDIR)/etc/gogios
	install -d $(DESTDIR)/usr/bin
	install -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/views/ -type d); do install -d -m 764 $(DESTDIR)/usr/share/gogios/views/$$(echo $$d | cut -d"/" -f3-); done
	for f in $$(find web/views/ -type f); do install -D --mode 764 "$$f" $(DESTDIR)/usr/share/gogios/views/$$(echo $$f | cut -d"/" -f3-); done
	install -m 664 package_files/example.json $(DESTDIR)/etc/gogios
	install -m 664 package_files/gogios.toml $(DESTDIR)/etc/gogios/gogios.sample.toml
	install -o root -g root -m 644 scripts/gogios.service $(DESTDIR)/usr/lib/systemd/system
	install -o root -g root -T -m 755 scripts/gogios-parse-nmap $(DESTDIR)/usr/bin/gogios-parse-nmap
	ls bin
	install -o root -g root -T -m 755 bin/gogios-$(VERSION) $(DESTDIR)/usr/bin/gogios

.PHONY: package
package:
	./scripts/build.py --package --platform=all --arch=all

.PHONY: package-release
package-release:
	./scripts/build.py --release --package --platform=all --arch=all

.PHONY: build
build:
	go build -v ${LDFLAGS} -o bin/gogios-$(VERSION) ${MOD} ./cmd/gogios
