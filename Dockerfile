FROM golang:alpine AS builder2
ENV GO111MODULE=on CGO_ENABLED=0

ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64}

WORKDIR /build

ADD go.mod go.sum ./
RUN go mod download

COPY . .
# 复制本地已经编译好的前端静态文件
COPY ./web/dist ./web/dist

RUN go build -ldflags "-s -w -X 'github.com/QuantumNous/new-api/common.Version=$(cat VERSION)'" -o new-api

FROM alpine

RUN apk upgrade --no-cache \
    && apk add --no-cache ca-certificates tzdata \
    && update-ca-certificates

# 复制构建好的二进制文件
COPY --from=builder2 /build/new-api /
# 同时复制前端静态文件到最终镜像（如果需要的话）
COPY --from=builder2 /build/web/dist ./web/dist

EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/new-api"]