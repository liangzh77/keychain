# API 设计

> 说明：本文是项目内部 API 总设计，包含后台管理 API 和调用系统 Runtime API。
> 给外部项目对接时，请优先发送 `docs/外部调用接入指南.md`，不要直接发送本文档。

## 通用约定

所有 API 返回 JSON。

错误响应统一使用：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": {}
  }
}
```

常用状态码：

- `200`：成功。
- `201`：创建成功。
- `400`：请求格式错误。
- `401`：未登录或未认证。
- `403`：无权限。
- `404`：资源不存在。
- `409`：资源冲突。
- `422`：语义校验失败。
- `500`：服务端错误。

列表接口应支持分页：

```text
page=1&pageSize=20
```

## 健康检查 API

### GET /healthz

公开部署探活接口。不要求管理员登录。部署 hook、systemd 或反向代理应优先使用这个接口。

响应：

```json
{
  "ok": true
}
```

### GET /api/health

健康检查接口。不要求管理员登录，用于部署平台、反向代理或监控系统探活。

响应：

```json
{
  "status": "ok",
  "database": "ok",
  "time": "2026-05-06T00:00:00Z"
}
```

## 后台认证 API

### POST /api/auth/login

管理员登录。

请求：

```json
{
  "username": "admin",
  "password": "password-from-env"
}
```

响应：

```json
{
  "username": "admin"
}
```

### POST /api/auth/logout

退出登录。

### GET /api/auth/me

查询当前登录管理员。

响应：

```json
{
  "username": "admin"
}
```

## 后台管理 API

后台管理 API 必须要求管理员登录。

### 字段展示约定

后台管理界面以“名称”为主要识别信息，不要求管理员填写 `code`。

- `provider.code`、`model.code`、`channel.code` 是内部字段，可由服务端根据名称自动生成。
- 后台创建和更新 provider、model、channel 时，客户端只需要提交 `name`。
- 后台用户只展示一个名称字段，API 统一使用 `name`。
- Runtime API 使用稳定的 `id` 字段，不接受 `*Code` 字段。

### Providers

```http
GET    /api/providers
POST   /api/providers
GET    /api/providers/:id
PATCH  /api/providers/:id
DELETE /api/providers/:id
```

Provider 字段：

```json
{
  "id": "provider_001",
  "name": "OpenAI",
  "isEnabled": true,
  "rotationStrategy": "ROUND_ROBIN",
  "createdAt": "2026-05-06T00:00:00Z",
  "updatedAt": "2026-05-06T00:00:00Z"
}
```

创建或更新 provider：

```json
{
  "name": "OpenAI",
  "isEnabled": true,
  "rotationStrategy": "ROUND_ROBIN"
}
```

说明：`code` 是内部字段，前端不展示，也不要求管理员填写。

`rotationStrategy` 可选：

- `ROUND_ROBIN`
- `STICKY_FIRST_AVAILABLE`

### Models

```http
GET    /api/models
POST   /api/models
GET    /api/models/:id
PATCH  /api/models/:id
DELETE /api/models/:id
```

Models 列表必须按 provider 过滤。调用方必须传 `providerId`，接口只返回一个 provider 下的 models。

```http
GET /api/models?providerId=provider_001
```

Model 字段：

```json
{
  "id": "model_001",
  "providerId": "provider_001",
  "name": "gpt-4.1",
  "isEnabled": true
}
```

创建或更新 model：

```json
{
  "providerId": "provider_001",
  "name": "gpt-4.1",
  "isEnabled": true
}
```

说明：`code` 是内部字段，前端不展示，也不要求管理员填写。

### Keys

```http
GET    /api/keys
POST   /api/keys
GET    /api/keys/:id
PATCH  /api/keys/:id
DELETE /api/keys/:id
```

列表响应不返回完整 key：

```json
{
  "id": "key_001",
  "providerId": "provider_001",
  "alias": "openai-main-01",
  "maskedValue": "sk-****abcd",
  "isEnabled": true,
  "isAvailable": true,
  "sortOrder": 10,
  "failureCount": 0,
  "lastFailedAt": null
}
```

创建或更新 key 时可提交明文：

```json
{
  "providerId": "provider_001",
  "alias": "openai-main-01",
  "secretValue": "sk-real-key",
  "isEnabled": true,
  "isAvailable": true,
  "sortOrder": 10
}
```

### Channels

```http
GET    /api/channels
POST   /api/channels
GET    /api/channels/:id
PATCH  /api/channels/:id
DELETE /api/channels/:id
```

Channel 字段：

```json
{
  "id": "channel_001",
  "name": "School A",
  "defaultPermissionMode": "DENY",
  "isEnabled": true
}
```

创建或更新 channel：

```json
{
  "name": "School A",
  "defaultPermissionMode": "DENY",
  "isEnabled": true
}
```

说明：`code` 是内部字段，后台不展示。创建时由服务端自动生成。

`defaultPermissionMode` 可选：

- `ALLOW`
- `DENY`

### Users

```http
GET    /api/users
POST   /api/users
GET    /api/users/:id
PATCH  /api/users/:id
DELETE /api/users/:id
```

支持按 channel 过滤：

```http
GET /api/users?channelId=channel_001
```

User 字段：

```json
{
  "id": "user_001",
  "channelId": "channel_001",
  "name": "Student 001",
  "isEnabled": true
}
```

创建或更新 user：

```json
{
  "channelId": "channel_001",
  "name": "Student 001",
  "isEnabled": true
}
```

说明：后台用户 API 只使用 `name` 作为用户名称。

### Permissions

```http
GET    /api/permissions
POST   /api/permissions
PATCH  /api/permissions/:id
DELETE /api/permissions/:id
```

用户显式权限：

```json
{
  "userId": "user_001",
  "providerId": "provider_001",
  "modelId": "model_001",
  "allowed": true
}
```

渠道 provider + model 默认权限：

```http
GET    /api/channel-permission-defaults
POST   /api/channel-permission-defaults
PATCH  /api/channel-permission-defaults/:id
DELETE /api/channel-permission-defaults/:id
```

```json
{
  "channelId": "channel_001",
  "providerId": "provider_001",
  "modelId": "model_001",
  "defaultAllowed": true
}
```

### Bulk permissions

```http
POST /api/bulk/channel-permissions
POST /api/bulk/user-permissions
POST /api/bulk/provider-permissions
```

批量设置用户权限：

```json
{
  "userIds": ["user_001", "user_002"],
  "providerModelPairs": [
    {
      "providerId": "provider_001",
      "modelId": "model_001"
    }
  ],
  "allowed": true
}
```

批量设置渠道默认权限：

```json
{
  "channelIds": ["channel_001", "channel_002"],
  "providerModelPairs": [
    {
      "providerId": "provider_001",
      "modelId": "model_001"
    }
  ],
  "defaultAllowed": false
}
```

批量设置 provider 权限时，后端应展开为 provider 下所有 model 的权限操作。

### Logs

```http
GET /api/dispatch-logs
GET /api/failure-reports
```

调用历史支持过滤：

```http
GET /api/dispatch-logs?startTime=2026-05-01T00:00:00Z&endTime=2026-05-06T23:59:59Z&userId=user_001&channelId=channel_001&providerId=provider_001&modelId=model_001&keyId=key_001&page=1&pageSize=20
```

历史记录响应示例：

```json
{
  "id": "dispatch_001",
  "createdAt": "2026-05-06T00:00:00Z",
  "channelName": "School A",
  "userDisplayName": "Student 001",
  "providerName": "OpenAI",
  "modelName": "gpt-4.1",
  "keyAlias": "openai-main-01",
  "status": "DISPATCHED"
}
```

## 调用系统 API

调用系统 API 不使用后台登录 cookie。第一版可以用内部共享 token 保护，例如请求头：

```http
Authorization: Bearer <RUNTIME_API_TOKEN>
```

该 token 从 `.env` 读取。

### POST /api/runtime/users

提交或更新用户信息。

请求：

```json
{
  "channelId": "channel_001",
  "name": "Student 001"
}
```

响应：

```json
{
  "id": "user_001",
  "channelId": "channel_001",
  "name": "Student 001",
  "isEnabled": true
}
```

### GET /api/runtime/users/:id/permissions

查询用户权限。

响应：

```json
{
  "userId": "user_001",
  "permissions": [
    {
      "providerId": "provider_001",
      "providerName": "OpenAI",
      "modelId": "model_001",
      "modelName": "gpt-4.1",
      "allowed": true
    }
  ]
}
```

### GET /api/runtime/providers

查询可用 providers 列表。

### GET /api/runtime/models

查询可用 models 列表。必须传 `providerId`，接口只返回一个 provider 下的 models。

```http
GET /api/runtime/models?providerId=provider_001
```

### POST /api/runtime/dispatch-key

申请一次调用。

请求：

```json
{
  "channelId": "channel_001",
  "userId": "user_001",
  "providerId": "provider_001",
  "modelId": "model_001"
}
```

响应：

```json
{
  "dispatchLogId": "dispatch_001",
  "providerName": "OpenAI",
  "modelName": "gpt-4.1",
  "keyId": "key_001",
  "keyAlias": "openai-main-01",
  "key": "sk-real-key"
}
```

### POST /api/runtime/key-failures

上报调用失败。

请求：

```json
{
  "dispatchLogId": "dispatch_001",
  "errorCode": "rate_limit",
  "errorMessage": "provider returned 429"
}
```

响应：

```json
{
  "reported": true,
  "keyId": "key_001",
  "keyAlias": "openai-main-01",
  "isAvailable": false
}
```

`isAvailable` 表示失败上报后该 key 当前是否仍被系统认为可用。
