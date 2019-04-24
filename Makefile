# TODO: Refactor this to make it more DRY

ROOTDIR := $(shell pwd)
VERSION = $(shell cat VERSION)
BUILD_DATE = $(shell date -u '+%s')
GIT_HASH = $(shell git rev-parse --short HEAD)
VERSION_FLAG=-X main.Version=$(VERSION) -X main.BuildTimestamp=$(BUILD_DATE) -X main.GitHash=$(GIT_HASH)

.PHONY: generated-code
generated-code:
	go generate -mod=vendor ./... && gofmt -w pkg && goimports -w pkg

SOURCES=$(shell find . -name "*.go" | grep -v test)

all: all-cli all-docker

all-cli: vrddt-admin vrddt-api vrddt-cli vrddt-web vrddt-worker

all-docker: vrddt-admin-docker vrddt-api-docker vrddt-cli-docker vrddt-web-docker vrddt-worker-docker

#
# ADMIN
#
vrddt-admin: $(SOURCES)
	go build -mod=vendor -a -installsuffix cgo -ldflags "-extldflags \"-static\"" -ldflags -v -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/admin

Dockerfile.admin: cmd/Dockerfile.generic
	cp $< $@

vrddt-admin-docker: ./Dockerfile.admin
	docker build \
		-t johnwyles/vrddt-admin:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg VRDDT_COMMAND_PATH="./cmd/admin" \
		--build-arg VRDDT_COMMAND="admin" \
		-f ./Dockerfile.admin .
	rm Dockerfile.admin

vrddt-admin-docker-run:
	docker run -mod=vendor johnwyles/vrddt-admin:$(VERSION)

#
# API
#
vrddt-api: $(SOURCES)
	go build -mod=vendor -a -installsuffix cgo -ldflags "-extldflags \"-static\"" -ldflags -v -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/api

# Dockerfile.api: cmd/Dockerfile.generic
# 	cp $< $@

vrddt-api-docker: #./Dockerfile.api
	docker build \
		-t johnwyles/vrddt-api:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg VRDDT_COMMAND_PATH="./cmd/api" \
		--build-arg VRDDT_COMMAND="api" \
		-f cmd/Dockerfile.api .
	#rm Dockerfile.api

vrddt-api-docker-run:
	docker run -mod=vendor johnwyles/vrddt-api:$(VERSION)

#
# CLI
#
vrddt-cli: $(SOURCES)
	go build -mod=vendor -a -installsuffix cgo -ldflags "-extldflags \"-static\"" -ldflags -v -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/cli

Dockerfile.cli: cmd/Dockerfile.generic
	cp $< $@

vrddt-cli-docker: ./Dockerfile.cli
	docker build \
		-t johnwyles/vrddt-cli:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg VRDDT_COMMAND_PATH="./cmd/cli" \
		--build-arg VRDDT_COMMAND="cli" \
		-f ./Dockerfile.cli .
	rm Dockerfile.cli

vrddt-cli-docker-run:
	docker run -mod=vendor johnwyles/vrddt-cli:$(VERSION)

#
# WEB
#
vrddt-web: $(SOURCES)
	go build -mod=vendor -a -installsuffix cgo -ldflags "-extldflags \"-static\"" -ldflags -v -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/web

# Dockerfile.vrddt-web: cmd/Dockerfile
# 	cp $< $@

vrddt-web-docker: #./Dockerfile.vrddt-web
	docker build \
		-t johnwyles/vrddt-web:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg VRDDT_COMMAND_PATH="./cmd/web" \
		--build-arg VRDDT_COMMAND="web" \
		-f cmd/Dockerfile.web .
	# rm Dockerfile.vrddt-web

vrddt-web-docker-run:
	docker run -mod=vendor johnwyles/vrddt-web:$(VERSION)

#
# WORKER
#
vrddt-worker: $(SOURCES)
	go build -mod=vendor -a -installsuffix cgo -ldflags "-extldflags \"-static\"" -ldflags -v -ldflags "$(VERSION_FLAG)" -o $@ ./cmd/worker

# Dockerfile.worker: cmd/Dockerfile
# 	cp $< $@

vrddt-worker-docker: #./Dockerfile.worker
	docker build \
		-t johnwyles/vrddt-worker:$(VERSION) \
		--build-arg VERSION_FLAG="$(VERSION_FLAG)" \
		--build-arg VRDDT_COMMAND_PATH="./cmd/worker" \
		--build-arg VRDDT_COMMAND="worker" \
		-f cmd/Dockerfile.worker .
	# rm Dockerfile.vrddt-worker

vrddt-worker-docker-run:
	docker run -mod=vendor johnwyles/vrddt-worker:$(VERSION)

clean:
	rm -f Dockerfile.*
	rm -rf vendor/