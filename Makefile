# TODO: Refactor this to make it more DRY

ROOTDIR := $(shell pwd)
VERSION = $(shell cat VERSION)
BUILD_DATE = $(shell date -u '+%s')
GIT_HASH = $(shell git rev-parse --short HEAD)
VERSION_FLAG=-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GitHash=$(GIT_HASH)

.PHONY: generated-code
generated-code:
	go generate -mod=vendor ./... && gofmt -w pkg && goimports -w pkg

SOURCES=$(shell find . -name "*.go" | grep -v test)

#
# ADMIN
#
vrddt-admin: $(SOURCES)
	go build -mod=vendor -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/admin

Dockerfile.vrddt-admin: cmd/Dockerfile
	cp $< $@

vrddt-admin-docker: ./Dockerfile.vrddt-admin
	docker build \
		-t johnwyles/vrddt-admin:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg COMMAND_PATH="./cmd/admin" \
		--build-arg COMMAND_SUFFIX="admin" \
		-f Dockerfile.vrddt-admin .
	rm Dockerfile.vrddt-admin

vrddt-admin-docker-run:
	docker run -mod=vendor johnwyles/vrddt-admin:$(VERSION)

#
# API
#
vrddt-api: $(SOURCES)
	go build -mod=vendor -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/api

Dockerfile.vrddt-api: cmd/Dockerfile
	cp $< $@

vrddt-api-docker: ./Dockerfile.vrddt-api
	docker build \
		-t johnwyles/vrddt-api:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg COMMAND_PATH="./cmd/api" \
		--build-arg COMMAND_SUFFIX="api" \
		-f Dockerfile.vrddt-api .
	rm Dockerfile.vrddt-api

vrddt-api-docker-run:
	docker run -mod=vendor johnwyles/vrddt-api:$(VERSION)

#
# CLI
#
vrddt-cli: $(SOURCES)
	go build -mod=vendor -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/cli

Dockerfile.vrddt-cli: cmd/Dockerfile
	cp $< $@

vrddt-cli-docker: ./Dockerfile.vrddt-cli
	docker build \
		-t johnwyles/vrddt-cli:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg COMMAND_PATH="./cmd/cli" \
		--build-arg COMMAND_SUFFIX="cli" \
		-f Dockerfile.vrddt-cli .
	rm Dockerfile.vrddt-cli

vrddt-cli-docker-run:
	docker run -mod=vendor johnwyles/vrddt-cli:$(VERSION)

#
# WEB
#
vrddt-web: $(SOURCES)
	go build -mod=vendor -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/web

Dockerfile.vrddt-web: cmd/Dockerfile
	cp $< $@

vrddt-web-docker: ./Dockerfile.vrddt-web
	docker build \
		-t johnwyles/vrddt-web:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg COMMAND_PATH="./cmd/web" \
		--build-arg COMMAND_SUFFIX="web" \
		-f Dockerfile.vrddt-web .
	rm Dockerfile.vrddt-web

vrddt-web-docker-run:
	docker run -mod=vendor johnwyles/vrddt-web:$(VERSION)

#
# WORKER
#
vrddt-worker: $(SOURCES)
	go build -mod=vendor -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/worker

Dockerfile.vrddt-worker: cmd/Dockerfile
	cp $< $@

vrddt-worker-docker: ./Dockerfile.vrddt-worker
	docker build \
		-t johnwyles/vrddt-worker:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg COMMAND_PATH="./cmd/worker" \
		--build-arg COMMAND_SUFFIX="worker" \
		-f Dockerfile.vrddt-worker .
	rm Dockerfile.vrddt-worker

vrddt-worker-docker-run:
	docker run -mod=vendor johnwyles/vrddt-worker:$(VERSION)

clean:
	rm -f Dockerfile.*
	rm -rf vendor/