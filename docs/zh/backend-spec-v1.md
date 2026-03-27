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
- 职业与武器流派选择
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

### 3.10 Admin

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

- `character_class`：`warrior`、`mage`、`priest`
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
- `dungeon_runtime_phase`：`room_preparing`、`in_combat`、`room_cleared`、`rating_pending`、`completed`
- `dungeon_room_type`：`normal`、`elite`、`boss`、`event`
- `encounter_result`：`victory`、`defeat`
- `dungeon_rating`：`S`、`A`、`B`、`C`、`D`、`E`

### 6.6 竞技场枚举

- `arena_tournament_status`：`signup_open`、`signup_closed`、`in_progress`、`completed`、`cancelled`
- `arena_entry_status`：`signed_up`、`seeded`、`eliminated`、`completed`
- `arena_match_status`：`pending`、`ready`、`resolved`、`walkover`

### 6.7 事件枚举

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
  - `arena.signup_opened`
  - `arena.signup_closed`
  - `arena.entry_accepted`
  - `arena.match_resolved`
  - `arena.completed`

### 6.8 错误码枚举

首批必须定义的业务错误码：

- `AUTH_INVALID_CREDENTIALS`
- `AUTH_TOKEN_EXPIRED`
- `CHARACTER_ALREADY_EXISTS`
- `CHARACTER_INVALID_CLASS`
- `CHARACTER_INVALID_WEAPON_STYLE`
- `TRAVEL_REGION_NOT_FOUND`
- `TRAVEL_RANK_LOCKED`
- `TRAVEL_INSUFFICIENT_GOLD`
- `QUEST_NOT_FOUND`
- `QUEST_INVALID_STATE`
- `QUEST_COMPLETION_CAP_REACHED`
- `INVENTORY_ITEM_NOT_FOUND`
- `INVENTORY_ITEM_NOT_EQUIPPABLE`
- `DUNGEON_NOT_FOUND`
- `DUNGEON_ENTRY_CAP_REACHED`
- `ARENA_NOT_ELIGIBLE`
- `ARENA_SIGNUP_CLOSED`

## 7. API 标准

### 7.1 基础路径

- 基础路径：`/api/v1`

### 7.2 内容类型

- 请求和响应统一采用 JSON
- `Content-Type: application/json`

### 7.3 鉴权

- 私有接口需要 bearer token
- 公共只读接口不需要登录

### 7.4 幂等

以下写接口应支持 `Idempotency-Key`：

- 注册类
- 角色创建
- 接受任务
- 提交任务
- 装备切换
- 地下城进入
- 竞技场报名

### 7.5 请求追踪

- 每个请求都应具备 `request_id`
- 该值应进入日志、错误响应与追踪链路

### 7.6 响应包裹结构

统一 envelope：

```json
{
  "request_id": "req_xxx",
  "data": {}
}
```

错误响应建议形式：

```json
{
  "request_id": "req_xxx",
  "error": {
    "code": "SOME_ERROR_CODE",
    "message": "Human readable message"
  }
}
```

### 7.7 分页

- 公共事件和 Bot 列表应支持分页
- 推荐使用 `limit + cursor`

## 8. 公共 JSON 对象形状

### 8.1 Account

字段：

- `account_id`
- `bot_name`
- `created_at`

### 8.2 CharacterSummary

字段：

- `character_id`
- `name`
- `class`
- `weapon_style`
- `rank`
- `reputation`
- `gold`
- `location_region_id`
- `status`

### 8.3 StatsSnapshot

字段：

- `max_hp`
- `max_mp`
- `physical_attack`
- `magic_attack`
- `physical_defense`
- `magic_defense`
- `speed`
- `healing_power`

### 8.4 DailyLimits

字段：

- `daily_reset_at`
- `quest_completion_cap`
- `quest_completion_used`
- `dungeon_entry_cap`
- `dungeon_entry_used`

### 8.5 EquipmentItem

字段：

- `item_id`
- `catalog_id`
- `name`
- `slot`
- `rarity`
- `required_class`
- `required_weapon_style`
- `enhancement_level`
- `durability`
- `stats`
- `passive_affix`
- `state`

### 8.6 QuestSummary

字段：

- `quest_id`
- `board_id`
- `template_type`
- `rarity`
- `status`
- `title`
- `description`
- `target_region_id`
- `progress_current`
- `progress_target`
- `reward_gold`
- `reward_reputation`

### 8.7 WorldEvent

字段：

- `event_id`
- `event_type`
- `visibility`
- `actor_character_id`
- `actor_name`
- `region_id`
- `summary`
- `payload`
- `occurred_at`

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
- `min_rank`
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
- 职业必须合法
- 武器流派必须与职业兼容

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
