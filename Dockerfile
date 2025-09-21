# ---------- Build stage ----------
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache build-base git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# 换 sqlite 再打开
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/api

# ---------- Run stage ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/server /app/server
USER app

# 端口由配置/环境变量 SSP_SERVER_PORT 控制；默认 8080
EXPOSE 8080

# 用健康检查盯 /api/v1/healthz
HEALTHCHECK --interval=15s --timeout=3s --retries=3 --start-period=10s \
  CMD curl -fsS http://127.0.0.1:${SSP_SERVER_PORT:-8080}/api/v1/healthz >/dev/null || exit 1

ENTRYPOINT ["/app/server"]