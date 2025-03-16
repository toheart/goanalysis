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
	dlv debug --headless --listen :8082 --api-version 2 --accept-multiclient . -- server

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: sync-frontend
# 同步前端代码
sync-frontend:
	@echo "Starting to sync frontend code..."
	@if [ ! -d "web" ]; then \
		mkdir -p web; \
	fi;
	@echo "Getting the latest release version of the frontend..."
	@LATEST_RELEASE=$$(curl -s https://api.github.com/repos/toheart/goanalysis-web/releases/latest | grep "tag_name" | cut -d '"' -f 4); \
	if [ -z "$$LATEST_RELEASE" ]; then \
		echo "Failed to get the latest release version of the frontend, exiting with error" \
		exit 1; \
	else \
		echo "Obtained the latest release version of the frontend: $$LATEST_RELEASE"; \
		DOWNLOAD_URL=$$(curl -s https://api.github.com/repos/toheart/goanalysis-web/releases/latest | grep "browser_download_url.*zip" | cut -d '"' -f 4); \
		if [ -z "$$DOWNLOAD_URL" ]; then \
			 echo "download source code failed, exiting with error";   \
			 exit 1; \
		else \
			echo "Downloading release package: $$DOWNLOAD_URL"; \
			curl -sL "$$DOWNLOAD_URL" -o dist.zip; \
		fi; \
		rm -rf dist_temp ;\
		unzip -q dist.zip -d dist_temp &&  cp -r dist_temp/* web/;\
		rm -rf dist_temp dist.zip; \
	fi;
	@echo "Frontend code sync completed."

.PHONY: package-linux
# 打包Linux版本
package-linux: sync-frontend
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" -o goanalysis ./
	mkdir -p release
	tar -czvf release/goanalysis-linux-$(VERSION).tar.gz ./goanalysis ./configs ./web

.PHONY: package-windows
# 打包Windows版本
package-windows: sync-frontend
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" -o goanalysis.exe ./
	mkdir -p release
	tar -czvf release/goanalysis-windows-$(VERSION).tar.gz ./goanalysis.exe ./configs ./web

.PHONY: docker
# 构建Docker镜像
docker:
	@FRONTEND_VERSION=$$(curl -s https://api.github.com/repos/toheart/goanalysis-web/releases/latest | grep "tag_name" | cut -d '"' -f 4 || echo "unknown"); \
	echo "使用前端版本: $$FRONTEND_VERSION"; \
	docker build -t ghcr.io/toheart/goanalysis:$(VERSION) \
		--build-arg VERSION=$(VERSION) \
		--build-arg FRONTEND_VERSION=$$FRONTEND_VERSION .

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

wire:
	cd ./cmd && wire && cd ..

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
