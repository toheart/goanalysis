FROM golang:1.23 AS builder

COPY . /src
WORKDIR /src

ARG VERSION=dev
RUN GOPROXY=https://goproxy.cn make build VERSION=${VERSION}

# 前端构建阶段（如果前端需要构建）
# FROM node:18 AS frontend
# WORKDIR /app
# COPY frontweb/ .
# RUN npm install && npm run build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app
# 如果前端已经构建好，直接复制
COPY frontweb/dist /app/frontweb/dist
# 或者如果前端需要构建
# COPY --from=frontend /app/dist /app/frontweb/dist

COPY configs /app/configs

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./server", "-conf", "/data/conf"]
