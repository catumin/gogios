.PHONY: test lint build install package

DESTDIR :=

VERSION := 1.4
LDFLAGS := -gcflags=all=-trimpath=${PWD} -asmflags=all=-trimpath=${PWD} -ldflags=-extldflags=-zrelro -ldflags=-extldflags=-znow -ldflags '-s -w -X main.version=${VERSION}'
MOD := -mod=vendor
export G111MODULE=on
PLATFORM := $(shell uname | tr [:upper:] [:lower:])
ARCH := $(shell uname -m)
PKGS := $(shell go list ./... | grep -v /vendor)
GOPLUGINS := $(shell go list ./... | grep -v /vendor | grep plugins | awk -F'/' '{print $$4"/"$$5}')
PLUGINS := $(shell echo plugins/*)
PLUGINS := $(shell echo ${GOPLUGINS} ${PLUGINS} | tr ' ' '\n' | sort | uniq -u)
GOCC := $(shell go version)


ifneq (,$(findstring gccgo,$(GOCC)))
	export GOPATH=$(shell pwd)/.go
	LDFLAGS := -gccgoflags '-s -w'
	MOD :=
endif

default: build

all: lint test build

test:
	go test $(PKGS)

lint:
	golangci-lint run ./
	for p in plugins/*; do golangci-lint run $$p; done

install: build
	useradd --system --user-group --home-dir /var/spool/gogios --shell /sbin/nologin gogios
	install -d -o gogios -g gogios -m 644 $(DESTDIR)/var/log/gingertechnology
	install -d -o gogios -g gogios -m 664 $(DESTDIR)/etc/gingertechengine
	install -d -o gogios -g gogios -m 775 $(DESTDIR)/usr/lib/gingertechengine/plugins
	install -d $(DESTDIR)/usr/bin
	install -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/views/ -type d); do install -d -o gogios -g gogios -m 764 $(DESTDIR)/opt/gingertechengine/$$(echo $$d | cut -d"/" -f3-); done
	for f in $$(find web/views/ -type f); do install -D -o gogios -g gogios --mode 764 "$$f" $(DESTDIR)/opt/gingertechengine/$$(echo $$f | cut -d"/" -f3-); done
	install -d -o gogios -g gogios -m 764 $(DESTDIR)/opt/gingertechengine/js/output
	install -o gogios -g gogios -T -m 764 package_files/example.json $(DESTDIR)/opt/gingertechengine/js/current.json
	install -o gogios -g gogios -m 775 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	install -o gogios -g gogios -m 664 package_files/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	install -o root -g root -m 644 package_files/gogios.service $(DESTDIR)/usr/lib/systemd/system
	install -o root -g root -T -m 755 bin/gogios-$(VERSION)-$(PLATFORM) $(DESTDIR)/usr/bin/gogios


package: build
	install -d $(DESTDIR)/var/log/gingertechnology
	install -d $(DESTDIR)/etc/gingertechengine
	install -d $(DESTDIR)/usr/lib/gingertechengine/plugins
	install -d $(DESTDIR)/usr/bin
	install -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/views/ -type d); do install -d $(DESTDIR)/opt/gingertechengine/$$(echo $$d | cut -d"/" -f3-); done
	for f in $$(find web/views/ -type f); do install --mode 764 "$$f" $(DESTDIR)/opt/gingertechengine/$$(echo $$f | cut -d"/" -f3-); done
	install -d $(DESTDIR)/opt/gingertechengine/js/output
	touch $(DESTDIR)/opt/gingertechengine/js/output/.keep
	install -m 775 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	install -m 664 package_files/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	install -m 644 package_files/gogios.service $(DESTDIR)/usr/lib/systemd/system
	install -T -m 755 bin/gogios-$(VERSION)-$(PLATFORM) $(DESTDIR)/usr/bin/gogios

build:
	mkdir -p bin/plugins
	go build -v ${LDFLAGS} -o bin/gogios-$(VERSION)-$(PLATFORM) ${MOD}
	for p in ${GOPLUGINS}; do go build -o bin/$$p ./$$p; done
	for f in ${PLUGINS}; do cp "$$f"/* bin/plugins; done
