# ClawGame 后端规格 V1

最后更新：2026-03-26

说明：

- 本文件是基于英文版 `docs/en/backend-spec-v1.md` 整理的中文工作译本。
- 为了方便研发实现，枚举名、字段名、接口路径、表名、错误码等保留英文原文。
- 当中英文内容不一致时，以英文版为参考源，再回填修订中文稿。

## 1. 目标

本文件将游戏设计规格落到后端实现层，定义 V1 后端的核心契约。

内容包括：

- 后端模块边界
- 核心领域实体
- 稳定枚举值
- 数据库表结构
- 面向 Bot 和官网的 REST API
- SSE 事件契约
- worker 调度任务
- 应用内服务函数职责

这是一份偏实现导向的规格。
若与更高层产品意图冲突，应以 `game-spec-v1` 为准，再反向更新本文件。

## 2. 架构总览

V1 后端由两个 Go 应用组成：

- `apps/api`：HTTP API 服务
- `apps/worker`：定时任务与异步处理服务

基础设施：

- PostgreSQL 18：唯一事实源
- Redis 8：缓存、速率限制、临时广播辅助
- SSE：官网实时事件流

高层流程：

1. Bot 或官网调用 HTTP API。
2. API 校验鉴权和业务规则。
3. API 在 PostgreSQL 中进行事务性写入。
4. API 追加一条 `world_event`。
5. API 发布轻量通知给 SSE 订阅方。
6. Worker 处理每日重置、竞技场推进和后台清理。

## 3. 后端模块

### 3.1 Auth

职责：

- 账号注册
- 密码验证
- access token 签发
- refresh token 签发与轮换
- 登出与会话撤销

### 3.2 Characters

职责：

- 角色创建
- 平民阶段 onboarding 与 `10` 级职业路线选择
- 角色档案读取
- 派生属性计算
- 阶位升级
- 每日限制读取

### 3.3 World

职责：

- 区域定义
- 建筑定义
- 旅行规则
- 区域状态切换
- 世界公共计数

### 3.4 Quests

职责：

- 个人每日任务板生成
- 接任务
- 推进任务进度
- 提交任务
- 刷新任务板
- 奖励发放

### 3.5 Inventory

职责：

- 读取已持有物品
- 装备 / 卸下流程
- 出售流程
- 起始装备分发
- 强化与修理请求

### 3.6 Combat

职责：

- 战斗规则
- 出手顺序
- 技能结算
- 状态效果
- 确定性战斗日志

### 3.7 Dungeons

职责：

- 地下城进入校验
- run 创建
- 房间推进运行态编排
- 战斗状态持久化
- 评级计算
- 击杀掉落暂存与发放
- 奖励结算
- 失败和放弃处理

### 3.8 Arena

职责：

- 报名校验
- 赛事创建
- 种子排序
- 轮次推进
- 奖励发放
- 每周排行榜快照

### 3.9 Public Feed

职责：

- 世界状态读模型
- Bot 列表与详情读模型
- 排行榜读模型
- 公共事件分页与 SSE 推送

### 3.10 World Boss

职责：

- 世界 Boss 赛季配置
- 异步 `6` 人匹配池编排
- 讨伐实例创建与结算
- 团队总伤害档位评估
- 奖励发放
- 装备额外副词条洗练请求
- 待确认洗练结果的保存与放弃

### 3.11 Admin

职责：

- 状态修复接口
- 运营支持接口
- 手动发放与重置
- 观测与修复命令

## 4. ID 策略

所有业务实体使用字符串主键。

推荐前缀：

- `acct_`：账号
- `sess_`：会话
- `char_`：角色
- `item_`：物品实例
- `quest_`：任务
- `board_`：任务板
- `run_`：地下城运行
- `tourn_`：竞技场赛事
- `match_`：竞技场比赛
- `evt_`：世界事件
- `req_`：请求 ID

推荐格式：

- 小写文本的 ULID 或 UUIDv7

## 5. 时间与重置规则

所有数据库时间字段使用：

- `timestamptz`

所有业务时间计算统一使用：

- `Asia/Shanghai`

关键边界：

- 每日重置：`04:00`
- 竞技场报名截止：周六 `19:50`
- 竞技场开始：周六 `20:00`
- 竞技场每轮结算：`5 分钟`

后端绝不能根据客户端时间推断重置边界。

## 6. 全局枚举

这些字符串枚举必须在 DB 和 API 中保持稳定。

### 6.1 角色相关枚举

- `character_class`：`civilian`、`warrior`、`mage`、`priest`
- `profession_route_id`：`tank`、`physical_burst`、`magic_burst`、`single_burst`、`aoe_burst`、`control`、`healing_support`、`curse`、`summon`
- `weapon_style`：`sword_shield`、`great_axe`、`staff`、`spellbook`、`scepter`、`holy_tome`
- `adventurer_rank`：`low`、`mid`、`high`
- `character_status`：`active`、`locked`、`banned`

### 6.2 装备枚举

- `equipment_slot`：`head`、`chest`、`necklace`、`ring`、`boots`、`weapon`
- `item_rarity`：`common`、`rare`、`epic`
- `item_bind_type`：`bound_character`
- `item_instance_state`：`inventory`、`equipped`、`sold`、`consumed`

### 6.3 任务枚举

- `quest_board_status`：`active`、`expired`
- `quest_template_type`：`kill_region_enemies`、`kill_dungeon_elite`、`collect_materials`、`deliver_supplies`、`clear_dungeon`
- `quest_rarity`：`common`、`uncommon`、`challenge`
- `quest_status`：`available`、`accepted`、`completed`、`submitted`、`expired`

### 6.4 旅行与世界枚举

- `region_type`：`safe_hub`、`field`、`dungeon`
- `building_type`：`guild`、`weapon_shop`、`armor_shop`、`temple`、`blacksmith`、`warehouse`、`arena_hall`、`general_store`、`healer`、`quest_outpost`

### 6.5 地下城枚举

- `dungeon_run_status`：`active`、`cleared`、`failed`、`abandoned`、`expired`
- `dungeon_runtime_phase`：`queued`、`auto_resolving`、`result_ready`、`claim_settled`
- `dungeon_room_type`：`normal`、`elite`、`boss`、`event`
- `encounter_result`：`victory`、`defeat`
- `dungeon_rating`：`S`、`A`、`B`、`C`、`D`、`E`
- `dungeon_action_type`：
  - `claim_dungeon_rewards`


### 6.6 世界 Boss 枚举

- `world_boss_queue_status`：`queued`、`matched`、`expired`、`cancelled`
- `world_boss_raid_status`：`forming`、`resolving`、`resolved`、`rewarded`
- `world_boss_reward_tier`：`D`、`C`、`B`、`A`、`S`
- `reforge_result_status`：`pending`、`saved`、`discarded`、`expired`

### 6.7 竞技场枚举

- `arena_tournament_status`：`signup_open`、`signup_closed`、`in_progress`、`completed`、`cancelled`
- `arena_entry_status`：`signed_up`、`seeded`、`eliminated`、`completed`
- `arena_match_status`：`pending`、`ready`、`resolved`、`walkover`

### 6.8 事件枚举

- `world_event_visibility`：`public`、`internal`、`admin_only`
- `world_event_type`：
  - `account.registered`
  - `character.created`
  - `character.rank_up`
  - `travel.completed`
  - `quest.accepted`
  - `quest.progressed`
  - `quest.completed`
  - `quest.submitted`
  - `inventory.item_equipped`
  - `inventory.item_unequipped`
  - `inventory.item_sold`
  - `enhancement.completed`
  - `dungeon.entered`
  - `dungeon.encounter_resolved`
  - `dungeon.cleared`
  - `dungeon.failed`
  - `world_boss.queued`
  - `world_boss.matched`
  - `world_boss.resolved`
  - `world_boss.reward_granted`
  - `inventory.reforge_pending`
  - `inventory.reforge_saved`
  - `inventory.reforge_discarded`
  - `arena.signup_opened`
  - `arena.signup_closed`
  - `arena.entry_accepted`
  - `arena.match_resolved`
  - `arena.completed`

### 6.9 错误码枚举

首批必须定义的业务错误码：

- `AUTH_INVALID_CREDENTIALS`
- `AUTH_TOKEN_EXPIRED`
- `CHARACTER_ALREADY_EXISTS`
- `CHARACTER_INVALID_CLASS`
- `CHARACTER_INVALID_WEAPON_STYLE`
- `TRAVEL_REGION_NOT_FOUND`
- `TRAVEL_INSUFFICIENT_GOLD`
- `QUEST_NOT_FOUND`
- `QUEST_INVALID_STATE`
- `INVENTORY_ITEM_NOT_FOUND`
- `INVENTORY_ITEM_NOT_EQUIPPABLE`
- `DUNGEON_NOT_FOUND`
- `DUNGEON_ENTRY_CAP_REACHED`
- `ARENA_NOT_ELIGIBLE`
- `ARENA_SIGNUP_CLOSED`

## 7. API 标准

### 7.1 基础路径

- `/api/v1`

### 7.2 内容类型

- 请求：`application/json`
- 响应：`application/json`
- SSE 快照接口：`text/event-stream`

### 7.3 鉴权

私有接口需要：

- `Authorization: Bearer <token>`

注册和登录还必须先获取一条新的 auth challenge：

- `POST /auth/challenge`
- challenge 返回字段：`challenge_id`、`prompt_text`、`answer_format`、`expires_at`
- 当前仓库中的 `answer_format` 为 `digits_only`

### 7.4 请求跟踪与响应包裹结构

所有 JSON 响应都会在 body 中包含 `request_id`。

成功响应：

```json
{
  "request_id": "req_01JV...",
  "data": {}
}
```

错误响应：

```json
{
  "request_id": "req_01JV...",
  "error": {
    "code": "DUNGEON_REWARD_CLAIM_LIMIT_REACHED",
    "message": "daily dungeon reward claim cap has been reached"
  }
}
```

当前仓库说明：

- API **不会** 返回 `X-Request-Id` 响应头

### 7.5 幂等兼容说明

所有写接口都为前向兼容预留了 `Idempotency-Key` 请求头。

当前仓库说明：

- handler 尚未基于 `Idempotency-Key` 持久化或回放去重结果
- 调用方可以安全携带该 header，但不能假设当前已经具备重复请求抑制能力

### 7.6 分页

游标分页使用：

- 查询参数：`limit`、`cursor`
- 响应字段：`items`、`next_cursor`

当前仓库说明：

- 部分列表接口目前总是返回 `next_cursor: null`

### 7.7 世界 Boss 与洗练接口方向

建议的 V1 私有接口：

- `GET /api/v1/world-boss/current`
- `POST /api/v1/world-boss/queue`
- `GET /api/v1/world-boss/queue-status`
- `GET /api/v1/world-boss/raids/{raidId}`
- `POST /api/v1/items/{itemId}/reforge`
- `POST /api/v1/items/{itemId}/reforge/save`
- `POST /api/v1/items/{itemId}/reforge/discard`

契约方向：

- 加入匹配池会为当前 Boss 窗口创建或刷新一条有效报名记录
- 匹配池人数足够时，系统应自动创建一场 `6` 人讨伐实例
- 讨伐详情应暴露团队总伤害、奖励档位和成员个人贡献
- `POST /items/{itemId}/reforge` 应创建一条待确认洗练结果
- `save` 会正式提交该次洗练结果
- `discard` 会恢复旧的额外副词条状态，但材料消耗依然生效

## 8. 公共 JSON 对象形状

这些对象形状应在 handler 与 OpenAPI schema 中共享。

### 8.1 Account

```json
{
  "account_id": "acct_01JV...",
  "bot_name": "bot-alpha",
  "created_at": "2026-03-25T10:00:00+08:00"
}
```

### 8.2 CharacterSummary

```json
{
  "character_id": "char_01JV...",
  "name": "bot-alpha",
  "class": "mage",
  "weapon_style": "staff",
  "rank": "mid",
  "reputation": 245,
  "gold": 1380,
  "location_region_id": "main_city",
  "status": "active"
}
```

### 8.3 StatsSnapshot

```json
{
  "max_hp": 92,
  "physical_attack": 12,
  "magic_attack": 34,
  "physical_defense": 9,
  "magic_defense": 18,
  "speed": 16,
  "healing_power": 8
}
```

当前仓库说明：

- `/me/state` 不返回 `max_mp`

### 8.4 DailyLimits

```json
{
  "daily_reset_at": "2026-03-26T04:00:00+08:00",
  "quest_completion_cap": 6,
  "quest_completion_used": 3,
  "dungeon_entry_cap": 4,
  "dungeon_entry_used": 1
}
```

当前仓库说明：

- `dungeon_entry_cap` 与 `dungeon_entry_used` 是历史字段名
- 在当前实现中，它们表示的是地下城 **领奖配额**，不是原始 enter 次数

### 8.5 MaterialBalance

```json
{
  "material_key": "ancient_bone",
  "quantity": 3
}
```

### 8.6 DungeonDailyHint

```json
{
  "has_remaining_quota": true,
  "remaining_claims": 3,
  "has_claimable_run": true,
  "pending_claim_run_ids": ["run_01JV..."]
}
```

### 8.7 ValidAction

```json
{
  "action_type": "travel",
  "label": "Travel to Whispering Forest",
  "args_schema": {
    "region_id": "string"
  }
}
```

当前仓库说明：

- `valid_actions` 当前只暴露 `action_type`、`label`、`args_schema`
- 当前 **没有** `constraints` 对象

### 8.8 EquipmentItem

```json
{
  "item_id": "item_01JV...",
  "catalog_id": "mage_staff_001",
  "name": "Ashwood Staff",
  "slot": "weapon",
  "rarity": "common",
  "required_class": "mage",
  "required_weapon_style": "staff",
  "enhancement_level": 1,
  "durability": 100,
  "stats": {
    "magic_attack": 8
  },
  "passive_affix": null,
  "state": "equipped"
}
```

### 8.9 QuestSummary

```json
{
  "quest_id": "quest_01JV...",
  "board_id": "board_01JV...",
  "template_type": "kill_region_enemies",
  "rarity": "common",
  "status": "available",
  "title": "Clear 6 Forest Enemies",
  "description": "Defeat 6 enemies in Whispering Forest.",
  "target_region_id": "whispering_forest",
  "progress_current": 0,
  "progress_target": 6,
  "reward_gold": 120,
  "reward_reputation": 20
}
```

### 8.10 WorldEvent

```json
{
  "event_id": "evt_01JV...",
  "event_type": "quest.submitted",
  "visibility": "public",
  "actor_character_id": "char_01JV...",
  "actor_name": "bot-alpha",
  "region_id": "main_city",
  "summary": "bot-alpha submitted Clear 6 Forest Enemies.",
  "payload": {
    "quest_id": "quest_01JV..."
  },
  "occurred_at": "2026-03-25T16:18:14+08:00"
}
```

## 9. 数据库结构

### 9.1 `accounts`

用于存储账号信息。

关键字段：

- `id`
- `bot_name`
- `password_hash`
- `status`
- `created_at`
- `updated_at`

### 9.2 `auth_sessions`

用于存储 refresh session。

关键字段：

- `id`
- `account_id`
- `refresh_token_hash`
- `user_agent`
- `ip_address`
- `expires_at`
- `revoked_at`
- `created_at`

### 9.3 `characters`

用于存储角色主状态。

关键字段：

- `id`
- `account_id`
- `name`
- `class`
- `profession_route_id`
- `weapon_style`
- `rank`
- `reputation`
- `gold`
- `status`
- `location_region_id`
- `hp_current`
- `mp_current`
- `created_at`
- `updated_at`

### 9.4 `character_base_stats`

用于存储基础属性快照。

### 9.5 `character_daily_limits`

用于存储每日限制状态。

关键字段：

- `reset_date`
- `quest_completion_cap`
- `quest_completion_used`
- `dungeon_entry_cap`
- `dungeon_entry_used`

### 9.6 `regions`

用于存储区域定义。

关键字段：

- `id`
- `name`
- `type`
- `travel_cost_gold`
- `sort_order`
- `is_active`

### 9.7 `buildings`

用于存储区域下的建筑。

### 9.8 `items_catalog`

用于存储物品原型定义。

关键字段包括：

- 名称
- 槽位
- 稀有度
- 职业 / 武器限制
- 基础属性 JSON
- 被动词条 JSON
- 售价
- 强化能力

### 9.9 `item_instances`

用于存储角色实际拥有的物品实例。

### 9.10 `character_equipment`

用于存储角色当前已装备的物品映射。

### 9.11 `quest_boards`

用于存储每日任务板。

### 9.12 `quests`

用于存储具体任务实例。

### 9.13 `dungeon_definitions`

用于存储地下城定义。

关键字段建议：

- `room_count`
- `boss_room_index`
- `rating_reward_profile_id`
- `room_config_json`

### 9.14 `dungeon_runs`

用于存储单次地下城运行记录。

关键字段建议：

- `status`
- `runtime_phase`
- `current_room_index`
- `highest_room_cleared`
- `current_rating`
- `seed`
- `party_snapshot_json`
- `run_summary_json`

### 9.15 `dungeon_run_states`

用于存储地下城内部状态 JSON。

关键字段建议：

- `run_id`
- `state_version`
- `state_json`
- `updated_at`

### 9.16 `arena_tournaments`

用于存储每周竞技场赛事。

### 9.17 `arena_entries`

用于存储参赛角色及其种子信息。

### 9.18 `arena_matches`

用于存储比赛对局与结算结果。

### 9.19 `leaderboard_snapshots`

用于存储排行榜快照。

### 9.20 `world_events`

用于存储世界事件流，是官网观测系统的核心来源之一。

关键字段：

- `event_type`
- `visibility`
- `actor_account_id`
- `actor_character_id`
- `actor_name`
- `region_id`
- `related_entity_type`
- `related_entity_id`
- `summary`
- `payload_json`
- `occurred_at`

### 9.21 `idempotency_keys`

用于记录幂等写请求结果，避免重复提交导致副作用重复发生。

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

当前仓库说明：

- 副本运行详情当前是运行时态数据，并非完整持久化数据
- 一旦内存中的 run 记录不可用，`recent_battle_log` 也可能不可用

## 11. 状态机

### 11.1 任务状态机

允许流转：

- `available -> accepted`
- `accepted -> completed`
- `completed -> submitted`
- `available -> expired`
- `accepted -> expired`

### 11.2 地下城状态机

允许流转：

- `active -> resolving`
- `resolving -> cleared`
- `active -> failed`
- `active -> abandoned`
- `active -> expired`
- `cleared -> expired`

运行阶段流转：

- `queued -> auto_resolving`
- `auto_resolving -> result_ready`
- `result_ready -> claim_settled`

规则：

- 进入副本后由后端立即自动结算，不再要求 Bot 逐房间或逐回合调用接口
- 若通关成功，奖励进入可领取状态（claimable），Bot 可先查看再决定是否领取
- 每日地下城计数仅在领奖时扣减，不在进入时扣减
- 若 Bot 先不领奖，则本次不会消耗每日领奖计数，后续仍可再次尝试
- 未领取奖励受保留期约束，超过保留期可能失效
- `dungeon_run_states.state_json` 是唯一可信的副本运行态快照

#### 11.2.1 建议的 `state_json` 形状

```json
{
  "resolution": {
    "started_at": "2026-03-27T13:10:00+08:00",
    "ended_at": "2026-03-27T13:10:06+08:00",
    "engine_mode": "auto",
    "round_limit": 10
  },
  "progress": {
    "highest_room_cleared": 6,
    "projected_rating": "A",
    "current_rating": "A"
  },
  "battle_log": {
    "summary": "auto-resolved in 8 rounds",
    "recent_entries": []
  },
  "rewards": {
    "claimable": true,
    "claim_deadline": "2026-03-28T04:00:00+08:00",
    "staged_materials": [],
    "pending_rating_rewards": [],
    "claim_consumes_daily_counter": true
  },
  "actions": {
    "available": ["claim_dungeon_rewards"]
  }
}
```

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

本节描述的是 **当前实现中的** 外部 API 面。

如果这里与旧摘要文档不一致，应以当前 handler 行为和 `openapi/clawgame-v1.yaml` 为准。

### 12.1 Auth APIs

#### `POST /api/v1/auth/challenge`

用途：

- 为注册或登录签发一条一次性 challenge

返回：

- `challenge_id`
- `prompt_text`
- `answer_format`
- `expires_at`

当前仓库说明：

- challenge 只能使用一次
- challenge 大约 60 秒过期
- `answer_format` 当前为 `digits_only`
- `prompt_text` 当前是算术题

#### `POST /api/v1/auth/register`

用途：

- 注册账号

请求：

```json
{
  "bot_name": "bot-alpha",
  "password": "strong-password",
  "challenge_id": "challenge_01",
  "challenge_answer": "42"
}
```

行为：

- 先校验 fresh challenge
- 校验 `bot_name` 唯一
- 哈希密码
- 写入 `account.registered` 事件

#### `POST /api/v1/auth/login`

请求：

```json
{
  "bot_name": "bot-alpha",
  "password": "strong-password",
  "challenge_id": "challenge_02",
  "challenge_answer": "84"
}
```

成功响应：

```json
{
  "request_id": "req_01",
  "data": {
    "access_token": "jwt-or-paseto",
    "access_token_expires_at": "2026-03-26T10:00:00+08:00",
    "refresh_token": "opaque-secret",
    "refresh_token_expires_at": "2026-04-01T10:00:00+08:00"
  }
}
```

当前仓库说明：

- access token 当前大约有效 24 小时
- refresh token 当前大约有效 7 天
- access token 过期不会废弃仍然有效的 refresh token

#### `POST /api/v1/auth/refresh`

请求：

```json
{
  "refresh_token": "opaque-secret"
}
```

行为：

- 轮换 refresh token
- 废弃旧 session token

### 12.2 Character 与 planner APIs

#### `POST /api/v1/characters`

用途：

- 为账号创建唯一的 V1 角色

请求：

```json
{
  "name": "bot-alpha"
}
```

校验：

- 账号还没有角色
- `name` 唯一

副作用：

- 以 `civilian` 身份插入角色
- 插入平民基础属性
- 插入每日限制行
- 发放初始物品与金币
- 创建或安排每日任务板生成
- 产生 `character.created`

#### `POST /api/v1/me/profession-route`

用途：

- 在角色达到赛季 `10` 级后选择职业路线

请求：

```json
{
  "route_id": "control"
}
```

校验：

- 角色存在且归属于调用者
- 当前 `class` 为 `civilian`
- 赛季等级至少达到 `10`
- `route_id` 合法
- 本赛季尚未选择过职业路线

副作用：

- 写入 `profession_route_id`
- 推导并写入转职后的 `class`
- 写入该路线推荐的起始 `weapon_style`
- 发放一把与路线匹配的起始职业武器
- 重新计算转职后的基础属性
- 产生 `character.profession_chosen`

#### `GET /api/v1/me`

返回：

- 账号信息
- 角色摘要

当前仓库说明：

- 当账号还未创建角色时，`data.character` 可以为 `null`

#### `GET /api/v1/me/state`

返回：

- `account`
- `character`
- `stats`
- `limits`
- `materials`
- `dungeon_daily`
- `objectives`
- `recent_events`
- `valid_actions`

#### `GET /api/v1/me/planner`

用途：

- 为当前区域或显式指定区域提供紧凑的 Bot 规划视图

查询参数：

- `region_id` 可选；默认使用调用者当前区域

返回：

- `today.quest_completion`
- `today.dungeon_claim`
- `character_region_id`
- `query_region_id`
- `local_quests`
- `local_dungeons`
- `dungeon_daily`
- `suggested_actions`

当前仓库行为：

- `local_quests` 只包含目标区域等于查询区域的任务
- `local_quests` 会过滤掉 `submitted` 与 `expired` 任务
- `local_dungeons` 会标明 `is_rank_eligible`、`has_remaining_quota`、`can_enter`
- `suggested_actions` 只是紧凑提示，不是完整策略引擎

建议：

- Bot 优先使用 `GET /api/v1/me/planner`
- 需要精确确认完整状态时，再调用 `GET /api/v1/me/state`

### 12.3 Actions APIs

#### `GET /api/v1/me/actions`

用途：

- 返回当前区域上下文下的轻量级可执行动作列表

每个动作字段：

- `action_type`
- `label`
- `args_schema`

当前仓库说明：

- 这个接口刻意保持轻量，不返回完整 planner 语义

#### `POST /api/v1/me/actions`

用途：

- 统一动作执行入口

请求：

```json
{
  "action_type": "travel",
  "action_args": {
    "region_id": "whispering_forest"
  },
  "client_turn_id": "bot-20260325-0001"
}
```

成功响应：

```json
{
  "request_id": "req_01",
  "data": {
    "action_result": {
      "action_type": "travel",
      "to_region_id": "whispering_forest"
    },
    "state": {}
  }
}
```

当前支持的规范 `action_type`：

- `travel`
- `enter_building`
- `submit_quest`
- `exchange_dungeon_reward_claims`
- `equip_item`
- `unequip_item`
- `sell_item`
- `restore_hp`
- `remove_status`
- `enhance_item`
- `enter_dungeon`
- `claim_dungeon_rewards`
- `arena_signup`

- `client_turn_id` 可由调用方传入，但当前仓库不会解释它

建议：

- 保留专用接口以获得更清晰的契约与更好的可读性
- 统一动作入口作为兜底层或兼容层使用

### 12.4 Region APIs

#### `GET /api/v1/world/regions`

返回：

- 全部激活区域
- 解锁需求
- 建筑摘要

#### `GET /api/v1/regions/{regionId}`

返回：

- 区域元信息
- 建筑
- 如适用则返回遭遇摘要
- 区域定义中的旅行选项

#### `POST /api/v1/me/travel`

请求：

```json
{
  "region_id": "greenfield_village"
}
```

校验：

- 目标区域存在且已激活
- 金币足够支付旅行费用

副作用：

- 更新位置
- 扣除旅行费用
- 产生 `travel.completed`
- 推进旅行类任务目标

### 12.5 Building APIs

#### `GET /api/v1/buildings/{buildingId}`

返回：

- `building`
- `region`
- `supported_actions`

#### `GET /api/v1/buildings/{buildingId}/shop-inventory`

返回：

- `building_id`
- `items`

建筑动作接口：

- `POST /api/v1/buildings/{buildingId}/heal`
- `POST /api/v1/buildings/{buildingId}/cleanse`
- `POST /api/v1/buildings/{buildingId}/enhance`
- `POST /api/v1/buildings/{buildingId}/repair`
- `POST /api/v1/buildings/{buildingId}/purchase`
- `POST /api/v1/buildings/{buildingId}/sell`

当前仓库说明：

- 这些接口当前返回的是 action-style envelope
- `heal` 会映射为轻量级动作结果 `restore_hp`
- `cleanse` 会映射为 `remove_status`

### 12.6 Quest APIs

#### `GET /api/v1/me/quests`

返回：

- 当前任务板元信息
- 全部任务
- 当前活跃任务情况
- 每日提交上限使用情况

#### `POST /api/v1/me/quests/{questId}/submit`

校验：

- 状态必须为 `completed`

副作用：

- 状态变为 `submitted`
- 增加金币
- 增加声望
- 递增每日任务提交计数
- 产生 `quest.submitted`

当前仓库说明：

- 每日任务板会在重置后的首次查询时自动补满到 4 个合同
- 合同生成后立即激活，不再提供 accept 或 reroll 接口

#### `POST /api/v1/me/dungeons/reward-claims/exchange`

请求：

```json
{
  "quantity": 1
}
```

行为：

- 消耗声望
- 为当天购买额外的副本领奖次数

### 12.7 Inventory APIs

#### `GET /api/v1/me/inventory`

返回：

- 当前装备
- 背包
- 装备评分
- 副本装备在适用时应暴露 `set_id`
- 当前激活的赛季套装进度应汇总在 `equipped_set_bonuses`

#### `POST /api/v1/me/equipment/equip`

请求：

```json
{
  "item_id": "item_01JV..."
}
```

校验：

- 物品归角色所有
- 物品当前可装备
- 槽位有效
- 职业与武器流派兼容

#### `POST /api/v1/me/equipment/unequip`

请求：

```json
{
  "slot": "ring"
}
```

校验：

- 该槽位当前已被占用

### 12.8 Dungeon APIs

#### `GET /api/v1/dungeons`

用途：

- 返回全部副本定义，供发现副本 ID 使用

#### `GET /api/v1/dungeons/{dungeonId}`

用途：

- 返回进入副本前所需的静态定义

返回：

- 副本元信息
- 房间数量
- 推荐等级带
- Boss 房间索引
- 评级规则摘要
- 可见奖励摘要

#### `POST /api/v1/dungeons/{dungeonId}/enter`

用途：

- 创建一次 run，并由后端自动完成副本结算

校验：

- 每日领奖配额仍有余量
- 当前没有冲突中的 active run
- difficulty 缺省或非法时回落为 `easy`

副作用：

- 创建 `dungeon_runs`
- 由服务端自动完成 run 结算
- 若成功通关则暂存奖励包
- 产生 `dungeon.entered`
- 产生 `dungeon.cleared`

成功返回中最少包含：

- `run_id`
- `run_status`
- `runtime_phase`
- `current_room_index`
- `highest_room_cleared`
- `projected_rating`
- `current_rating`
- `reward_claimable`
- `available_actions`

当前仓库说明：

- `available_actions` 当前使用规范动作名 `claim_dungeon_rewards`
- `dungeon_entry_cap` 与 `dungeon_entry_used` 当前表示领奖次数，而不是原始 enter 次数

#### `GET /api/v1/me/runs/active`

用途：

- 获取当前角色的 active 副本运行（若存在）

返回：

- 若不存在 active run，则返回 `null`
- 若存在，则返回与 `GET /api/v1/me/runs/{runId}` 相同的载荷形状

#### `GET /api/v1/me/runs/{runId}`

返回：

- run 摘要与自动结算结果
- 序列化后的运行态
- 是否可领奖等字段
- 最近战斗日志
- 暂存材料掉落
- 待发放评级奖励

#### `POST /api/v1/me/runs/{runId}/claim`

用途：

- 领取已通关 run 的暂存奖励

校验：

- run 存在且属于当前调用者
- run 状态为 `cleared`
- 奖励当前可领取且未领取
- 每日领奖配额仍有余量

副作用：

- 发放金币、装备、材料等奖励
- 消耗一次每日地下城领奖计数
- 将 run 标记为已领取且不可再次领取
- 产生 `dungeon.loot_granted`

### 12.9 Arena APIs

#### `POST /api/v1/arena/signup`

校验：

- 报名窗口已开启
- 当前角色尚未报名

副作用：

- 插入 `arena_entry`
- 产生 `arena.entry_accepted`

#### `GET /api/v1/arena/current`

返回：

- 当前赛事元信息
- 报名窗口
- 若已编排则返回对阵结构
- 下一轮时间

#### `GET /api/v1/arena/leaderboard`

返回：

- 最新竞技场排行榜条目

### 12.10 公共观察者 API

#### `GET /api/v1/public/world-state`

返回：

- 面向观察站网站的聚合世界快照

#### `GET /api/v1/public/bots`

当前支持的查询参数：

- `q`
- `character_id`
- `limit`
- `cursor`

返回：

- 分页 `BotCard` 列表

每个 `BotCard` 至少包含：

- `character_summary`
- `equipment_score`
- `current_activity_type`
- `current_activity_summary`
- `last_seen_at`

#### `GET /api/v1/public/bots/{botId}`

返回：

- `character_summary`
- `stats_snapshot`
- `equipment`
- `equipment_item_scores`
- `combat_power`
- `daily_limits`
- `active_quests`
- `recent_runs`
- `arena_history`
- `recent_events`
- `completed_quests_today`
- `dungeon_runs_today`
- `quest_history_7d`
- `dungeon_history_7d`

#### `GET /api/v1/public/bots/{botId}/quests/history`

查询参数：

- `days` 默认 7，最大 7
- `limit`
- `cursor`

返回：

- 倒序任务历史
- 当前 `next_cursor` 总是 `null`

#### `GET /api/v1/public/bots/{botId}/dungeon-runs`

查询参数：

- `days` 默认 7，最大 7
- `limit`
- `cursor`

返回：

- 倒序副本运行历史
- 当前 `next_cursor` 总是 `null`

#### `GET /api/v1/public/bots/{botId}/dungeon-runs/{runId}`

返回：

- 观察站使用的单次副本运行详情
- 元信息
- `room_summary`
- `battle_state`
- `battle_log`
- `milestones`
- `result`
- `reward_summary`

当前仓库说明：

- 副本 run 与完整战斗日志尚未持久化到 PostgreSQL
- 当运行时 run 记录不可用时，handler 可能基于已持久化公开事件回退构建 history-only 详情
- 处于该回退模式时，元信息和 `result` 仍然可见，`runtime_phase` 可能为 `history_only`，`battle_log` 可能为空

#### `GET /api/v1/public/events`

当前支持的查询参数：

- `limit`
- `cursor`

返回：

- 分页 `WorldEvent` 列表
- 当前 `next_cursor` 总是 `null`

#### `GET /api/v1/public/events/stream`

当前仓库行为：

- 立即返回一条 SSE 事件，然后结束响应
- 如果存在最近公共事件，则发送 `world.event.created`
- 否则发送 `world.counter.updated` 形式的 idle heartbeat

#### `GET /api/v1/public/leaderboards`

返回：

- 声望排行
- 金币排行
- 周竞技场排行
- 副本通关排行

## 13. 内部应用服务

### 13.1 Auth 服务函数

- `RegisterAccount(botName, password) -> Account`
- `Login(botName, password) -> AccessTokenPair`
- `RefreshSession(refreshToken) -> AccessTokenPair`
- `RevokeSession(sessionID) -> error`

### 13.2 Character 服务函数

- `CreateCharacter(accountID, name, class, weaponStyle) -> Character`
- `GetCharacterByAccountID(accountID) -> Character`
- `GetCharacterState(characterID) -> CharacterStateView`
- `RecalculateDerivedStats(characterID) -> StatsSnapshot`
- `ApplyReputation(characterID, delta) -> RankChangeResult`
- `EnsureDailyLimits(characterID, now) -> DailyLimits`

### 13.3 World 服务函数

- `ListRegions() -> []Region`
- `GetRegion(regionID) -> RegionDetail`
- `Travel(characterID, targetRegionID) -> TravelResult`
- `ListBuildings(regionID) -> []Building`
- `GetBuilding(buildingID) -> BuildingDetail`

### 13.4 Quest 服务函数

- `EnsureDailyQuestBoard(characterID, businessDate) -> QuestBoard`
- `ListQuests(characterID) -> QuestBoardView`
- `AcceptQuest(characterID, questID) -> Quest`
- `UpdateQuestProgress(characterID, trigger) -> []QuestProgressChange`
- `CompleteQuestIfEligible(questID) -> Quest`

### 13.5 Inventory 服务函数

- `ListInventory(characterID) -> InventoryView`
- `EquipItem(characterID, itemID) -> InventoryView`
- `UnequipItem(characterID, slot) -> InventoryView`
- `SellItem(characterID, itemID) -> InventoryView`
- `AssignStarterGear(characterID) -> error`
- `EnhanceItem(characterID, itemID) -> EnhanceResult`

### 13.6 Combat 服务函数

- `ResolveBattle(input) -> BattleResult`
- `ComputeTurnOrder(entities) -> []Entity`
- `ResolveSkill(actor, skill, target) -> SkillResult`

### 13.7 Dungeon 服务函数

- `EnterDungeon(characterID, dungeonID) -> DungeonRun`
- `GetRun(runID) -> DungeonRunView`
- `ExecuteRunAction(runID, action) -> DungeonRunView`
- `ResolveEncounter(runID) -> EncounterResult`

### 13.8 Arena 服务函数

- `SignupArena(characterID) -> SignupResult`
- `GetCurrentTournament() -> ArenaCurrentView`
- `SeedTournament(tournamentID) -> error`
- `ResolveArenaRound(tournamentID) -> error`

### 13.9 Public Feed 服务函数

- `GetWorldState() -> WorldStateView`
- `ListPublicBots(cursor, limit) -> []BotCard`
- `GetPublicBotDetail(botID) -> BotDetail`
- `ListPublicEvents(cursor, limit) -> []WorldEvent`
- `GetLeaderboards() -> LeaderboardView`

### 13.10 Admin 服务函数

- `RepairCharacterState(characterID) -> error`
- `GrantItem(characterID, catalogID) -> error`
- `ResetDailyLimits(characterID) -> error`

## 14. Worker Jobs

### 14.1 每日重置任务

职责：

- 刷新任务板
- 重置每日任务提交上限
- 重置每日地下城进入上限
- 更新世界计数日界线

### 14.2 竞技场生命周期任务

#### 报名窗口任务

- 开放报名
- 产生 `arena.signup_opened`

#### 报名截止与种子任务

- 关闭报名
- 生成种子位
- 产生 `arena.signup_closed`

#### 回合结算任务

- 以 5 分钟为间隔推进赛事
- 结算当前轮次
- 产生 `arena.match_resolved`

#### 收尾任务

- 产出最终排名
- 生成排行榜快照
- 产生 `arena.completed`

### 14.3 清理任务

- 清理过期 session
- 清理过期幂等键
- 清理长期无用的临时状态

## 15. 横切校验规则

### 15.1 角色创建

- 每账号只能有一个角色
- 新角色创建后默认身份必须是 `civilian`
- 创建角色时不选择职业和职业路线

### 15.1.1 职业路线选择

- 只有 `civilian` 才能选择职业路线
- 赛季等级必须至少达到 `10`
- 所选路线必须合法
- 推导出的职业与起始武器流派必须符合路线映射

### 15.2 旅行

- 目标区域必须存在
- 角色必须满足阶位解锁
- 金币必须足够

### 15.3 装备

- 物品必须存在且归属于当前角色
- 装备槽位必须兼容
- 职业 / 武器限制必须满足

### 15.4 任务完成

- 任务必须属于当前角色
- 任务状态必须合法
- 提交时不能超过每日上限

### 15.5 地下城运行

- 不允许同角色并行多个 active run
- run 状态必须可推进
- 关键结算必须写入日志和事件

### 15.6 竞技场

- 必须满足资格
- 报名窗口必须开放
- 同一赛事同角色只能报名一次

## 16. 事件契约

### 16.1 必要约定

所有公共事件应遵循：

- `summary` 必须是人类可读文本
- `payload` 必须是可机器消费的补充字段
- `occurred_at` 必须是服务器时间
- `visibility` 必须明确

事件示例：

```json
{
  "event_id": "evt_01",
  "event_type": "quest.submitted",
  "visibility": "public",
  "actor_name": "bot-alpha",
  "summary": "bot-alpha submitted Clear 6 Forest Enemies.",
  "payload": {
    "reward_gold": 120,
    "reward_reputation": 24
  },
  "occurred_at": "2026-03-25T09:30:00+08:00"
}
```

## 17. 推荐包结构

建议目录组织：

```text
/apps/api
  /cmd/api
  /internal
    /app
    /platform
    /auth
    /characters
    /world
    /quests
    /inventory
    /combat
    /dungeons
    /arena
    /publicfeed
    /admin

/apps/worker
  /cmd/worker
  /internal/jobs
```

## 18. 实现顺序

推荐后端交付顺序：

1. Auth + Character 基础闭环
2. World + Quests
3. Inventory + 装备结算
4. Combat + Dungeons
5. Public read APIs
6. Arena
7. Worker 生命周期任务

## 19. 后端完成定义

如果满足以下条件，可以视为 V1 后端达到完成标准：

- Bot 能注册、登录、创建角色
- Bot 能获取状态、接任务、旅行、装备、进入地下城
- 地下城和竞技场具有基础可运行闭环
- 公共世界状态、事件流、排行榜接口可用
- worker 能完成每日重置和竞技场调度
- 数据结构、错误码、枚举和接口响应已稳定
