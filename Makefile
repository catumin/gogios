
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
	useradd --system --no-create-home --user-group --shell /sbin/nologin gogios
	$(INSTALL) -d -o gogios -g gogios -m 644 $(DESTDIR)/var/log/gingertechnology
	$(INSTALL) -d -o gogios -g gogios -m 664 $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -d -o gogios -g gogios -m 774 $(DESTDIR)/usr/lib/gingertechengine/plugins
	$(INSTALL) -d $(DESTDIR)/usr/bin
	$(INSTALL) -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/ -type d); do $(INSTALL) -d -o gogios -g gogios -m 764 $(DESTDIR)/opt/gingertechengine/$$(echo $$d | cut -d"/" -f2-); done
	$(INSTALL) -o gogios -g gogios -m 774 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	for f in $$(find web/ -type f); do $(INSTALL) -D -o gogios -g gogios --mode 764 "$$f" $(DESTDIR)/opt/gingertechengine/$$(echo $$f | cut -d"/" -f2-); done
	$(INSTALL) -o gogios -g gogios -m 664 checkengine/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -o root -g root -m 644 checkengine/gogios.service $(DESTDIR)/usr/lib/systemd/system
	$(INSTALL) -o root -g root -T -m 755 bin/gogios-$(VERSION)-$(PLATFORM) $(DESTDIR)/usr/bin/gogios

package: build
	$(INSTALL) -d $(DESTDIR)/var/log/gingertechnology
	$(INSTALL) -d $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -d $(DESTDIR)/usr/lib/gingertechengine/plugins
	$(INSTALL) -d $(DESTDIR)/usr/bin
	$(INSTALL) -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/ -type d); do $(INSTALL) -d $(DESTDIR)/opt/gingertechengine/$$(echo $$d | cut -d"/" -f2-); done
	$(INSTALL) -m 774 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	for f in $$(find web/ -type f); do $(INSTALL) --mode 764 "$$f" $(DESTDIR)/opt/gingertechengine/$$(echo $$f | cut -d"/" -f2-); done
	$(INSTALL) -m 664 checkengine/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -m 644 checkengine/gogios.service $(DESTDIR)/usr/lib/systemd/system
	$(INSTALL) -T -m 755 bin/gogios-$(VERSION)-$(PLATFORM) $(DESTDIR)/usr/bin/gogios

debug: lint
	mkdir -p debug/bin/plugins
	GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -o debug/bin/gogios-$(VERSION)-$(PLATFORM)
	for p in ./plugins/*; do GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -o debug/bin/$$p ./$$p; done

build:
	mkdir -p bin/plugins
	GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -o bin/gogios-$(VERSION)-$(PLATFORM)
	for p in ./plugins/*; do GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -o bin/$$p ./$$p; done

.PHONY: test lint build debug install package