
SHELL := /bin/bash
PLATFORM := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
PKG := $(shell go list ./ | grep -v /vendor)
PLUGINS := $(shell for plugin in plugins/*; do go list ./$$plugin/; done)
VERSION := 1.0
INSTALL := /usr/bin/install

all: build test

test: build	
	go test $(PKG)
	go test $(PLUGINS)

golangci:
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

lint: golangci
	golangci-lint run ./
	for p in plugins/*; do golangci-lint run $$p; done

install: build
	useradd --system --no-create-home --shell /sbin/nologin gogios
	$(INSTALL) -d -o gogios -g gogios -m 644 $(DESTDIR)/var/log/gingertechnology
	$(INSTALL) -d -o gogios -g gogios -m 664 $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -d -o gogios -g gogios -m 774 $(DESTDIR)/usr/lib/gingertechengine/plugins
	$(INSTALL) -d -o gogios -g gogios -m 764 $(DESTDIR)/opt/gingertechengine
	$(INSTALL) -o gogios -g gogios -m 774 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	$(INSTALL) -o gogios -g gogios -m 764 web/* $(DESTDIR)/opt/gingertechengine
	$(INSTALL) -o gogios -g gogios -m 664 checkengine/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -o root -g root -m 644 checkengine/gogios.service /usr/lib/systemd/system
	$(INSTALL) -o root -g root -m 755 bin/gogios-$(VERSION)-$(GOOS) $(DESTDIR)/usr/bin

build: lint
	mkdir -p bin/plugins
	GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -o bin/gogios-$(VERSION)-$(GOOS)
	for p in ./plugins/*; do GOOS=$(os) GOARCH=amd64 go build -o bin/$$p ./$$p; done

.PHONY: test lint build install