VERSION := $(shell git describe --exact-match --tags 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git rev-parse --short HEAD)
DESTDIR :=

LDFLAGS := $(LDFLAGS)-gcflags=all=-trimpath=${PWD} -asmflags=all=-trimpath=${PWD} -X main.commit=$(COMMIT) -X main.branch=$(BRANCH) 
ifdef VERSION
	LDFLAGS += -X main.version=$(VERSION)
endif
MOD := -mod=vendor
export G111MODULE=on
ARCH := $(shell uname -m)
PKGS := $(shell go list ./... | grep -v /vendor)
GOPLUGINS := $(shell go list ./... | grep -v /vendor | grep plugins | awk -F'/' '{print $$4"/"$$5}')
PLUGINS := $(shell echo plugins/*)
PLUGINS := $(shell echo ${GOPLUGINS} ${PLUGINS} | tr ' ' '\n' | sort | uniq -u)
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
	install -d -m 644 $(DESTDIR)/var/log/gingertechnology
	install -d -m 664 $(DESTDIR)/etc/gingertechengine
	install -d -m 775 $(DESTDIR)/usr/lib/gingertechengine/plugins
	install -d $(DESTDIR)/usr/bin
	install -d $(DESTDIR)/usr/lib/systemd/system
	for d in $$(find web/views/ -type d); do install -d -m 764 $(DESTDIR)/opt/gingertechengine/$$(echo $$d | cut -d"/" -f3-); done
	for f in $$(find web/views/ -type f); do install -D --mode 764 "$$f" $(DESTDIR)/opt/gingertechengine/$$(echo $$f | cut -d"/" -f3-); done
	install -d -m 764 $(DESTDIR)/opt/gingertechengine/js/output
	install -T -m 764 package_files/example.json $(DESTDIR)/opt/gingertechengine/js/current.json
	install -m 775 bin/plugins/* $(DESTDIR)/usr/lib/gingertechengine/plugins
	install -m 664 package_files/{example.json,gogios.sample.toml,nginx_example.conf} $(DESTDIR)/etc/gingertechengine
	install -o root -g root -m 644 package_files/gogios.service $(DESTDIR)/usr/lib/systemd/system
	install -o root -g root -T -m 755 scripts/gogios-parse-nmap $(DESTDIR)/usr/bin/gogios-parse-nmap
	install -o root -g root -T -m 755 bin/gogios-$(VERSION) $(DESTDIR)/usr/bin/gogios

.PHONY: package
package:
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
	install -T -m 755 scripts/gogios-parse-nmap $(DESTDIR)/usr/bin/gogios-parse-nmap
	install -T -m 755 bin/gogios-$(VERSION) $(DESTDIR)/usr/bin/gogios

.PHONY: build
build:
	mkdir -p bin/plugins
	go build -v ${LDFLAGS} -o bin/gogios-$(VERSION) ${MOD}
	for p in ${GOPLUGINS}; do go build -o bin/$$p ./$$p; done
	for f in ${PLUGINS}; do cp "$$f"/* bin/plugins; done
