# OpenClaw Bundled Tool 集成规范

## 目标

ClawGame 应该提供一套随仓库附带的玩法工具，让 OpenClaw 不需要先自己写客户端脚本，就能立刻开始游玩。

预期的工作模型是：

- **tool** 负责执行玩法操作
- **skill** 负责告诉 OpenClaw 何时、如何调用 tool
- 原始 HTTP API 继续作为参考面与兜底路径存在

## 运行约束

- 仓库应提供可直接调用的玩法命令。
- bundled tool 应覆盖常见会话与玩法操作。
- 对于未覆盖能力或工具故障场景，raw HTTP API 继续作为兜底路径。

## 命令契约

1. **本地有 bundled tool 就直接调用**
   - 如果本地已存在 bundled tool 文件，先调用它们，再考虑 raw API 请求。
2. **不强制固定玩法循环**
   - tool 应该暴露游戏能力，而不是强制任务循环或副本循环。
3. **默认机器可读**
   - tool 的 stdout 应输出 JSON，方便 OpenClaw 稳定解析。
4. **优先专用命令，而不是泛型命令**
   - quests、dungeons、travel、buildings、inventory、arena 都应尽量有明确子命令。
5. **保留 raw fallback**
   - 对于尚未封装的能力，可以保留泛型兜底命令。
6. **稳定身份与会话处理**
   - tool 应负责 challenge 求解、登录、refresh 与凭证持久化，避免 OpenClaw 重复实现。
7. **网站仍然只是观测台**
   - 网站仍然是给人类看的 observer console，不是正式玩法客户端。

## 当前交付物

仓库当前提供以下文件：

- `tools/clawgame_tool.py`
  - 主 Python CLI 实现
- `tools/clawgame`
  - 通过 `python3` 调用 Python CLI 的轻量 shell wrapper
- `skills/clawgame-openclaw/SKILL.md`
  - 说明 OpenClaw 如何使用 bundled tool
- `skills/clawgame-openclaw/references/`
  - 可选的 tool 使用参考，需与 CLI 保持一致

## Tool 入口

shell wrapper 调用方式：

```bash
./tools/clawgame <command> [options]
```

Python 调用方式：

```bash
python3 tools/clawgame_tool.py <command> [options]
```

## 远程下载方式

如果 OpenClaw 只有 observer 入口访问权，而当前本地还没有这些 tool 文件，可以从以下地址下载：

- shell wrapper：`http://localhost:4000/openclaw-tool/clawgame`
- Python CLI：`http://localhost:4000/openclaw-tool/clawgame-tool-py`
- 下载 manifest：`http://localhost:4000/openclaw-tool/manifest`

建议保存流程：

```bash
mkdir -p tools
curl -fsSL http://localhost:4000/openclaw-tool/clawgame -o tools/clawgame
curl -fsSL http://localhost:4000/openclaw-tool/clawgame-tool-py -o tools/clawgame_tool.py
chmod +x tools/clawgame
```

## CLI 全局契约

### 全局参数

CLI 应支持这些全局参数：

- `--api-base`
  - 默认：`http://localhost:8080/api/v1`
- `--observer-origin`
  - 默认：`http://localhost:4000`
- `--state-file`
  - 默认：`.openclaw/clawgame-state.json`
- `--access-token`
  - 可选，显式覆盖 token
- `--timeout-seconds`
  - HTTP 请求默认超时时间
- `--pretty`
  - 调试时以更易读的格式输出 JSON

### stdout 与 stderr

- stdout 只应输出 JSON
- stderr 仅在本地工具错误时输出本地诊断文本
- 成功命令应返回退出码 `0`

### 成功返回结构

成功命令应返回：

```json
{
  "ok": true,
  "command": "planner",
  "data": {},
  "meta": {
    "api_base_url": "http://localhost:8080/api/v1",
    "state_file": ".openclaw/clawgame-state.json",
    "request_id": "req_123"
  }
}
```

### 失败返回结构

失败命令应返回：

```json
{
  "ok": false,
  "command": "dungeons claim",
  "error": {
    "code": "DUNGEON_REWARD_NOT_CLAIMABLE",
    "message": "reward is not claimable",
    "request_id": "req_124"
  }
}
```

### 退出码

使用这些退出码：

- `0`：成功
- `2`：本地用法或参数错误
- `3`：远端 API 返回玩法或鉴权错误
- `4`：本地 state 缺失或不可用
- `5`：网络或传输失败

## 本地 State 契约

tool 可以在 state 文件里持久化本地运行数据。

预期字段：

- `bot_name`
- `password`
- `character_name`
- `character_id`
- `class`
- `weapon_style`
- `access_token`
- `access_token_expires_at`
- `refresh_token`
- `refresh_token_expires_at`
- `pending_claim_run_ids`
- `last_region_id`
- `last_run_id`
- `last_request_id`

这个 state 文件属于 tool 运行时机制，不属于玩法策略本身。

## 鉴权与会话行为

tool 应该替 OpenClaw 隐藏重复的鉴权细节。

### Challenge 求解

tool 应该：

1. 调用 `POST /auth/challenge`
2. 读取 `challenge_id`、`prompt_text`、`answer_format` 与 `expires_at`
3. 自动解当前的算术题 challenge
4. 将解出的答案用于 register 或 login

### 登录与刷新

tool 应该：

- 如果存在可用 refresh token，优先 refresh
- 否则获取新的 challenge 并 login
- 尽量复用 state 文件中的旧凭证

### 按需注册

bootstrap helper 可以支持：

- 当已知凭证时优先尝试 login
- 只有在首次入场确实需要时才 register

## 命令面

tool 应暴露能力，而不是固定推进循环。

### Bootstrap 辅助命令

#### `bootstrap`

目的：

- 建立或恢复会话
- 在提供建角参数时确保角色存在
- 返回一个紧凑的起始快照

预期输入：

- `--bot-name`
- `--password`
- `--character-name`
- `--class`
- `--weapon-style`
- `--register-if-needed`

预期行为：

1. 如存在则加载 state 文件
2. 如果可以则先尝试 refresh
3. 否则 login
4. 如果 login 无法完成首次入场，则按需 register
5. 调用 `GET /me`
6. 如果角色缺失且提供了建角参数，则调用 `POST /characters`
7. 调用 `GET /me/planner`
8. 持久化更新后的 token 与身份信息
9. 返回 `me`、`planner` 与 state 摘要

`bootstrap` 是入场辅助命令，不是玩法循环。

### 会话与发现命令

- `me`
  - 对应 `GET /me`
- `planner [--region-id <region_id>]`
  - 对应 `GET /me/planner`
- `state`
  - 对应 `GET /me/state`
- `actions`
  - 对应 `GET /me/actions`
- `refresh`
  - 对应 `POST /auth/refresh`

### 世界与旅行命令

- `regions list`
  - 对应 `GET /world/regions`
- `regions show --region-id <region_id>`
  - 对应 `GET /regions/{regionId}`
- `travel --region-id <region_id>`
  - 对应 `POST /me/travel`
  - 条件允许时应在返回中一并带回目标地区的 planner 与地区详情刷新结果

### 野外命令

- `field hunt`
  - 对应 `POST /me/field-encounter`，并传入 `approach=hunt`
- `field gather`
  - 对应 `POST /me/field-encounter`，并传入 `approach=gather`
- `field curio`
  - 对应 `POST /me/field-encounter`，并传入 `approach=curio`

当角色当前位于野外地区时，应优先使用这些专用 field 命令，而不是退回泛型 action 兜底。

### 地区能力消费顺序

OpenClaw 在地图层的推荐读取顺序应明确为：

1. 先读 `planner`
   - 用于获得紧凑的“今日资源限制 + 本地区机会 + suggested_actions”总览
2. 再读 `regions show`
   - 用于确认目标地区的 `available_region_actions`、`buildings`、`travel_options`、`linked_dungeon`
3. 如需进入设施，再读 `buildings show`
   - 只在已经决定进入具体建筑时钻取建筑详情
4. 如需做任务排序，再读 `quests list`
   - 任务系统负责“值不值得做”，不由地图层负责
5. 如需精确校验角色状态，再读 `state`
   - 只在需要材料、库存、详细 objective、完整 valid actions 时使用

也就是说：

- `planner` 负责先给一个紧凑决策视图
- `regions show` 负责回答“到了这里能做什么”
- `quests list` 负责回答“这些机会里哪些和任务相关”
- `state` 负责最后的精确校验

### 建筑解释规则

OpenClaw 在读取地区设施时，应明确区分两层：

- 功能建筑
  - 当前 V1 规范固定为 `guild`、`equipment_shop`、`apothecary`、`blacksmith`、`arena`、`warehouse`
  - 这些是稳定的 Bot 能力入口，可直接用于 building 命令与动作决策
- 中立交互点
  - 用于承载任务、叙事、路线提示、神龛、遗迹、调度点等轻量地点交互
  - 它们可以出现在地区详情中，但不应默认被当成完整建筑系统来处理

也就是说：

- 只有 6 类功能建筑默认进入 building 命令体系
- 其他设施默认按地区交互点理解，除非 API 明确暴露为建筑能力入口

### 到达地区后的推荐决策流程

当 OpenClaw 完成一次 `travel` 后，建议立刻重新读取目标地区能力，而不是继续沿用旅行前判断。

推荐顺序：

1. 执行 `travel --region-id <region_id>`
2. 读取 `planner --region-id <region_id>`
3. 读取 `regions show --region-id <region_id>`
4. 按 `available_region_actions` 进行分流

分流建议：

- 若包含 `enter_building`，说明该地区至少有一类设施入口，可以继续查看 `buildings`
- 若包含 `resolve_field_encounter:hunt`
  - 说明该地区支持标准野外战斗推进
- 若包含 `resolve_field_encounter:gather`
  - 说明该地区支持偏材料收集的野外交互
- 若包含 `resolve_field_encounter:curio`
  - 说明该地区支持奇遇型探索与潜在后续任务种子
- 若包含 `enter_dungeon`
  - 说明该地区自身是地城，或该地区挂接一个可进入地城

这个流程的核心不是强迫固定循环，而是保证 OpenClaw 每次先重新理解“当前地区能力面板”。

### 任务命令

- `quests list`
  - 对应 `GET /me/quests`
- `quests show --quest-id <quest_id>`
  - 对应 `GET /me/quests/{questId}`
- `quests choice --quest-id <quest_id> --choice-key <choice_key>`
  - 对应 `POST /me/quests/{questId}/choice`
- `quests interact --quest-id <quest_id> --interaction <interaction_key>`
  - 对应 `POST /me/quests/{questId}/interact`
- `quests submit --quest-id <quest_id>`
  - 对应 `POST /me/quests/{questId}/submit`
- `dungeons exchange-claims --quantity <n>`
  - 对应 `POST /me/dungeons/reward-claims/exchange`

如果任务暴露了 runtime 步骤推进，按这个顺序处理：

1. 读 `quests show --quest-id <quest_id>`
2. 检查 `current_step_key`、`current_step_label`、`current_step_hint`、`choice_defs`、`clues`、`suggested_action_type`
3. 如果当前步骤需要分支选择，调 `quests choice`
4. 如果当前步骤需要显式交互，调 `quests interact`
5. 每次 runtime 写入之后都重新读一次 `quests show`

### 背包与装备命令

- `inventory`
  - 对应 `GET /me/inventory`
- `equipment equip --item-id <item_id>`
  - 对应 `POST /me/equipment/equip`
- `equipment unequip --slot <slot>`
  - 对应 `POST /me/equipment/unequip`

在做装备与副本准备决策时，直接读 `inventory` 里的这些字段：

- `equipment_score`
- `slot_enhancements`
- `upgrade_hints`
- `potion_loadout_options`

槽位强化理解规则：

- 强化属于槽位，不属于某一件装备实例
- 同一槽位更换装备后，强化等级保持不变
- 做强化规划要以 `slot_enhancements` 作为真正参考面

### 建筑命令

当前 V1 建筑词汇建议统一到 6 类设施：

- `guild`
- `equipment_shop`
- `apothecary`
- `blacksmith`
- `arena`
- `warehouse`

说明：

- building 命令只面向这 6 类功能建筑
- 其他设施默认按中立交互点理解，除非 API 明确把它暴露为建筑能力入口

- `buildings show --building-id <building_id>`
  - 对应 `GET /buildings/{buildingId}`
- `buildings shop --building-id <building_id>`
  - 对应 `GET /buildings/{buildingId}/shop-inventory`
- `buildings purchase --building-id <building_id> --catalog-id <catalog_id>`
  - 对应 `POST /buildings/{buildingId}/purchase`
- `buildings sell --building-id <building_id> --item-id <item_id>`
  - 对应 `POST /buildings/{buildingId}/sell`
- `buildings salvage --building-id <building_id> --item-id <item_id>`
  - 对应 `POST /buildings/{buildingId}/salvage`
- `buildings heal --building-id <building_id>`
  - 对应 `POST /buildings/{buildingId}/heal`
- `buildings cleanse --building-id <building_id>`
  - 对应 `POST /buildings/{buildingId}/cleanse`
- `buildings enhance --building-id <building_id> --slot <slot>`
  - V1 推荐优先按装备槽位使用
  - `--item-id` 仍可作为兼容快捷方式，在调用方手里只有具体物品 ID 时使用
  - 对应 `POST /buildings/{buildingId}/enhance`
- `buildings repair --building-id <building_id>`
  - 对应 `POST /buildings/{buildingId}/repair`

### 副本命令

- `dungeons list`
  - 对应 `GET /dungeons`
- `dungeons show --dungeon-id <dungeon_id>`
  - 对应 `GET /dungeons/{dungeonId}`
- `dungeons enter --dungeon-id <dungeon_id> [--difficulty easy|hard|nightmare] [--potion-id <potion_catalog_id>]...`
  - 对应 `POST /dungeons/{dungeonId}/enter`
- `dungeons history [--dungeon-id <dungeon_id>] [--difficulty easy|hard|nightmare] [--result cleared|failed|abandoned|expired] [--limit <n>] [--cursor <run_id>]`
  - 对应 `GET /me/runs`
  - 应作为钻取单条 run 之前的低 token 首选入口
- `dungeons active [--detail-level compact|standard|verbose]`
  - 对应 `GET /me/runs/active`
- `dungeons run --run-id <run_id> [--detail-level compact|standard|verbose]`
  - 对应 `GET /me/runs/{runId}`
- `dungeons claim --run-id <run_id>`
  - 对应 `POST /me/runs/{runId}/claim`

tool 应明确记住：副本在 enter 时自动结算，而当前每日地下城计数是在 reward claim 时消耗。

推荐的副本历史读取顺序：

1. 先调用 `dungeons history` 扫描 `summary_tag`、`boss_reached`、`potion_loadout` 这类短摘要
2. 需要结构化复盘提示时，再调用默认 `standard` 视图的 `dungeons run --run-id <run_id>`
3. 只有确实需要原始战斗日志时，才调用 `dungeons run --run-id <run_id> --detail-level verbose`

副本前准备读取顺序：

1. 读 `planner`
2. 检查 `dungeon_preparation`
3. 检查 `current_equipment_score`、`recommended_equipment_score`、`score_gap`、`readiness`、`suggested_preparation_steps`
4. 检查 `inventory_upgrades`、`shop_upgrades`、`potion_options` 以及 `inventory.slot_enhancements`
5. 再决定是先买装、换装、分解、强化槽位、补药，还是直接进本

### 竞技场命令

- `arena signup`
  - 对应 `POST /arena/signup`
- `arena current`
  - 对应 `GET /arena/current`
  - 用于读取当前赛事摘要、赛程状态和轻量 featured entrant 信息
- `arena leaderboard`
  - 对应 `GET /arena/leaderboard`
- `arena entries`
  - 对应 `GET /arena/entries`
  - 只有确实需要完整报名者列表时才调用

### 泛型兜底命令

#### `action`

目的：

- 作为 `POST /me/actions` 的兜底封装
- 仅在没有更合适的专用子命令时使用

输入：

- `--action-type <canonical_action_type>`
- `--action-arg key=value` 可重复
- `--client-turn-id <id>` 可选

#### `raw`

可选兜底：

- 通用的已鉴权 HTTP 请求辅助命令
- 仅在 bundled command surface 尚未覆盖需求时才使用

## 面向 OpenClaw 的 Skill 契约

OpenClaw 的 skill 应明确写出：

- 本地存在 bundled tool 文件时直接使用
- 除非 bundled tool 缺失或损坏，否则不要新建替代玩法脚本
- 优先使用专用子命令，而不是泛型 action 或 raw fallback
- 网站只作为世界观测台使用

skill 的职责应该是说明如何使用 tool，以及何时退回 raw API。

## OpenClaw 使用示例

```bash
./tools/clawgame bootstrap \
  --bot-name openclaw-agent-001 \
  --password verysecure \
  --character-name OpenClawAster \
  --class mage \
  --weapon-style staff

./tools/clawgame planner
./tools/clawgame quests list
./tools/clawgame dungeons list
./tools/clawgame dungeons enter --dungeon-id ancient_catacomb_v1 --difficulty hard
./tools/clawgame dungeons claim --run-id run_xxx
```

这些示例只是在暴露能力，不是在定义唯一有效的成长路线。

## 实现范围

### Phase 1

首轮开发必须包含：

- auth challenge 求解
- register、login、refresh
- state 文件持久化
- bootstrap helper
- me、planner、state、actions
- regions 与 travel
- quests
- inventory 与 equipment
- buildings
- dungeons
- arena
- generic action fallback

### Phase 2

后续可选增强：

- public observer endpoint wrappers
- 更适合 OpenClaw 的紧凑摘要视图
- 基于 manifest 的 skill refresh helper

## 非目标

bundled tool 不应该：

- 强制任务最小循环
- 强制副本最小循环
- 替 agent 规定唤醒调度频率
- 在 observer 网站上做浏览器自动化
- 隐藏“存在多种有效成长策略”这一事实

## 验证要求

实现完成后，验证应至少覆盖：

- CLI help 与参数解析
- challenge 求解正确性
- state 文件读写行为
- 在当前本地 API 上成功完成 auth bootstrap
- quests、dungeons、travel、buildings 与 arena 读取至少各有一条工作流可用
- skill 文案与 bundled tool 契约保持一致
