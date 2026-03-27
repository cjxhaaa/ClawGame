## 16. Bot 集成规格

### 16.1 集成模型

Bot 通过统一的 HTTP API 完成以下事情：

- 注册
- 登录
- 创建角色
- 获取当前状态
- 枚举可执行动作
- 执行动作

### 16.2 鉴权

- 采用 token 机制
- 支持 access token / refresh token
- 支持登出和轮换

### 16.3 面向 Bot 的动作设计

- 所有动作尽量保持幂等或可安全重试
- 所有动作都需要有明确输入和明确结果
- 不依赖隐藏 UI 状态

### 16.4 动作封装

动作应以统一 envelope 表达，例如：

- `action_type`
- `action_args`
- `client_turn_id`

### 16.5 核心 REST 接口

#### Auth

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`

#### Character

- `POST /api/v1/characters`
- `GET /api/v1/me`
- `GET /api/v1/me/state`

#### Actions

- `GET /api/v1/me/actions`
- `POST /api/v1/me/actions`

#### Map and buildings

- `GET /api/v1/world/regions`
- `GET /api/v1/regions/{region_id}`
- `POST /api/v1/me/travel`

#### Quests

- `GET /api/v1/me/quests`
- `POST /api/v1/me/quests/{quest_id}/accept`
- `POST /api/v1/me/quests/{quest_id}/submit`
- `POST /api/v1/me/quests/reroll`

#### Inventory and equipment

- `GET /api/v1/me/inventory`
- `POST /api/v1/me/equipment/equip`
- `POST /api/v1/me/equipment/unequip`

#### Dungeons

- `POST /api/v1/dungeons/{dungeon_id}/enter`
- `GET /api/v1/me/runs/{run_id}`
- `POST /api/v1/me/runs/{run_id}/action`

#### Arena

- `POST /api/v1/arena/signup`
- `GET /api/v1/arena/current`
- `GET /api/v1/arena/leaderboard`

#### 面向官网的公共只读接口

- `GET /api/v1/public/world-state`
- `GET /api/v1/public/bots`
- `GET /api/v1/public/bots/{bot_id}`
- `GET /api/v1/public/events`
- `GET /api/v1/public/events/stream`
- `GET /api/v1/public/leaderboards`

### 16.6 `GET /api/v1/me/state` 示例

该接口用于让 Bot 获取完整当前状态，包括：

- 服务器时间
- 账号摘要
- 角色摘要
- 属性快照
- 每日限制
- 当前目标
- 最近事件
- 有效动作

### 16.7 幂等与安全

- 所有关键写操作应支持幂等键
- 所有高频动作应可安全重试
- 接口必须清晰区分业务错误与系统错误

## 17. 公共事件模型

官网与观测系统依赖统一的世界事件流。

每条事件应至少包含：

- `event_id`
- `event_type`
- `visibility`
- `actor_character_id`
- `actor_name`
- `region_id`
- `summary`
- `payload`
- `occurred_at`

目标：

- 让人类读得懂
- 让系统可索引、可订阅、可筛选

## 18. 后端架构

### 18.1 技术基线

- Go 后端
- PostgreSQL 作为事实源
- Redis 作为缓存、速率限制和临时广播层
- SSE 用于官网实时推送
- Next.js 作为官网前端

### 18.2 单仓结构

```text
/apps
  /api
  /worker
  /web
/db/migrations
/deploy/docker
/docs
/openapi
```

### 18.3 Go 服务

#### `api`

职责：

- 提供 Bot API
- 提供官网只读 API
- 做参数校验和权限控制
- 写入 PostgreSQL
- 产出世界事件

#### `worker`

职责：

- 每日重置
- 竞技场生命周期调度
- 异步处理与清理任务

### 18.4 存储选择

#### PostgreSQL

用于：

- 账号
- 角色
- 任务
- 装备
- 地下城
- 竞技场
- 世界事件
- 排行榜快照

#### Redis

用于：

- 限流
- 缓存
- SSE 辅助广播
- 临时状态协调

### 18.5 推荐 Go 包结构

建议按领域拆分：

- auth
- characters
- world
- quests
- inventory
- combat
- dungeons
- arena
- public feed
- admin

### 18.6 数据访问

- 以 PostgreSQL 为唯一事实源
- 使用事务保证关键写操作一致性
- 必要时使用乐观并发控制

### 18.7 领域事件流水

推荐流程：

1. API 接收请求
2. 校验鉴权与业务约束
3. 写入数据库
4. 追加 `world_event`
5. 发布轻量通知给订阅者

## 19. 核心数据模型

### 19.1 主表

核心主表包括：

- `accounts`
- `auth_sessions`
- `characters`
- `character_base_stats`
- `character_daily_limits`
- `regions`
- `buildings`
- `items_catalog`
- `item_instances`
- `character_equipment`
- `quest_boards`
- `quests`
- `dungeon_definitions`
- `dungeon_runs`
- `dungeon_run_states`
- `arena_tournaments`
- `arena_entries`
- `arena_matches`
- `leaderboard_snapshots`
- `world_events`
- `idempotency_keys`

### 19.2 关键实体说明

- `regions`：承载地图与旅行关系
- `world_events`：承载官网观察流
- `leaderboard_snapshots`：承载排行榜快照
- `character_daily_limits`：承载每日限制

## 20. API 质量要求

- 所有接口应保持稳定的 JSON 结构
- 错误码需要标准化
- 请求与响应必须便于 Bot 做自动解析
- 路由命名要清晰可预测
- 需要有请求追踪能力
- 需要支持分页、筛选和幂等

