# syntax=docker/dockerfile:1.7

########################
#  Build (builder)     #
########################
# 使用与运行期相同的 Alpine 主次版本，减少跨版本差异
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

# ---- 可调参数（按需覆盖） ----
ARG ALPINE_MIRROR=https://mirrors.aliyun.com        # 可传入: https://mirrors.aliyun.com / https://mirrors.tuna.tsinghua.edu.cn / https://mirrors.cloud.tencent.com
ARG USE_CGO=0                                            # 需要 sqlite 时传 1
ARG GOPROXY=https://proxy.golang.org,direct              # 国内可传 https://goproxy.cn,direct
ARG GONOSUMDB=
ARG GOPRIVATE=
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE
ARG TARGETOS
ARG TARGETARCH

# ---- 基础环境 ----
ENV CGO_ENABLED=$USE_CGO \
    GO111MODULE=on \
    GOPROXY=$GOPROXY \
    GONOSUMDB=$GONOSUMDB \
    GOPRIVATE=$GOPRIVATE

# 设置更近的镜像源 + apk 重试（解决 APKINDEX / 包下载卡住）
RUN --mount=type=cache,target=/var/cache/apk <<'SH'
set -euo pipefail
retry() { for i in 1 2 3 4 5; do "$@" && break || { echo "apk retry $i"; sleep $((i*2)); }; done; }
ver="v$(cut -d. -f1,2 /etc/alpine-release)"   # 例如 v3.20
echo "${ALPINE_MIRROR}/alpine/${ver}/main" > /etc/apk/repositories
echo "${ALPINE_MIRROR}/alpine/${ver}/community" >> /etc/apk/repositories
retry apk update
# CGO=0 时只装 git/证书/时区；CGO=1 再补 build-base 与（示例）sqlite 头文件
retry apk add --no-cache ca-certificates tzdata git
if [ "${CGO_ENABLED}" = "1" ]; then
  retry apk add --no-cache build-base sqlite-dev
fi
update-ca-certificates
SH

WORKDIR /src

# 先仅复制 go.mod/go.sum 以最大化利用缓存
COPY go.mod go.sum ./

# 预下载依赖（缓存：mod + 编译缓存）
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# 再拷贝全部源码
COPY . .

# 编译（可跨架构、裁剪符号表，禁用 VCS 以避免额外 git 查询）
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -buildvcs=false -mod=readonly \
      -ldflags="-s -w -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_DATE}'" \
      -o /out/server ./cmd/api


########################
#  Runtime (release)   #
########################
FROM alpine:3.20 AS release

ARG ALPINE_MIRROR=https://dl-cdn.alpinelinux.org
ARG INSTALL_CURL=1   # 如不想额外装 curl，可传 0，并把健康检查改用 busybox 的 wget

# 同样切换镜像源并加重试，避免运行期安装卡住
RUN --mount=type=cache,target=/var/cache/apk <<'SH'
set -euo pipefail
retry() { for i in 1 2 3 4 5; do "$@" && break || { echo "apk retry $i"; sleep $((i*2)); }; done; }
ver="v$(cut -d. -f1,2 /etc/alpine-release)"
echo "${ALPINE_MIRROR}/alpine/${ver}/main" > /etc/apk/repositories
echo "${ALPINE_MIRROR}/alpine/${ver}/community" >> /etc/apk/repositories
retry apk update
# 仅保留运行必须依赖，体积更小
retry apk add --no-cache ca-certificates tzdata
if [ "${INSTALL_CURL}" = "1" ]; then
  retry apk add --no-cache curl
fi
addgroup -S app && adduser -S -u 10001 -G app app
update-ca-certificates
SH

WORKDIR /app
COPY --from=builder /out/server /app/server
USER app

# 端口由 SSP_SERVER_PORT 控制；默认 8080
EXPOSE 8080

# 健康检查：若未安装 curl，改用 busybox 的 wget（Alpine 自带）
HEALTHCHECK --interval=15s --timeout=3s --retries=3 --start-period=10s \
  CMD sh -c 'URL="http://127.0.0.1:${SSP_SERVER_PORT:-8080}/api/v1/healthz"; \
             if command -v curl >/dev/null 2>&1; then curl -fsS "$URL" >/dev/null; \
             else wget -q -O - "$URL" >/dev/null; fi' || exit 1

ENTRYPOINT ["/app/server"]


########################
#  Dev (air)           #
########################
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS dev

ARG ALPINE_MIRROR=https://dl-cdn.alpinelinux.org
ARG USE_CGO=0
ARG GOPROXY=https://proxy.golang.org,direct

ENV CGO_ENABLED=$USE_CGO \
    GO111MODULE=on \
    GOPROXY=$GOPROXY

# 开发容器：更少步骤，但仍具镜像源/重试与模块缓存
RUN --mount=type=cache,target=/var/cache/apk <<'SH'
set -euo pipefail
retry() { for i in 1 2 3 4 5; do "$@" && break || { echo "apk retry $i"; sleep $((i*2)); }; done; }
ver="v$(cut -d. -f1,2 /etc/alpine-release)"
echo "${ALPINE_MIRROR}/alpine/${ver}/main" > /etc/apk/repositories
echo "${ALPINE_MIRROR}/alpine/${ver}/community" >> /etc/apk/repositories
retry apk update
retry apk add --no-cache ca-certificates tzdata git
if [ "${CGO_ENABLED}" = "1" ]; then
  retry apk add --no-cache build-base sqlite-dev
fi
update-ca-certificates
SH

# 安装 air（固定版本，便于复现）
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go install github.com/air-verse/air@v1.63.0

WORKDIR /app
# 先同步 go.mod/go.sum，加速依赖缓存
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
# 源码实际开发中建议挂载；这里也允许直接拷贝
COPY . .

CMD ["air"]