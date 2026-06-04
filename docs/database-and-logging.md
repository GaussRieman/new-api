# 数据库 Schema 与日志系统

本文档介绍 new-api 项目的数据库表结构设计及日志系统架构。

## 数据库

### 支持的数据库类型

new-api 支持三种数据库，通过 `SQL_DSN` 环境变量切换：

| 数据库 | DSN 格式 | 备注 |
|--------|----------|------|
| SQLite | `local` 或不设置 `SQL_DSN` | 默认选项，适合单机部署 |
| MySQL | 标准 MySQL DSN | 需指定 `parseTime=true` |
| PostgreSQL | `postgres://` 或 `postgresql://` | 支持 |

数据库初始化代码位于 `model/main.go`，使用 GORM 的 `AutoMigrate` 自动创建表结构。

### 核心数据表

#### users - 用户表

用户账户信息，包含身份认证、配额、分组等核心字段。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | integer | 主键 |
| `username` | varchar(20) | 用户名，唯一索引 |
| `password` | varchar | 哈希后的密码 |
| `display_name` | varchar(20) | 显示名称 |
| `role` | int | 角色：1=普通用户，10=管理员，100=超级管理员 |
| `status` | int | 状态：1=启用，其他=禁用 |
| `email` | varchar(50) | 邮箱地址 |
| `quota` | int | 配额余额（内部单位） |
| `used_quota` | int | 已消耗配额 |
| `request_count` | int | 请求次数统计 |
| `group` | varchar(64) | 用户分组，用于渠道路由 |
| `aff_code` | varchar(32) | 邀请码 |
| `inviter_id` | int | 邀请人 ID |
| `access_token` | char(32) | 系统管理令牌（可选） |
| `stripe_customer` | varchar(64) | Stripe 客户 ID |
| `created_at` | bigint | 创建时间戳 |
| `last_login_at` | bigint | 最后登录时间戳 |
| `deleted_at` | timestamp | 软删除时间 |

定义文件：`model/user.go`

#### tokens - API 令牌表

用户创建的 API 访问令牌，用于调用 AI API。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | integer | 主键 |
| `user_id` | int | 所属用户 ID |
| `key` | varchar(128) | 令牌密钥，唯一索引 |
| `name` | varchar | 令牌名称 |
| `status` | int | 状态：1=启用 |
| `created_time` | bigint | 创建时间 |
| `expired_time` | bigint | 过期时间，-1 表示永不过期 |
| `remain_quota` | int | 剩余配额 |
| `unlimited_quota` | bool | 是否无限配额 |
| `used_quota` | int | 已使用配额 |
| `model_limits_enabled` | bool | 是否启用模型限制 |
| `model_limits` | text | 允许的模型列表（JSON） |
| `allow_ips` | varchar | IP 白名单 |
| `group` | varchar | 分组，覆盖用户默认分组 |
| `cross_group_retry` | bool | 是否允许跨分组重试 |
| `deleted_at` | timestamp | 软删除时间 |

定义文件：`model/token.go`

#### channels - 渠道表

上游 AI 提供商的渠道配置，支持多密钥、优先级、模型映射等。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | integer | 主键 |
| `type` | int | 渠道类型（OpenAI=1, Claude=2, Gemini=3 等） |
| `key` | text | API 密钥 |
| `name` | varchar | 渠道名称 |
| `status` | int | 状态：1=启用 |
| `base_url` | varchar | API 基础 URL（自定义端点） |
| `models` | text | 支持的模型列表 |
| `group` | varchar(64) | 分组，用于路由匹配 |
| `weight` | uint | 权重，用于负载均衡 |
| `priority` | bigint | 优先级，数值越大优先级越高 |
| `used_quota` | bigint | 已消耗配额 |
| `model_mapping` | text | 模型映射规则（JSON） |
| `test_model` | varchar | 测试用的模型名称 |
| `response_time` | int | 响应时间（毫秒） |
| `balance` | float | 渠道余额（USD） |
| `auto_ban` | int | 是否自动禁用（失败后） |
| `tag` | varchar | 标签，用于筛选 |
| `setting` | text | 渠道额外设置（JSON） |
| `param_override` | text | 参数覆盖规则（JSON） |
| `header_override` | text | HTTP 头覆盖规则（JSON） |
| `channel_info` | json | 多密钥配置信息 |
| `remark` | varchar(255) | 备注 |

**channel_info JSON 结构**（多密钥模式）：

```json
{
  "is_multi_key": true,
  "multi_key_size": 3,
  "multi_key_status_list": {"0": 1, "1": 1, "2": 2},
  "multi_key_disabled_reason": {"2": "rate_limit"},
  "multi_key_polling_index": 0,
  "multi_key_mode": 0
}
```

定义文件：`model/channel.go`

#### logs - 日志表

API 调用的消费记录，存储元数据而非对话内容。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | integer | 主键 |
| `user_id` | int | 用户 ID |
| `username` | varchar | 用户名 |
| `created_at` | bigint | 创建时间戳 |
| `type` | int | 日志类型 |
| `content` | text | 日志内容摘要 |
| `model_name` | varchar | 模型名称 |
| `token_name` | varchar | 令牌名称 |
| `token_id` | int | 令牌 ID |
| `channel_id` | int | 渠道 ID |
| `prompt_tokens` | int | 输入 token 数 |
| `completion_tokens` | int | 输出 token 数 |
| `quota` | int | 消耗配额 |
| `use_time` | int | 使用时间（秒） |
| `is_stream` | bool | 是否流式请求 |
| `group` | varchar | 分组 |
| `ip` | varchar | 客户端 IP（可选） |
| `request_id` | varchar(64) | 请求唯一标识 |
| `upstream_request_id` | varchar(128) | 上游请求 ID |
| `other` | text | 额外信息（JSON） |

**日志类型常量**：

| 值 | 类型 | 说明 |
|----|------|------|
| 0 | LogTypeUnknown | 未知 |
| 1 | LogTypeTopup | 充值 |
| 2 | LogTypeConsume | 消费（API 调用） |
| 3 | LogTypeManage | 管理操作 |
| 4 | LogTypeSystem | 系统日志 |
| 5 | LogTypeError | 错误 |
| 6 | LogTypeRefund | 退款 |

**重要说明**：`content` 字段不存储实际的对话内容（prompt/response），仅存储消费摘要信息，例如：
- `"模型 gpt-4o"`
- `"Web Search 调用 3 次，调用花费 0.05"`
- `"上游无计费信息"`

定义文件：`model/log.go`

#### 其他重要表

| 表名 | 说明 | 定义文件 |
|------|------|----------|
| `options` | 系统配置选项 | `model/option.go` |
| `topups` | 充值记录 | `model/topup.go` |
| `redemptions` | 兑换码 | `model/redemption.go` |
| `tasks` | 任务队列 | `model/task.go` |
| `midjourney` | Midjourney 任务 | `model/midjourney.go` |
| `subscription_orders` | 订阅订单 | `model/subscription.go` |
| `subscription_plans` | 订阅计划 | `model/subscription.go` |
| `user_subscriptions` | 用户订阅 | `model/subscription.go` |
| `checkins` | 签到记录 | `model/checkin.go` |
| `abilities` | 能力映射 | `model/ability.go` |
| `models` | 模型元数据 | `model/model_meta.go` |
| `vendors` | 提供商元数据 | `model/vendor_meta.go` |
| `prefill_groups` | 预填分组 | `model/prefill_group.go` |
| `passkey_credentials` | Passkey 认证 | `model/passkey.go` |
| `twofas` | 双因素认证 | `model/twofa.go` |
| `custom_oauth_providers` | 自定义 OAuth | `model/custom_oauth_provider.go` |
| `user_oauth_bindings` | OAuth 绑定 | `model/user_oauth_binding.go` |
| `perf_metrics` | 性能指标 | `model/perf_metric.go` |
| `setup` | 系统初始化状态 | `model/setup.go` |

### 数据库兼容性注意事项

项目要求同时兼容 SQLite、MySQL >= 5.7.8、PostgreSQL >= 9.6。代码中需注意：

1. **列名引用差异**：
   - PostgreSQL: `"column"`
   - MySQL/SQLite: `` `column` ``

   使用 `model/main.go` 中定义的 `commonGroupCol`、`commonKeyCol` 变量处理保留字列名。

2. **布尔值差异**：
   - PostgreSQL: `true`/`false`
   - MySQL/SQLite: `1`/`0`

   使用 `commonTrueVal`、`commonFalseVal` 变量。

3. **JSON 存储**：
   - 使用 `TEXT` 类型而非 `JSONB`（PostgreSQL 特有）

4. **列修改限制**：
   - SQLite 不支持 `ALTER COLUMN`，需使用 `ADD COLUMN` 替代方案

---

## 日志系统

### 系统架构

new-api 内置简单的文件日志系统，不集成外部日志服务（如 ELK、Loki）。

#### 日志目录配置

```bash
# 命令行参数
--log-dir ./logs

# 默认值：./logs
```

日志文件命名格式：`oneapi-{时间戳}.log`，例如 `oneapi-20260529083000.log`。

#### 日志级别与函数

| 级别 | 函数 | 输出流 | 说明 |
|------|------|--------|------|
| INFO | `SysLog()`, `LogInfo()` | stdout | 系统信息 |
| WARN | `LogWarn()` | stderr | 警告 |
| ERROR | `SysError()`, `LogError()` | stderr | 错误 |
| DEBUG | `LogDebug()` | stderr | 仅 `DEBUG_ENABLED=true` 时启用 |
| FATAL | `FatalLog()` | stderr | 致命错误，程序退出 |

定义文件：`common/sys_log.go`, `logger/logger.go`

#### 日志轮转

当日志条数超过 100 万条（`maxLogCount = 1000000`）时，自动创建新日志文件。

#### 日志格式示例

```
[INFO] 2026/05/29 - 09:30:00 | SYSTEM | database migration started
[ERR] 2026/05/29 - 09:30:01 | abc123 | failed to connect upstream
[DEBUG] 2026/05/29 - 09:30:02 | abc123 | request body: {...}
```

请求日志会包含 `request_id` 用于追踪。

### Grafana Pyroscope（性能分析）

可选集成，用于性能 profiling，而非日志内容记录。

**环境变量配置**：

```bash
PYROSCOPE_URL=http://pyroscope:4040
PYROSCOPE_APP_NAME=new-api
PYROSCOPE_BASIC_AUTH_USER=user
PYROSCOPE_BASIC_AUTH_PASSWORD=pass
HOSTNAME=new-api
PYROSCOPE_MUTEX_RATE=5
PYROSCOPE_BLOCK_RATE=5
```

**采集的指标类型**：
- CPU Profile
- 内存分配（AllocObjects、AllocSpace）
- 内存使用（InuseObjects、InuseSpace）
- Goroutine 数量
- Mutex 锁竞争
- Block 阻塞

定义文件：`common/pyro.go`

### 数据库消费日志

`logs` 表记录 API 调用的消费信息，参见上文日志表结构说明。

**重要**：数据库日志不存储实际的对话内容（prompt/response）。如果需要完整的请求/响应记录，需要：

1. 自行集成外部日志系统（ELK、Loki、Fluentd 等）
2. 在 `relay` 层添加中间件拦截请求体
3. 使用上游 provider 的审计日志功能

### 相关环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SQL_DSN` | 空 | 数据库连接串 |
| `LOG_SQL_DSN` | 空 | 日志数据库连接串（可与主库分离） |
| `DEBUG_ENABLED` | false | 启用 DEBUG 日志 |
| `ERROR_LOG_ENABLED` | false | 启用错误日志 |
| `LOG_CONSUME_ENABLED` | true | 启用消费日志记录 |
| `PYROSCOPE_URL` | 空 | Pyroscope 服务地址 |

---

## 参考文件

| 文件 | 内容 |
|------|------|
| `model/main.go` | 数据库初始化、迁移、兼容性处理 |
| `model/user.go` | 用户表定义 |
| `model/token.go` | 令牌表定义 |
| `model/channel.go` | 渠道表定义 |
| `model/log.go` | 日志表定义及查询 |
| `common/sys_log.go` | 系统日志函数 |
| `logger/logger.go` | 日志文件管理 |
| `common/pyro.go` | Pyroscope 集成 |
| `service/quota.go` | 配额消费逻辑 |
| `service/text_quota.go` | 文本请求消费计算 |