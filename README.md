# Student Services Platform - Backend

本项目是学生服务平台的 Go 语言后端，使用 Gin 和 GORM 构建。

-   **语言**: Go
-   **框架**: Gin
-   **ORM**: GORM
-   **数据库**: PostgreSQL
-   **部署**: Docker

---

## 🐳 Docker 部署 (测试环境)

1.  **创建配置文件**:
    > **注意**: 这些文件不应提交到 Git。

`.env.staging`

    ```dotenv
    # .env.staging
    POSTGRES_PASSWORD=YourStrongPasswordHere
    JWT_SECRET=YourJWTSECRETHere
    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=JHWL2025-8
    POSTGRES_DB=ssp
    SSP_DATABASE_DSN="postgres://postgres:YourStrongPasswordHere@db:5432/ssp?sslmode=disable"
    ALPINE_MIRROR=https://mirrors.tuna.tsinghua.edu.cn
    ```

`.env.dev`

    ```dotenv
    # .env.dev
    # Local Development Database Config
    POSTGRES_USER=postgres
    POSTGRES_PASSWORD=YourStrongPasswordHere
    POSTGRES_DB=ssp
    JWT_SECRET=YourStrongPasswordHere
    ALPINE_MIRROR=https://mirrors.tuna.tsinghua.edu.cn
    ```

`config.yaml` (not applicable in the staging environment)

2.  **使用 Taskfile 管理服务**:

## 任务列表 (tasks)

### 测试环境管理 (Staging)

这组命令使用 `docker-compose.staging.yml` 和 `.env.staging` 文件来管理测试环境。

#### `docker:up`

> 启动测试服（构建镜像并后台运行）

-   **运行命令**: `task docker:up`
-   **执行内容**: `docker compose -f docker-compose.staging.yml --env-file .env.staging up -d --build`
-   **说明**: 强制重新构建镜像，并在后台（`-d`）启动所有服务。

#### `docker:down`

> 停止测试服（保留数据库卷）

-   **运行命令**: `task docker:down`
-   **执行内容**: `docker compose -f docker-compose.staging.yml --env-file .env.staging down`
-   **说明**: 停止并移除所有容器，但保留与数据库等相关的卷（Volume），以便下次启动时数据不丢失。

#### `docker:logs`

> 查看 API 日志

-   **运行命令**: `task docker:logs`
-   **执行内容**: `docker compose -f docker-compose.staging.yml --env-file .env.staging logs -f api`
-   **说明**: 实时跟踪（`-f`）并显示名为 `api` 服务的日志输出。

#### `docker:ps`

> 查看容器状态

-   **运行命令**: `task docker:ps`
-   **执行内容**: `docker compose -f docker-compose.staging.yml --env-file .env.staging ps`
-   **说明**: 列出当前测试环境下所有容器的运行状态。

#### `docker:purge`

> 停止测试服并彻底删除所有卷（清空数据库）

-   **运行命令**: `task docker:purge`
-   **执行内容**: `docker compose -f docker-compose.staging.yml --env-file .env.staging down --volumes`
-   **说明**: 彻底清理测试环境，不仅停止并移除容器，还会删除所有关联的卷（`--volumes`），**此操作会导致数据库等持久化数据被清空**。

---

### 开发环境管理 (Development with Air)

这组命令使用 `docker-compose.dev.yml` 和 `.env.dev` 文件来管理开发环境，通常集成了 [Air](https://github.com/cosmtrek/air) 工具以实现代码热重载。

#### `air:up`

> 启动 Air 开发服（构建镜像并后台运行）

-   **运行命令**: `task air:up`
-   **执行内容**: `docker compose -f docker-compose.dev.yml --env-file .env.dev up -d --build`
-   **说明**: 启动开发环境，当代码文件发生变化时，服务会自动重新编译和运行。

#### `air:down`

> 停止 Air 开发服（保留数据库卷）

-   **运行命令**: `task air:down`
-   **执行内容**: `docker compose -f docker-compose.dev.yml --env-file .env.dev down`
-   **说明**: 停止开发环境的容器，并保留数据卷。

#### `air:logs`

> 查看 API 日志

-   **运行命令**: `task air:logs`
-   **执行内容**: `docker compose -f docker-compose.dev.yml --env-file .env.dev logs -f api`
-   **说明**: 实时查看开发环境中 `api` 服务的日志，方便调试。

#### `air:ps`

> 查看容器状态

-   **运行命令**: `task air:ps`
-   **执行内容**: `docker compose -f docker-compose.dev.yml --env-file .env.dev ps`
-   **说明**: 列出当前开发环境下所有容器的运行状态。

#### `air:purge`

> 停止 Air 开发服并彻底删除所有卷（清空数据库）

-   **运行命令**: `task air:purge`
-   **执行内容**: `docker compose -f docker-compose.dev.yml --env-file .env.dev down --volumes`
-   **说明**: 彻底清理开发环境，包括容器和所有数据卷。

---

## 📚 API 文档

API 规范定义在 `internal/openapi/学生服务平台 API.openapi.json`。