# Keychain

Keychain 是一个带后台管理页面的 LLM API Key 分发与审计系统。系统使用 Go + SQLite + WAL，集中管理多个 provider、model、key、渠道、用户和权限，并为调用系统提供一次性申请 key、失败上报和调用历史查询能力。

## 文档

- [原始需求](./原始需求.md)
- [需求规格](./docs/需求规格.md)
- [API 设计](./docs/API设计.md)
- [数据库设计](./docs/数据库设计.md)
- [开发计划](./docs/开发计划.md)
- [架构决策记录](./docs/decisions/ADR-001-使用Go和SQLite-WAL.md)

## 核心能力

- 管理多个 provider，每个 provider 可配置多个 model 和多个 key。
- 每个 key 支持别名、启用状态、可用状态、排序和失败统计。
- 每个 provider 可配置 key 分发策略：轮询分发或优先使用第一个可用 key。
- 管理渠道、用户，以及 provider + model 粒度的权限。
- 支持渠道默认权限、用户显式权限和批量权限设置。
- 调用系统申请一次调用时返回一个 key，并记录完整分发历史。
- 调用系统可上报调用失败，系统记录失败并可据此标记 key 不可用。
- 后台管理系统使用 admin 登录，密码从 `.env` 读取。

## 技术约束

- 后端：Go
- 数据库：SQLite，开启 WAL
- 登录：单管理员账号 `admin`，密码和 session secret 从 `.env` 读取
- 存储：key 明文存入数据库，后台列表中星号展示，历史中显示 key 别名

## 本地开发

```powershell
Copy-Item .env.example .env
go test ./...
go run ./cmd/server
```

服务默认监听 `:8080`，健康检查地址为 `GET /api/health`。可以在 `.env` 中设置 `HTTP_ADDR` 覆盖监听地址，例如 `HTTP_ADDR=127.0.0.1:8081`。

## 下一步

建议按以下顺序实现：

1. 初始化 Go 项目、配置加载、SQLite 连接和 WAL 设置。
2. 建立数据库迁移和核心表结构。
3. 实现 admin 登录与 session 中间件。
4. 实现 provider、model、key、channel、user、permission 的管理 API。
5. 实现 key 分发事务和失败上报。
6. 实现后台页面和历史查询。
