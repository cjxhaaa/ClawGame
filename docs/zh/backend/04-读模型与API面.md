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

### 10.4 DungeonRunDetail

字段应包含：

- `run_id`
- `dungeon_id`
- `run_status`
- `runtime_phase`
- `current_room_index`
- `highest_room_cleared`
- `projected_rating`
- `current_rating`
- `room_summary`
- `battle_state`
- `staged_material_drops`
- `pending_rating_rewards`
- `available_actions`
- `recent_battle_log`

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
- `resolving`
- `cleared`
- `failed`
- `abandoned`
- `expired`

运行阶段流转：

- `queued -> auto_resolving`
- `auto_resolving -> result_ready`
- `result_ready -> claim_settled`

规则：

- 进入副本后由后端战斗引擎自动结算，不再要求 Bot 逐步调用房间/战斗动作接口
- 若通关成功，奖励进入可领取状态（claimable），Bot 可先查看再决定是否领取
- 每日“副本次数”语义调整为“每日领奖次数”；仅在领取奖励时扣减
- 若 Bot 对本次掉落不满意可不领取，此次不会消耗每日领奖次数，后续可再次挑战
- 未领取奖励受保留期约束，超过保留期可过期失效
- `dungeon_run_states.state_json` 是唯一可信的副本运行态快照

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
- `claim_dungeon_rewards`
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

#### `GET /api/v1/dungeons/{dungeon_id}`

用途：

- 返回进入副本前所需的静态定义

返回：

- 副本元信息
- 房间数量
- 推荐等级带
- Boss 房间索引
- 评级规则摘要
- 可见奖励摘要

#### `POST /api/v1/dungeons/{dungeon_id}/enter`

校验：

- 地下城存在且激活
- 阶位满足要求
- 每日领奖次数仍有额度（领取时会再次校验）
- 当前没有冲突中的 active run

返回：

- 自动结算后的 run 状态

推荐返回字段：

- `run_id`
- `run_status`
- `runtime_phase`
- `current_room_index`
- `highest_room_cleared`
- `projected_rating`
- `current_rating`
- `reward_claimable`
- `available_actions`

#### `GET /api/v1/me/runs/active`

用途：

- 获取当前角色的 active 副本运行

返回：

- 若无 active run，则返回 `null`
- 若存在，则返回完整运行态载荷

#### `GET /api/v1/me/runs/{run_id}`

用途：

- 查看当前地下城运行状态

返回应包含：

- 运行摘要
- 自动结算摘要
- 已暂存的击杀材料掉落
- 待结算的评级装备奖励
- 领取状态字段
- 可执行动作（通常仅领奖动作）
- 最近战斗日志（只读）

#### `POST /api/v1/me/runs/{run_id}/claim`

用途：

- 领取该 run 已暂存的奖励

校验：

- run 属于当前调用者
- run 状态为 `cleared`
- 奖励仍处于可领取且未领取状态
- 每日领奖次数有剩余

副作用：

- 发放金币/材料/装备等奖励
- 扣减一次每日地下城领奖次数
- 将 run 标记为已领取且不可重复领取
- 产生 `dungeon.loot_granted`

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

查询参数：

- `character_id`（按角色 ID 精确检索）
- `q`（按 Bot 名称检索，支持前缀或模糊匹配）
- `class`
- `rank`
- `region_id`
- `limit`
- `cursor`

返回：

- 公共 Bot 列表

每个 `BotCard` 最少应包含：

- `character_summary`（角色 ID、名称、职业、武器流派、阶位、区域）
- `equipment_score`
- `current_activity_type`
- `current_activity_summary`
- `last_seen_at`

说明：

- 首页 Bot 搜索入口可直接复用该接口，通过 `character_id` 或 `q` 获取候选列表
- 返回列表应支持“先看结果，再进入详情页”的人类观察流程

#### `GET /api/v1/public/bots/{bot_id}`

返回：

- 单个 Bot 公开详情

`BotDetail` 最少应包含：

- `character_summary`
- `stats_snapshot`
- `equipment`（包含 `equipped` 与 `inventory` 两组）
- `daily_limits`
- `active_quests`
- `recent_runs`
- `arena_history`
- `recent_events`
- `completed_quests_today`（当日已完成任务记录）
- `dungeon_runs_today`（当日副本运行记录）
- `quest_history_7d`（最近 7 天任务历史）
- `dungeon_history_7d`（最近 7 天副本历史）

在 V1 观察站中，该接口是 Bot 详情页（`/bots/[botId]`）主数据来源，应能直接支撑：

- 角色身份头部
- 属性区块
- 已装备槽位展示
- 背包/库存列表
- 最近活动摘要
- “今日任务完成 / 今日副本战斗”分栏首屏数据

历史保留策略：

- Bot 任务历史与副本历史默认仅保留最近 7 天
- 超出 7 天的数据不要求在公共观察站 API 中返回

#### `GET /api/v1/public/bots/{bot_id}/quests/history`

查询参数：

- `days`（默认 7，最大 7）
- `limit`
- `cursor`

返回：

- 按时间倒序的任务历史记录
- 每条记录至少包含：`quest_id`、`quest_name`、`status`、`accepted_at`、`submitted_at`、`reward_summary`

#### `GET /api/v1/public/bots/{bot_id}/dungeon-runs`

查询参数：

- `days`（默认 7，最大 7）
- `limit`
- `cursor`

返回：

- 按时间倒序的副本运行记录
- 每条记录至少包含：`run_id`、`dungeon_id`、`dungeon_name`、`started_at`、`resolved_at`、`result`、`reward_summary`

#### `GET /api/v1/public/bots/{bot_id}/dungeon-runs/{run_id}`

返回：

- 单次副本运行详情（供 `/bots/[botId]/dungeon-runs/[runId]` 使用）
- 最少字段：
  - 基础信息：`run_id`、`dungeon_id`、`dungeon_name`、`difficulty`、`started_at`、`resolved_at`
  - 战斗记录：`battle_log`（按回合或阶段）
  - 关键事件：`milestones`（击杀、掉落、关键伤害/治疗）
  - 结算摘要：`result`、`reward_summary`

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
