FROM golang:1.18.7-alpine3.15 as builder
RUN #mkdir /src
#RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
#RUN apk add build-base
#ADD . /src
#WORKDIR /src
WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum
# 坑：
# 报错 go mod download: google.golang.org/api@v0.44.0: read tcp 172.17.0.3:60862->14.204.51.154:443: read: connection reset by peer
# The command '/bin/sh -c go mod download' returned a non-zero code: 1
# make: *** [docker-build] 错误 1
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
# # 需要把该放入的都copy进去，如果报出 package xxxxx is not in GOROOT  => 就是这个问题。
COPY main.go main.go
COPY cmd/ cmd/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o container-agent main.go
RUN  chmod +x container-agent


FROM alpine:3.12
#ENV ZONEINFO=/app/zoneinfo.zip
RUN mkdir /app
WORKDIR /app

#COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /app

COPY --from=builder /app/container-agent .
ENTRYPOINT  ["./container-agent"]
