
SHELL := /bin/bash
PLATFORM := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)
GOPATH := $(shell go env GOPATH)
GOBIN := $(GOPATH)/bin
PKG := $(shell go list ./ | grep -v /vendor)
PLUGINS := $(shell for plugin in plugins/*; do go list ./$$plugin/; done)
VERSION := 1.0

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
	install -d -o gogios -g gogios -m 644 /var/log/gingertechnology
	install -d -o gogios -g gogios -m 664 /etc/gingertechengine
	install -d -o gogios -g gogios -m 774 /usr/lib/gingertechengine/plugins
	install -d -o gogios -g gogios -m 764 /opt/gingertechengine
	install -o gogios -g gogios -m 774 bin/plugins/* /usr/lib/gingertechengine/plugins
	install -o gogios -g gogios -m 764 web/* /opt/gingertechengine
	install -o gogios -g gogios -m 664 checkengine/{example.json,gogios.sample.toml,nginx_example.conf} /etc/gingertechengine
	install -o root -g root -m 644 checkengine/gogios.service /usr/lib/systemd/system
	install -o root -g root -m 755 bin/gogios-$(VERSION)-$(GOOS) /usr/bin

build: lint
	mkdir -p bin/plugins
	GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -o bin/gogios-$(VERSION)-$(GOOS)
	for p in ./plugins/*; do GOOS=$(os) GOARCH=amd64 go build -o bin/$$p ./$$p; done

.PHONY: test lint build install