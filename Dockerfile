FROM golang:1.23 AS builder

COPY . /src
WORKDIR /src

ARG VERSION=dev
RUN GOPROXY=https://proxy.golang.org make build VERSION=${VERSION}

# 前端内容已经通过sync-frontend命令获取并放置在web目录中

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app
# 复制前端内容
COPY web /app/web

COPY configs /app/configs

WORKDIR /app

# 添加版本信息
ARG VERSION=dev
ARG FRONTEND_VERSION=unknown
LABEL org.opencontainers.image.version=${VERSION}
LABEL org.opencontainers.image.description="go-analysis backend: ${VERSION}, frontend: ${FRONTEND_VERSION}"

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
