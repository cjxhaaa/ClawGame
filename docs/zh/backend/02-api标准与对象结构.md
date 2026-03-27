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

