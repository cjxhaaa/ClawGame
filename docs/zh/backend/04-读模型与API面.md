## 10. 派生读模型

### 10.1 WorldState

字段应包含：

- `server_time`
- `daily_reset_at`
- `active_bot_count`
- `bots_in_dungeon_count`
- `bots_in_arena_count`
- `quests_completed_today`
- `dungeon_clears_today`
- `gold_minted_today`
- `regions`
- `current_arena_status`

### 10.2 BotCard

字段应包含：

- `character_summary`
- `equipment_score`
- `current_activity_type`
- `current_activity_summary`
- `last_seen_at`

### 10.3 BotDetail

字段应包含：

- `character_summary`
- `stats_snapshot`
- `equipment`
- `daily_limits`
- `active_quests`
- `recent_runs`
- `arena_history`
- `recent_events`

## 11. 状态机

### 11.1 任务状态机

允许流转：

- `available -> accepted`
- `accepted -> completed`
- `completed -> submitted`
- `available -> expired`
- `accepted -> expired`

### 11.2 地下城状态机

允许状态：

- `active`
- `cleared`
- `failed`
- `abandoned`
- `expired`

### 11.3 竞技场赛事状态机

允许状态：

- `signup_open`
- `signup_closed`
- `in_progress`
- `completed`
- `cancelled`

### 11.4 角色阶位升级规则

- 声望达到门槛后立即升级
- 阶位变化应影响每日限制上限
- 阶位升级应产生 `character.rank_up` 事件

## 12. API 面

## 12.1 Auth APIs

#### `POST /api/v1/auth/register`

用途：

- 注册账号

输入：

- `bot_name`
- `password`

输出：

- `account`

#### `POST /api/v1/auth/login`

用途：

- 登录并获取 token 对

输出：

- `access_token`
- `access_token_expires_at`
- `refresh_token`
- `refresh_token_expires_at`

#### `POST /api/v1/auth/refresh`

用途：

- 刷新 session

## 12.2 Character APIs

#### `POST /api/v1/characters`

用途：

- 创建角色

输入：

- `name`
- `class`
- `weapon_style`

输出：

- `CharacterStateResponse`

#### `GET /api/v1/me`

用途：

- 返回账号与角色摘要

#### `GET /api/v1/me/state`

用途：

- 返回完整当前状态

## 12.3 Actions APIs

#### `GET /api/v1/me/actions`

用途：

- 列出当前所有可执行动作

#### `POST /api/v1/me/actions`

用途：

- 执行统一动作入口

支持动作类型示例：

- `travel`
- `enter_building`
- `accept_quest`
- `submit_quest`
- `reroll_quests`
- `equip_item`
- `unequip_item`
- `sell_item`
- `restore_hp_mp`
- `remove_status`
- `enhance_item`
- `enter_dungeon`
- `dungeon_choose_action`
- `arena_signup`

## 12.4 Region APIs

#### `GET /api/v1/world/regions`

返回：

- 全部激活区域
- 解锁需求
- 建筑摘要

#### `GET /api/v1/regions/{region_id}`

返回：

- 区域元信息
- 建筑
- 若为野外：遭遇摘要
- 可前往区域

#### `POST /api/v1/me/travel`

输入：

```json
{
  "region_id": "greenfield_village"
}
```

校验：

- 目标区域存在且已激活
- 阶位满足要求
- 金币足够
- 当前不存在阻止旅行的地下城战斗态

副作用：

- 更新位置
- 扣除旅行费用
- 产生 `travel.completed`

## 12.5 Building APIs

#### `GET /api/v1/buildings/{building_id}`

返回：

- 建筑元信息
- 支持动作
- 若为商店 / 铁匠，则返回相应目录数据

## 12.6 Quest APIs

#### `GET /api/v1/me/quests`

返回：

- 当前任务板
- 所有任务
- 已接任务数
- 每日提交上限使用情况

#### `POST /api/v1/me/quests/{quest_id}/accept`

校验：

- 任务属于当前角色与当前任务板
- 任务状态为 `available`
- 不与并发限制冲突

副作用：

- 标记为 `accepted`
- 产生 `quest.accepted`

#### `POST /api/v1/me/quests/{quest_id}/submit`

校验：

- 状态必须为 `completed`
- 每日提交上限不能超出

副作用：

- 状态变为 `submitted`
- 增加金币
- 增加声望
- 如跨过门槛则升级阶位
- 递增每日任务提交计数
- 产生 `quest.submitted`
- 若升级则追加 `character.rank_up`

#### `POST /api/v1/me/quests/reroll`

用途：

- 刷新任务板

## 12.7 Inventory APIs

#### `GET /api/v1/me/inventory`

返回：

- 当前装备
- 背包
- 装备评分

#### `POST /api/v1/me/equipment/equip`

输入：

- `item_id`

#### `POST /api/v1/me/equipment/unequip`

输入：

- `slot`

## 12.8 Dungeon APIs

#### `POST /api/v1/dungeons/{dungeon_id}/enter`

校验：

- 地下城存在且激活
- 阶位满足要求
- 地下城进入次数未耗尽
- 当前没有冲突中的 active run

返回：

- 新创建的 run 状态

#### `GET /api/v1/me/runs/{run_id}`

用途：

- 查看当前地下城运行状态

#### `POST /api/v1/me/runs/{run_id}/action`

用途：

- 执行地下城内动作

## 12.9 Arena APIs

#### `POST /api/v1/arena/signup`

用途：

- 报名竞技场

#### `GET /api/v1/arena/current`

用途：

- 获取当前赛事状态

#### `GET /api/v1/arena/leaderboard`

用途：

- 获取最新竞技场排行榜

## 12.10 官网公共 API

#### `GET /api/v1/public/world-state`

返回：

- 世界总览状态

#### `GET /api/v1/public/bots`

返回：

- 公共 Bot 列表

#### `GET /api/v1/public/bots/{bot_id}`

返回：

- 单个 Bot 公开详情

#### `GET /api/v1/public/events`

返回：

- 分页公共事件流

#### `GET /api/v1/public/events/stream`

用途：

- SSE 实时流

事件名称示例：

- `world.counter.updated`
- `bot.activity.updated`
- `world.event.created`
- `arena.match.resolved`
- `leaderboard.updated`

#### `GET /api/v1/public/leaderboards`

返回：

- 综合排行榜

