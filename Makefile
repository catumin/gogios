export G111MODULE=on
PLATFORM := $(shell uname | tr [:upper:] [:lower:])
ARCH := $(shell uname -m)
PKGS := $(shell go list ./... | grep -v /vendor)
GOCC := $(shell go version)
VERSION := 1.0.2
INSTALL := $(shell which install)

LDFLAGS := -gcflags=all=-trimpath=${PWD} -asmflags=all=-trimpath=${PWD} -ldflags=-extldflags=-zrelro -ldflags=-extldflags=-znow -ldflags '-s -w -X main.version=${VERSION}'
MOD := -mod=vendor
ifneq (,$(findstring gccgo,$(GOCC)))
	export GOPATH=$(shell pwd)/.go
	LDFLAGS := -gccgoflags '-s -w'
	MOD :=
endif

all: lint build test

test: build	
	go test $(PKGS)

lint:
	golangci-lint run ./
	for p in plugins/*; do golangci-lint run $$p; done

install: build
	useradd --system --user-group --home-dir /var/spool/gogios --shell /sbin/nologin gogios
	$(INSTALL) -d -o gogios -g gogios -m 644 $(DESTDIR)/var/log/gingertechnology
	$(INSTALL) -d -o gogios -g gogios -m 664 $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -d -o gogios -g gogios -m 775 $(DESTDIR)/usr/lib/gingertechengine/plugins
	$(INSTALL) -d $(DESTDIR)/usr/bin
	$(INSTALL) -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/ -type d); do $(INSTALL) -d -o gogios -g gogios -m 764 $(DESTDIR)/opt/gingertechengine/$$(echo $$d | cut -d"/" -f2-); done
	$(INSTALL) -o gogios -g gogios -m 775 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	for f in $$(find web/ -type f); do $(INSTALL) -D -o gogios -g gogios --mode 764 "$$f" $(DESTDIR)/opt/gingertechengine/$$(echo $$f | cut -d"/" -f2-); done
	$(INSTALL) -o gogios -g gogios -m 664 checkengine/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -o root -g root -m 644 checkengine/gogios.service $(DESTDIR)/usr/lib/systemd/system
	$(INSTALL) -o root -g root -T -m 755 bin/gogios-$(VERSION)-$(PLATFORM) $(DESTDIR)/usr/bin/gogios
	touch $(DESTDIR)/var/log/gingertechnology/service_check.log
	chown gogios:gogios $(DESTDIR)/var/log/gingertechnology/service_check.log


package: build
	$(INSTALL) -d $(DESTDIR)/var/log/gingertechnology
	$(INSTALL) -d $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -d $(DESTDIR)/usr/lib/gingertechengine/plugins
	$(INSTALL) -d $(DESTDIR)/usr/bin
	$(INSTALL) -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/ -type d); do $(INSTALL) -d $(DESTDIR)/opt/gingertechengine/$$(echo $$d | cut -d"/" -f2-); done
	$(INSTALL) -m 775 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	for f in $$(find web/ -type f); do $(INSTALL) --mode 764 "$$f" $(DESTDIR)/opt/gingertechengine/$$(echo $$f | cut -d"/" -f2-); done
	$(INSTALL) -m 664 checkengine/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	$(INSTALL) -m 644 checkengine/gogios.service $(DESTDIR)/usr/lib/systemd/system
	$(INSTALL) -T -m 755 bin/gogios-$(VERSION)-$(PLATFORM) $(DESTDIR)/usr/bin/gogios
	touch $(DESTDIR)/var/log/gingertechnology/service_check.log

debug: lint
	mkdir -p debug/bin/plugins
	go build -v ${LDFLAGS} -o debug/bin/gogios-$(VERSION)-$(PLATFORM) ${MOD}
	for p in ./plugins/*; do go build -o debug/bin/$$p ./$$p; done

build:
	mkdir -p bin/plugins
	go build -v ${LDFLAGS} -o bin/gogios-$(VERSION)-$(PLATFORM) ${MOD}
	for p in ./plugins/*; do go build -o bin/$$p ./$$p; done

.PHONY: test lint build debug install package
