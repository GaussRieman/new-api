# 1. 核心定义
**Modi Token Efficiency Stack** 是一套面向企业 Agent 应用的 **Token 使用效率工程体系**。
它通过：
```text
统一 Gateway 接入模型
TokenScope 分析消耗
Modi-Harness 优化上下文
Modi Apps 承载业务应用
```
实现：
```text
Token 消耗可观测
上下文问题可诊断
缓存收益可验证
优化效果可闭环
```
一句话：
> **模型从 Provider 来，数据在 Gateway 看，问题由 TokenScope 诊断，优化在 Modi-Harness 做，效果回 Gateway 验证。**
---
# 2. 核心目标
不是单纯"省 Token"，而是建立企业 AI / Agent 应用的 **Token 使用效率控制能力**。
目标分四个：
```text
1. 看得见
   统一采集模型调用、Token、成本、缓存、延迟、错误等运行数据。
2. 算得清
   用统一指标衡量不同模型、项目、用户、Key、应用的 Token 使用效率。
3. 找得到
   通过 Raw Body / Context Parts 找到 Token 浪费来源。
4. 改得动
   通过 Modi-Harness 优化 Prompt、Tools、History、Memory、RAG、缓存友好度。
```
最终目标：
> **让企业 Agent 应用从"Token 黑盒消耗"变成"可观测、可诊断、可优化、可验证"的工程系统。**
---
# 3. 核心概念
## 3.1 Provider
模型供应层。
包括：
```text
Claude
OpenAI
DeepSeek
Qwen
GLM
私有化模型
第三方聚合商
```
Provider 负责提供：
```text
模型能力
价格体系
上下文窗口
缓存策略
工具调用协议
usage 字段
```
关键点：
> Provider 决定模型能力和服务端缓存策略。
---
## 3.2 Unified Gateway
统一接入层。
负责：
```text
统一接入
统一鉴权
统一路由
统一日志
统一计费
统一限流
统一 usage 字段归一化
```
Gateway 的定位：
> **Gateway 是观测入口，负责看见所有模型调用。**
它主要解决：
```text
谁在用？
用了哪个模型？
用了多少 Token？
花了多少钱？
缓存命中了多少？
是否异常？
```
---
## 3.3 TokenScope
核心分析模块。
负责：
```text
成本分析
Token 分析
缓存分析
异常识别
Raw Body Debug
Context Parts 解析
优化建议生成
优化效果验证
```
TokenScope 的定位：
> **TokenScope 是分析大脑，负责看懂 Token 消耗。**
它回答：
```text
哪里贵？
为什么贵？
上下文重在哪里？
缓存有没有生效？
优化后有没有真的变好？
```
---
## 3.4 Modi-Harness
自研 Agent Harness 框架。
负责：
```text
Prompt 管理
Tool 管理
Memory 管理
History 管理
RAG Context 管理
Context Budget
Provider Cache Adapter
Agent Runtime
```
Modi-Harness 的定位：
> **Modi-Harness 是控制层，负责改变 Token 使用结构。**
它真正能优化：
```text
上下文裁剪
上下文分层
稳定前缀
动态内容后置
工具定义瘦身
历史摘要
Memory 注入控制
RAG TopK 控制
工具结果压缩
Provider 缓存适配
```
---
## 3.5 Modi Apps
上层业务应用。
例如：
```text
Modi-Coding
Modi-Doc
Modi-Data
Modi-Case
Modi-Agent Apps
```
Modi Apps 的定位：
> **Modi Apps 是业务执行层，负责把 Harness 能力落到真实场景。**
特点：
```text
不是直接裸调模型
而是通过 Modi-Harness 构造上下文
通过 Gateway 调用模型
通过 TokenScope 反馈优化效果
```
---
# 4. 核心关系
```text
Provider
  ↓ 提供模型能力
Unified Gateway
  ↓ 统一接入与采集
TokenScope
  ↓ 分析与诊断
Modi-Harness
  ↓ 上下文控制与优化
Modi Apps
  ↓ 业务场景运行
Gateway / TokenScope
  ↑ 回收数据并验证效果
```
更简洁：
```text
Provider：供给模型
Gateway：记录调用
TokenScope：分析问题
Modi-Harness：执行优化
Modi Apps：产生业务价值
```
---
# 5. 核心闭环
```text
1. Modi Apps 发起任务
2. Modi-Harness 构造上下文
3. Gateway 调用 Provider
4. Gateway 记录 request_log
5. TokenScope 分析产出成本、上下文负载、缓存复用率
6. 发现异常后采样 Raw Body
7. TokenScope 解析 Context Parts
8. Modi-Harness 调整上下文结构
9. 再次运行任务
10. TokenScope 验证优化效果
```
一句话：
> **Gateway 发现问题，TokenScope 解释问题，Modi-Harness 改造问题，Modi Apps 验证业务可用性。**
---
# 6. 最核心指标
只保留三项运行指标：
```text
1. 产出成本
   = 总费用 / 输出 Token
   看 AI 输出贵不贵
2. 上下文负载
   = 输入 Token / 输出 Token
   看为了获得输出，喂了多少上下文
3. 缓存复用率
   = 缓存读取 Token / 输入 Token
   看输入中有多少来自缓存复用
```
深度诊断再看四项：
```text
1. 上下文结构
   Token 花在 system、history、tool、file、memory 哪些部分
2. 重复前缀率
   有多少稳定前缀被反复传入
3. 缓存友好度
   请求结构是否有利于 Provider 缓存
4. 缓存兑现率
   Provider 是否真正返回缓存收益
```
---
# 7. 最终一句话
> **Modi Token Efficiency Stack = Gateway 接入与观测 + TokenScope 分析与诊断 + Modi-Harness 上下文优化 + Modi Apps 场景验证。**
它的本质是：
> **企业 Agent 应用的 Token 工程化体系。**