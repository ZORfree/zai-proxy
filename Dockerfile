FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.Version=$(date +%Y%m%d.%H%M%S)" -o zai-proxy .

FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    openssl-dev \
    curl-dev \
    rust

WORKDIR /app

COPY --from=builder /app/zai-proxy .

ENV TZ=Asia/Shanghai

RUN apk add --no-cache \
    tzdata \
    ca-certificates \
    openssl

# data 目录可通过 docker run -v ./data:/app/data 挂载（包含 proxies.txt 等配置文件）
VOLUME ["/app/data"]

EXPOSE 8000

CMD ["./zai-proxy"]
