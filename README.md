# Student Services Platform - Backend

本项目是学生服务平台的 Go 语言后端，使用 Gin 框架和 GORM。

## 核心技术栈

- **语言**: Go
- **Web 框架**: [Gin](https://gin-gonic.com/)
- **ORM**: [GORM](https://gorm.io/)
- **数据库**: PostgreSQL
- **配置管理**: [Viper](https://github.com/spf13/viper)
- **开发工具**:
  - [Air](https://github.com/cosmtrek/air) (用于本地开发热重载)
  - [Taskfile](https://taskfile.dev/) (用于简化命令行任务)
- **部署**: Docker & Docker Compose

---

## 🚀 快速开始 (本地开发)

### 1. 环境准备

确保你已经安装了以下工具：

- **Go** (版本 1.21+)
- **Docker** & **Docker Compose**
- **Taskfile**: `go install github.com/go-task/task/v3/cmd/task@latest`
- **Air**: `go install github.com/cosmtrek/air@latest`

### 2. 项目配置

项目配置通过 `config/config.yaml` 和环境变量加载。

- 复制示例配置文件：
  ```bash
  cp config/config.example.yaml config/config.yaml
  ```
- **默认配置已为你设置好本地开发环境**，它会尝试连接本地 `localhost:5432` 的 PostgreSQL 数据库。

### 3. 启动本地数据库

为了方便开发，我们使用 Docker 启动一个 PostgreSQL 实例。

```bash
docker run --name ssp-db-local -e POSTGRES_DB=ssp -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -p 127.0.0.1:5432:5432 -d postgres:16-alpine
```

- 这个命令创建的数据库连接信息与 `config/config.yaml` 中的默认值完全匹配。
- 当你不需要时，可以停止并移除它：`docker stop ssp-db-local && docker rm ssp-db-local`

### 4. 运行应用

在项目根目录执行：

```bash
air
```
Air 会监控 `*.go` 文件的变动，自动重新编译和运行你的应用。服务将启动在 `http://localhost:8080`。

> **备选方案**: 你也可以使用 `Taskfile` 来启动，但这不会有热重载功能。
> ```bash
> task run
> ```

---

## 🐳 Docker 部署 (测试/Staging 环境)

我们使用 `docker-compose.staging.yml` 来管理测试服务器的部署。

### 1. 部署配置

部署配置通过 `.env.staging` 文件注入。**这个文件不应提交到 Git**。

- 在服务器的项目根目录下，创建一个 `.env.staging` 文件，内容如下：
  ```dotenv
  # 设置一个强密码
  POSTGRES_PASSWORD=YourStrongPasswordHere
  ```

### 2. 管理服务

我们已经在 `Taskfile.yml` 中集成了常用的 Docker Compose 命令。

- **构建并启动服务 (后台模式)**:
  ```bash
  task docker:up
  ```
- **停止服务 (会保留数据库数据)**:
  ```bash
  task docker:down
  ```
- **查看 API 服务日志**:
  ```bash
  task docker:logs
  ```
- **查看容器运行状态**:
  ```bash
  task docker:ps
  ```

---

## 📚 API 与数据库

### API 文档

本项目的 API 遵循 OpenAPI 3.0 规范，定义文件位于：
`internal/openapi/学生服务平台 API.openapi.json`

- **推荐查看方式**:
  - 将文件内容粘贴到 [Swagger Editor](https://editor.swagger.io/)。
  - 使用 Postman 或 Insomnia 等工具导入该文件来调试 API。

### 数据库 Schema

- **代码即文档**: 数据库的表结构由 GORM 模型定义，是唯一的事实来源。
  - 查看 `internal/db/models.go` 来了解所有表和字段。
- **自动迁移**: 应用在每次启动时会自动执行 `AutoMigrate`，确保数据库表结构与最新的模型代码保持一致。

---

## ✅ Taskfile 命令速查

- `task run`: 在本地直接运行应用。
- `task build`: 构建生产环境的二进制文件到 `bin/` 目录。
- `task test`: 运行所有单元测试。
- `task docker:up`: 构建并启动 Docker 测试环境。
- `task docker:down`: 停止 Docker 测试环境。
- `task docker:logs`: 实时查看 Docker 中 API 服务的日志。
- `task docker:ps`: 显示 Docker 测试环境中各容器的状态。