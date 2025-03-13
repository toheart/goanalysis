GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION?=$(shell git describe --tags --always || echo "dev")

ifeq ($(GOHOSTOS), windows)
	Git_Bash="$(subst \,/,$(subst cmd\git.exe,bin\bash.exe,$(shell where git)))"
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
endif

.PHONY: init
# init env
init:
	go install github.com/spf13/cobra-cli@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --grpc-gateway_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

debug:
	dlv debug --headless --listen :2346 --api-version 2 --accept-multiclient ./cmd -- server

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: sync-frontend
# 同步前端代码
sync-frontend:
	@if [ ! -d "frontweb" ]; then \
		git clone https://github.com/toheart/goanalysis-web.git frontweb; \
	else \
		cd frontweb && git pull; \
	fi
	@echo "前端代码同步完成"

.PHONY: package-linux
# 打包Linux版本
package-linux: sync-frontend
	go build -ldflags "-X main.Version=$(VERSION)" -o goanalysis ./cmd
	mkdir -p release
	tar -czvf release/goanalysis-linux-$(VERSION).tar.gz ./goanalysis ./configs ./frontweb/dist

.PHONY: package-windows
# 打包Windows版本
package-windows: sync-frontend
	go build -ldflags "-X main.Version=$(VERSION)" -o goanalysis.exe ./cmd
	mkdir -p release
	tar -czvf release/goanalysis-windows-$(VERSION).tar.gz ./goanalysis.exe ./configs ./frontweb/dist

.PHONY: docker
# 构建Docker镜像
docker:
	docker build -t ghcr.io/toheart/goanalysis:$(VERSION) --build-arg VERSION=$(VERSION) .

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all
all:
	make api;
	make config;
	make generate;

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
