# Changelog

## [Unreleased] — TokenScope 成本分析

### 新增

- **成本分析页面**：侧边栏新增"成本分析"菜单项，独立路由 `/token-efficiency`
  - L1 运行监控 Tab：三项核心指标卡片（产出成本、上下文负载、缓存复用率）+ 按模型分组明细表
  - L2 深度诊断 Tab（仅管理员可见）：
    - 最近采样请求表：显示 request ID、模型、分段数、估算 token 数、时间
    - 点击行展开查看上下文分段详情（类型、名称、token 数、prefix/cached/repeated 标记）
    - 聚合摘要表：采样数、各部分平均 Token、重复前缀率、缓存友好度、缓存兑现率
  - 筛选栏：支持 24h / 7d / 30d 快捷时间范围、模型名称筛选、分组筛选
  - L1 维度切换：用户 / API 密钥 / 渠道 / 模型，左上角 TAB 切换

- **后端 API**：
  - `/api/tokenscope/l1/summary` — L1 总览指标（admin）
  - `/api/tokenscope/l1/by-model` — L1 按模型分组（admin）
  - `/api/tokenscope/l1/by-dimension` — L1 按维度分组（admin）
  - `/api/tokenscope/l1/timeseries` — L1 时序数据（admin）
  - `/api/tokenscope/self/l1/summary` — L1 当前用户总览
  - `/api/tokenscope/self/l1/by-dimension` — L1 当前用户按维度分组
  - `/api/tokenscope/l2/summary` — L2 诊断指标（admin）
  - `/api/tokenscope/l2/by-dimension` — L2 按维度分组（admin）
  - `/api/tokenscope/self/l2/by-dimension` — L2 当前用户按维度分组
  - `/api/tokenscope/l2/recent` — L2 最近采样请求（admin）
  - `/api/tokenscope/l2/request/:request_id` — L2 请求详情（admin）
  - `/api/tokenscope/self/l2/recent` — L2 当前用户最近请求

- **数据层**：
  - `logs` 表新增 3 列：`cache_read_tokens`、`cache_write_tokens`、`input_tokens_total`（GORM AutoMigrate 自动添加）
  - `request_debug_payloads` 表：L2 调试数据存储（raw request body 采样）
  - `request_context_parts` 表：解析后的上下文结构（system/history/tool/file 分类 + 缓存分析标记）

- **L2 采样与解析**：
  - 请求体采样钩子（确定性哈希采样 + 模型白名单 + 异步捕获）
  - OpenAI / Claude / Gemini 三种格式自动解析
  - SHA-256 前缀去重、4-chars-per-token 快速估算
  - 自动标记：is_prefix / is_repeated / is_stable / is_cache_friendly

- **配置**：
  - `tokenscope_setting`：L2 采样开关、采样率、最大 payload、保留天数、捕获模型列表

- **i18n**：新增 40 个翻译 key（中/英），涵盖成本分析页面全部文案

### 修复

- 前端 API 调用适配后端 `{ success, data }` 统一响应格式（`res.data.data` 解包）
- 侧边栏"成本分析"菜单项默认可见（`DEFAULT_SIDEBAR_MODULES.console.tokenscope: true`）
- 翻译 key 放入 `translation` 命名空间内，确保中文切换生效
- L1SummaryCards 组件 null 安全处理（`?.` 可选链 + 空值回退）
- L2 维度查询修复：`request_debug_payloads` 无 username 列，L2 按用户维度时 LEFT JOIN logs 表
- 数据库迁移：将 `RequestDebugPayload` 和 `RequestContextPart` 加入 LOG_DB AutoMigrate
- 表名修复：GORM 自动复数化表名为 `request_debug_payloads`，查询代码已对齐

## [Unreleased] — TokenScope 成本分析改进

### 新增

- **L1 明细表排序**：请求数、产出成本、上下文负载、缓存复用、总配额、输入 Token、输出 Token、缓存读取等 8 个数值列支持点击表头排序
  - 升序/降序切换，当前排序列高亮显示，非激活列箭头半透明
  - 双箭头指示器（▲▼），适配 Table 组件 `[&_th_*]:text-sm` 样式覆盖

- **后端 API**：
  - `/api/tokenscope/filters` — 获取筛选选项（模型名称、分组列表，admin）
  - `/api/tokenscope/self/filters` — 获取当前用户筛选选项

- **筛选栏增强**：新增模型名称和分组下拉筛选，数据从后端动态加载

- **前端配置**：
  - Rsbuild dev server 固定端口 3003（避免端口漂移）
  - `.env.local` 配置 `VITE_REACT_APP_SERVER_URL` 指向后端端口

### 修复

- L1 指标卡片子文本优化：使用 i18n 插值替代裸拼接，Token 数量使用 K/M 缩写
  - 产出成本：`543 请求` → `543 次请求`
  - 上下文负载：`输入: 6,199,888 | 输出: 180,322` → `输入 6.2M · 输出 180.3K`
  - 缓存复用率：`缓存读取: 3,716,224` → `缓存读取 3.7M Token`
- L1 指标卡片描述文案优化（中文）：
  - 产出成本：`每 1K 输出 Token 的成本` → `每千输出 Token 成本`
  - 上下文负载：`每输出 1 Token 的输入 Token 数` → `输入与输出 Token 比值`
  - 缓存复用率：`输入中来自缓存的比例` → `输入 Token 的缓存命中占比`

## [Unreleased] — TokenScope L1 多维度分组

### 新增

- **L1 多维度分组**：L1 明细表新增 ToggleGroup 切换分组维度
  - 按模型（默认）、按 API Key、按用户（仅管理员）、按渠道（仅管理员）
  - 各维度首列自动适配：显示名称 + 关联 ID（如 username #user_id）
  - 渠道维度自动解析 channel name 并展示

- **后端 API**：
  - `/api/tokenscope/l1/by-dimension?dimension=model|user|key|channel` — L1 按维度分组（admin）
  - `/api/tokenscope/self/l1/by-dimension?dimension=model|key` — L1 当前用户按维度分组
  - 通用查询函数 `getTokenScopeL1ByDimension`，按 dimension 动态选择 GROUP BY 列
  - 渠道维度自动批量查询 channel name（`GetChannelsByIds`）
  - PostgreSQL 兼容：`channel_id::text` 替代 `CAST(channel_id AS CHAR)`

- **数据结构变更**：`TokenScopeL1Metrics` 新增 `name`（替代 `model_name`）、`dimension`、`sub_id`、`sub_name` 字段

- **i18n**：新增"按用户"、"按 API Key"、"按渠道"翻译 key（中/英）
