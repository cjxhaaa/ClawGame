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
- `cleared`
- `failed`
- `abandoned`
- `expired`

`active` 内部的运行态阶段流转：

- `room_preparing -> in_combat`
- `in_combat -> room_cleared`
- `in_combat -> rating_pending`
- `room_cleared -> room_preparing`
- `room_cleared -> rating_pending`
- `rating_pending -> completed`

规则：

- 房间胜利后应立即暂存该房间的击杀材料掉落
- 运行提前结束时，`current_rating` 由 `highest_room_cleared` 推导
- 通关第 `6` 房后进入 `rating_pending`，再结算成 `cleared`
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
- 地下城进入次数未耗尽
- 当前没有冲突中的 active run

返回：

- 新创建的 run 状态

推荐返回字段：

- `run_id`
- `run_status`
- `runtime_phase`
- `current_room_index`
- `highest_room_cleared`
- `projected_rating`
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
- 当前房间摘要
- 战斗快照
- 已暂存的击杀材料掉落
- 待结算的评级装备奖励
- 可执行动作
- 最近战斗日志

#### `POST /api/v1/me/runs/{run_id}/action`

用途：

- 执行地下城内动作

支持动作：

- `start_room`
- `battle_attack`
- `battle_skill`
- `battle_use_consumable`
- `battle_defend`
- `claim_room_drops`
- `continue_to_next_room`
- `settle_rating_rewards`
- `abandon_run`

动作规则：

- `start_room` 仅可在 `room_preparing` 使用
- 战斗类动作仅可在 `in_combat` 使用
- `claim_room_drops` 仅可在 `room_cleared` 且存在暂存掉落时使用
- `continue_to_next_room` 仅可在当前房间完成后使用
- `settle_rating_rewards` 仅可在 `rating_pending` 使用
- `abandon_run` 可在任意 active 运行阶段使用

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
