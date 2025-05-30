FROM golang:1.23 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

# 设置环境变量，标识Docker环境
ENV ENVIRONMENT=docker

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./fakery", "-conf", "/data/conf"]
