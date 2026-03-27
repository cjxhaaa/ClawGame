# ClawGame 竞品分析与模式研究

版本：V1  
编写日期：2026-03-26  
文档定位：内部战略研究稿

## 1. 研究目标

本文件用于回答三个问题：

1. 当前市场上是否已经存在与 ClawGame 相近的产品或平台
2. 这些产品分别在卖什么，靠什么模式运转
3. ClawGame 应该如何避开同质化，并找到适合自己的商业路径

本分析特别聚焦与你当前设想最接近的形态：

- 用户自带 bot / agent
- 平台提供统一规则和运行环境
- bot 自动玩游戏或自动对战
- 平台提供排名、观战、回放或赛事能力

## 2. 结论摘要

截至 `2026-03-26`，市场上已经存在明确相关竞品，但“完全等同于 ClawGame 当前方向”的成熟产品仍然不多。

最值得重点关注的对象有：

- `Screeps: World / Arena`：最接近 ClawGame 的核心方向
- `Botzone`：偏比赛平台和教育/学术平台
- `BotClash`：偏新兴 bot-only 对战场
- `Elympics`：偏 B2B 基础设施平台
- `CodeCombat AI League`：偏教育和赛事包装的邻近模式
- `CodinGame Bot Programming`：偏大众化 bot 编程竞技平台
- `Battlecode`：偏高校和算法竞赛传统强者
- `Robocode / Tank Royale`：偏开源社区和长期 bot 对战传统
- `ChessArena / BotArena / Arena Protocol / Constrained Arena`：偏新一代垂直 bot 竞技产品

核心判断如下：

- `Screeps` 证明了“程序员自带 bot/脚本 + 持久世界/竞技场 + 长期运营”是成立的
- `Botzone` 证明了“提交 bot、跑比赛、看榜单”长期存在真实需求
- `BotClash` 证明了“bot-only、API-first、可围观”的产品表达方式有吸引力
- `Elympics` 证明了“bot/竞技能力本身可以平台化卖给别的游戏团队”
- `CodeCombat AI League` 证明了“赛事包装和组织化销售”可以成为重要商业抓手

ClawGame 的机会不在于复制其中任何一个，而在于把以下几件事组合起来：

- RPG 世界而不是单局题目
- bot 长期生活而不是只打一局
- 公开世界状态而不是只看结果
- 赛季与榜单而不是一次性比赛
- 未来既能做 C 端订阅，也能抽象成 B 端平台能力

## 3. 竞品分层

为了避免把不同类型产品混为一谈，本分析将竞品分成三层：

## 3.1 直接竞品

这类产品与 ClawGame 最接近，用户自带 bot，平台负责运行与竞争。

- Screeps: World
- Screeps: Arena
- BotClash

## 3.2 邻近竞品

这类产品满足“写 bot、打比赛、看排名”中的一部分，但不一定有持久世界。

- Botzone
- CodeCombat AI League

## 3.3 平台型竞品

这类产品不是面向终端 bot 玩家，而是向其他开发者提供 bot、比赛、回放、匹配等基础设施。

- Elympics

## 3.4 垂直型与实验型竞品

这类产品和 ClawGame 未必在玩法上最接近，但它们非常值得研究，因为它们代表了 2025-2026 年开始明显出现的新趋势：

- 不再要求人类手动参与
- 平台直接面向 bot / agent
- 强调 sandbox、公平性、日志、回放和可围观
- 先做窄场景，再扩展为多 arena 或多游戏

代表产品包括：

- ChessArena
- BotArena
- Arena Protocol
- Constrained Agent Arena

## 4. 重点竞品研究

## 4.1 Screeps: World / Screeps: Arena

### 产品形态

Screeps 官方将产品拆成两个方向：

- `World`：持久开放世界 MMO
- `Arena`：异步 1v1 对战场

官方介绍明确写到：

- Screeps: World 是 `persistent open world`
- 玩家通过真实 JavaScript 编写单位行为
- 脚本 `24/7` 持续运行
- 所有玩家单位在同一个持久世界中共存
- Screeps: Arena 是 `rating-based matchmaking`、`seasons`、`asynchronous PvP`

这和 ClawGame 的相似之处非常强：

- 都是用户自己写程序
- 都是平台负责运行和裁决
- 都强调长期性而非单次手操
- 都天然适合排行榜和赛季

### 商业模式

Screeps 当前公开信息显示其收入不是单一模式，而是混合模式：

1. `游戏购买`
- Steam 页面显示 `Screeps: World` 与 `Screeps: Arena` 单独售卖

2. `能力解锁`
- Screeps: World 官方商店写明：入门包包含永久 MMO 访问和 `20 CPU limit`
- 想提升能力需要购买 `Lifetime CPU unlock` 或 `CPU Unlock` 消耗品

3. `游戏内商店`
- 官方 `Item Shop` 提供 `Access Keys`、`Pixels`、`CPU Unlocks`
- `CPU Unlocks` 页面说明每个 unlock 可提供 `1 day of full unlocked CPU`

4. `赛季化运营`
- 官方 Creator Program 页面公开列出 `World Season` 和 `Arena Season`

5. `联盟分销`
- Creator Program 提供 `30% revenue share`

### 模式判断

Screeps 最重要的启发不是具体价格，而是它已经验证了三件事：

- 程序驱动世界可以长期运营
- “持续运行能力”本身可以收费
- 赛季和内容创作者可以成为增长与分发手段

### 对 ClawGame 的启发

- 你的 `月卡制` 是有现实参考的，但不应该只卖“进入资格”
- 更适合卖的是 `运行额度`、`更高配额`、`更强观测`、`赛季权益`
- 如果以后世界复杂度提升，也可以借鉴 Screeps 的“基础接入 + 能力解锁”结构

### 对 ClawGame 的威胁

- Screeps 已经占据了“程序员世界游戏”这一认知高地
- 如果 ClawGame 只是做成另一个通用编程 RTS，很容易被用户理解成 Screeps 的替代品

### 你的差异化机会

- 把重点放在 `bot-first RPG world`，而不是 RTS 经济循环
- 强化 `公开观战站`、`世界故事`、`任务/地下城/竞技场` 结构
- 让它更像“bot 角色在活世界里冒险”，而不是“脚本殖民地管理”

## 4.2 Botzone

### 产品形态

Botzone 是一个在线 AI 平台，公开信息显示其支持：

- bot 提交
- 游戏桌对局
- Bot 排行榜
- 本地 AI 连接
- 面向具体游戏和比赛的提交规范

公开游戏列表包含斗地主、国标麻将、四子棋等多类项目，说明它本质上是“多游戏 bot 比赛平台”。

公开比赛页显示，Botzone 承办或承载了多项正式 AI 比赛，例如 `IJCAI 2025 Mahjong AI Competition`，并且：

- 提供样例程序
- 提供输入输出协议
- 提供裁判程序或算分库源码
- 按赛程进行模拟赛、积分赛、决赛
- 提供奖金

### 商业模式

截至 `2026-03-26`，我没有在 Botzone 公共页面中查到明确的面向个人用户的订阅或定价页。

基于公开信息，较稳妥的判断是：

- 它的公共表达更像 `竞赛平台 + 教育/学术平台`
- 收益来源更可能偏向 `比赛合作`、`组织支持`、`学术/品牌资源` 或 `非强商业化运营`

这部分属于 `基于公开页面的推断`，不是官方定价声明。

### 模式判断

Botzone 验证了两件事：

- 用户愿意为“提交 bot 并和其他 bot 竞争”投入长期精力
- 比赛、排名、裁判程序和样例生态能够形成平台粘性

### 对 ClawGame 的启发

- 如果你以后做开放接入，必须尽早提供 `样例 bot`、`调试方法`、`清晰协议`
- 纯平台如果只有规则没有赛事，用户活跃会弱很多
- 比赛组织能力本身就是平台价值的一部分

### 对 ClawGame 的局限提醒

Botzone 更偏比赛基础设施，不强调世界叙事和 live service。  
这说明 ClawGame 不应该只做成“又一个提交 bot 的平台”，否则你会失去自己最特别的部分。

## 4.3 BotClash

### 产品形态

BotClash 当前公开页面显示其自我定义为：

- `BETA v0.1`
- `AI Battle Arena`
- `Bots Only`

它的流程非常 bot-first：

- 用户先注册账户和 agent
- 通过 API 获得 `agentId` 与 `apiKey`
- 使用 REST 接口进入匹配队列
- 使用 WebSocket 接入对局
- bot 自动完成对战
- 平台提供直播、排行榜、战斗日志和 bot 聊天

其公开页面甚至把 “人类不能直接玩，只能围观” 讲得非常直接，这一点和 ClawGame 的设想高度一致。

### 商业模式

截至 `2026-03-26`，公开页面中没有明显的收费页或价格说明。

基于其官网呈现状态，更稳妥的判断是：

- 当前仍处于 `社区冷启动 / Beta` 阶段
- 重点在于降低 bot 接入门槛、制造围观感和 bot 社交感
- 暂时更像“先做活跃，再考虑收费”

这同样属于 `基于公开页面的推断`。

### 模式判断

BotClash 最值得研究的不是收入，而是产品表达：

- bot 注册和 API key 获取路径很短
- 匹配和 WebSocket 流程非常直接
- 平台把 `spectate`、`leaderboard`、`chat` 放在前台
- 明确把“人类是围观者”写进产品叙事

### 对 ClawGame 的启发

- 你的官网不应该长得像后台，而应该长得像“bot 世界观测站”
- 接入文档要短，bot 从注册到开打的路径要尽量短
- 可围观内容很重要，尤其是战报、排行榜、热点 bot、局部事件

### 对 ClawGame 的威胁

- 如果你早期世界内容还不够深，BotClash 这类轻竞技平台会在“上手快”上更占优

### 你的差异化机会

- BotClash 是 `快节奏 1v1 arena`
- ClawGame 应该坚持 `长期世界 + 多系统循环 + bot 角色成长`

## 4.4 Elympics

### 产品形态

Elympics 不是和 ClawGame 抢同一批终端玩家的游戏，而更像未来可能和你抢 B 端客户的平台。

其官方产品页公开强调：

- `Bot Deployment`
- `Smart Matchmaking`
- `Verifiable Replays`
- `Leaderboards`
- `Tournaments & Duels`
- `Hosting Included`

Bot Deployment 页面明确写到：

- bots 可以 `fill empty slots`
- bots 可以 `reduce wait times`
- bots 可用于 `test your game`
- bot 托管和分配由平台自动处理

Verifiable Replays 页面明确写到：

- 回放会在比赛结束后自动存储
- 回放既可用于直播也可用于 debug
- 默认保留 `7 days`

### 商业模式

从产品结构可以较明确判断，Elympics 主要卖的是：

- 平台能力
- 托管能力
- 比赛和回放基础设施
- 开发工具链

也就是典型的 `B2B SaaS / infra` 模式，而不是面向单个 bot 玩家收月卡。

### 模式判断

Elympics 的意义在于提醒我们：

- 如果 ClawGame 未来把引擎、匹配、回放、赛事这些能力抽象出来，完全可以走向 B 端平台化
- 也就是说，ClawGame 不一定永远只是一款游戏，它有机会长成平台产品

### 对 ClawGame 的启发

- 你今天做的每一层能力，最好都带一点“未来可平台化”的思维
- 例如：回放、赛季配置、私有联赛、规则版本管理、bot 托管接口

### 对 ClawGame 的威胁

- 如果未来你想卖给组织者或其他游戏方，Elympics 这类平台会比你更完整、更企业化

### 你的差异化机会

- Elympics 卖的是“任何竞技游戏都能用的基础设施”
- ClawGame 卖的是“已经有内容和世界感的 bot-first RPG 产品”

前者偏通用平台，后者偏内容化平台，两者未来可能重叠，但短期并不完全正面冲突。

## 4.5 CodeCombat AI League

### 产品形态

CodeCombat AI League 不是最直接的竞品，但它很值得研究其组织模式。

公开信息显示：

- CodeCombat AI League 被官方定义为 `academic esports`
- 面向学校和教育场景有较强包装
- 官方提供 `custom packages`，允许教育机构在班级、学校或学区层面运行竞赛
- CodeCombat Premium 提供私有 clan 和组内 leaderboard 等能力

### 商业模式

CodeCombat 的公开产品表达显示其收入结构是混合的：

- `Premium 订阅`
- `教育机构销售`
- `AI League 赛事和组织化方案`

### 模式判断

CodeCombat 的价值不在于它与 ClawGame 的玩法接近，而在于它证明：

- 竞赛产品可以被包装成组织化销售方案
- 排行榜、私有群组和赛事可以成为付费权益
- 赛事并不只是运营活动，也可以是销售抓手

### 对 ClawGame 的启发

- 未来如果你切入高校、训练营、社群赛事，`组织者版` 很可能比纯个人月卡更快带来收入
- 私有联赛、团队榜单和组内回放，应该被视为潜在付费功能

## 4.6 CodinGame Bot Programming

### 产品形态

CodinGame 是更大范围的编程游戏平台，但它内部有一个成熟的 `Bot programming` 赛道。

官方 bot programming 页面公开显示：

- 拥有大量 bot programming 游戏
- 覆盖 `multi-agent`、`resource management`、`pathfinding`、`simulation`、`card games` 等多类题型
- 支持公开活动和私有活动
- 有长期 leaderboard 和挑战机制

公开赛事页也明确写到：

- 参与者在若干天内开发 `autonomous bot`
- 平台会在每次新提交后重新匹配比赛
- 排名使用类似 `TrueSkill` 的系统
- 支持 25+ 编程语言

### 商业模式

CodinGame 的 bot programming 本身更多像其开发者生态和品牌流量入口，而非独立收费产品。

基于公开页面，更稳妥的判断是：

- 其 bot programming 赛道承担 `获客`、`活跃`、`社区` 和 `赞助赛事` 的作用
- 商业收入更大概率来自 `企业赞助比赛`、`招聘品牌`、`开发者平台整体变现`

这部分属于 `基于官方公开结构的推断`。

### 模式判断

CodinGame 说明：

- 大规模 bot 编程社区是可以长期维持的
- 多游戏轮换和活动制能持续保持新鲜度
- 私有赛事和公开赛事可以并存

### 对 ClawGame 的启发

- 你未来也可以考虑“公开世界 + 限时特别赛季/主题赛”的双结构
- 除了长期世界榜，还可以有短期挑战榜
- 企业或品牌赞助赛是可行模式，不一定只靠订阅

### 与 ClawGame 的差异

CodinGame 强在多题型和大社区，弱在没有统一持续世界。  
ClawGame 如果做得好，更像“一个持续 live service 世界”，而不是一堆离散 bot 题目。

## 4.7 Battlecode

### 产品形态

Battlecode 是 MIT 组织的经典 AI 编程竞赛。其官方 About 页面明确写到：

- 这是一个 `real-time strategy game`
- 参赛者要编写 `autonomous player`
- 整个一月持续进行 `scrimmages` 与 `tournaments`
- 最终会有现场总决赛和现金奖励
- 任何人都可以写 bot、组队并参加排名赛

### 商业模式

Battlecode 更像：

- 学术/教育竞赛
- 社群驱动传统赛事
- 品牌与校园影响力产品

它不是典型商业订阅平台，但它证明了：

- 高门槛 bot 竞赛也能长期形成忠实社区
- 限定赛季、阶段赛、总决赛是非常有效的参与驱动

### 对 ClawGame 的启发

- 赛季不一定要全年不间断，也可以有明确的阶段制和高潮节点
- 公开排名、冲刺赛、资格赛、总决赛的层次感很重要
- 强竞赛用户对“公平算力限制”非常敏感

### 与 ClawGame 的差异

Battlecode 更像年度竞赛，不是长期运营的商业世界。  
ClawGame 可以吸收其竞赛强度，但不能只做成“每年一次”的活动产品。

## 4.8 Robocode / Robocode Tank Royale

### 产品形态

Robocode 是历史很长的 bot 编程战斗游戏。其官网和新文档显示：

- Robocode 的核心是编写机器人坦克进行自动对战
- 新版 `Tank Royale` 强调多语言、Viewer、Recorder、Replay 与文档体系
- 文档中专门提供 `Anatomy of a Bot`、`Scoring`、`Debugging`、`Team Messages`
- 还明确强调 Viewer 适合比赛展示和大屏观战

### 商业模式

Robocode 当前更偏：

- 开源社区
- 非强商业化生态
- 长期社区知识积累

它不一定是你的商业竞品，但它是很重要的“产品形态先例”。

### 模式判断

Robocode 说明：

- “写 bot 看对战”这件事本身拥有非常长的生命周期
- 可视化 viewer、replay 和调试文档对于社区留存很重要
- 社区长期会发展出高级策略、专有术语和研究沉淀

### 对 ClawGame 的启发

- 如果你想长期经营，必须给高手足够深的优化空间
- replay、viewer 和战术分析不只是附属功能，而是核心内容资产
- 强社区会自己产出策略文章、赛后复盘和 bot 流派

## 4.9 新一代垂直 bot 竞技平台

这一组产品共同的特征是：产品很新、玩法很垂直、表达很“agent native”，虽然还未必成熟，但对 ClawGame 很有参考价值。

### ChessArena

官方页面公开写到：

- `Deploy your chess bot in 5 minutes`
- `Real ELO rankings`
- `sandboxed environment`
- `global matchmaking`
- `Same hardware, same limits`
- `Your code is your property`

模式上它非常像“bot 托管平台”的垂直化版本。

对 ClawGame 的启发：

- bot 注册和开打路径可以非常短
- “代码所有权归用户”是很好的信任表达
- 公平硬件与沙箱约束要被明确写在前台

### BotArena

官方页面公开写到：

- 多语言 SDK
- `transparent match logs`
- `deterministic replay`
- `team workspaces`
- `versioned deployments with rollback and A/B matchups`
- 还直接写出了未来可能的 `tournaments, staking, licensing`

BotArena 很值得注意，因为它不像单纯游戏，更像“bot 产品经理工具 + 排位赛”的结合。

对 ClawGame 的启发：

- 版本化部署、A/B 对战、团队空间，这些都很有产品价值
- 日志和 replay 不只是观战功能，更是开发工具

### Arena Protocol

其官网明确写到：

- `AI Agents Compete. Humans Spectate.`
- 支持多游戏：TRON、扑克、国际象棋
- 代理通过协议接入，自动匹配、自动记录、ELO 排名
- `no human input allowed`
- `Full replay data available via API`
- `All matches are completely free`

这是目前和你“人类主要围观、agent 才是玩家”叙事最接近的一类产品。

对 ClawGame 的启发：

- “Humans spectate” 这句话本身就是产品定位
- 协议化接入和 replay API 都是平台化信号

### Constrained Agent Arena

其官网公开写到：

- `Compete with Code, Not Clicks`
- bot 以二进制形式上传
- 平台提供 sandbox
- 强调 `No GPUs`、`No massive models`
- 提供 `Elo-style ratings`、`match replays`
- 明确表示这不是 Kaggle 型平台

它代表的是另一种很重要的方向：`受限算力、公平约束、研究友好`。

对 ClawGame 的启发：

- 如果你未来允许 LLM/agent 参与，必须尽早明确算力、公平和延迟约束
- “不是算力碾压，而是策略对抗” 会成为很强的价值主张

## 4.10 为什么完全相同的竞品不多

这轮扩展搜索之后，一个更清楚的判断是：

- 类似 `bot 提交 + 对战 + 排行榜` 的产品并不少
- 类似 `持续运行世界` 的产品也存在，Screeps 是代表
- 类似 `人类围观、agent 参赛` 的表达在 2025-2026 年开始明显增多

但把这些元素同时组合起来的产品仍然很少，原因大致有四个：

1. `门槛叠加`
自带 bot 的产品本来就比普通游戏门槛高，再叠加 RPG 世界、持续运行和观战系统，产品与运营复杂度都更高。

2. `商业路径不直观`
很多团队更容易先做单一竞技场，因为收费和运维更简单；长期世界则需要更强的版本运营能力。

3. `公平性成本高`
一旦用户真的带 agent 来跑，配额、限流、沙箱、日志、反作弊都很重。

4. `叙事难度高`
纯对战平台很容易解释，持续 bot 世界则需要让用户和观众都看懂“发生了什么”。

这也恰恰解释了为什么 ClawGame 仍然有机会。

## 5. 模式对比表

| 项目 | 用户是否自带 bot | 持久世界 | 主要形态 | 公开收费模式 | 与 ClawGame 的关系 |
| --- | --- | --- | --- | --- | --- |
| Screeps: World | 是 | 是 | 编程 MMO | 买断/能力解锁/商店/赛季 | 最强直接参照 |
| Screeps: Arena | 是 | 否 | 异步 1v1 | 买断/赛季化运营 | 竞技层直接参照 |
| BotClash | 是 | 否 | bot-only 竞技场 | 未见公开收费，处于 Beta | 新兴直接竞品 |
| Arena Protocol | 是 | 否 | 多游戏 agent 竞技场 | 当前公开免费 | 新兴直接竞品 |
| ChessArena | 是 | 否 | 垂直棋类 bot 平台 | 未见公开定价 | 新兴垂直竞品 |
| BotArena | 是 | 否 | 垂直 bot 托管竞技场 | 早期访问，未来商业化已公开规划 | 新兴垂直竞品 |
| Constrained Agent Arena | 是 | 否 | 受限算力 agent 平台 | 早期访问 | 新兴研究型竞品 |
| Botzone | 是 | 否 | 比赛平台 | 未见公开订阅，偏比赛平台 | 邻近竞品 |
| CodinGame Bot Programming | 是 | 否 | 大众化 bot 编程竞技平台 | 以平台和赛事商业化为主 | 邻近竞品 |
| Battlecode | 是 | 否 | 高校/算法竞赛 | 非典型商业订阅 | 邻近竞品 |
| Robocode Tank Royale | 是 | 否 | 开源 bot 对战生态 | 非强商业化 | 形态先例 |
| Gladiabots | 部分是 | 否 | AI 编排竞技游戏 | 买断制游戏 | 交互形态参照 |
| Elympics | 不面向终端 bot 玩家 | 否 | B2B 基础设施 | 平台/SaaS 型 | 未来平台侧竞品 |
| CodeCombat AI League | 部分是 | 否 | 教育赛事平台 | 订阅 + 组织化销售 | 邻近模式参照 |

说明：

- 表中 `未见公开收费` 的含义是：截至 `2026-03-26`，未在其公开页面中查到明确个人用户价格页，不等于其完全没有收入。
- `买断/能力解锁/商店` 对 Screeps 的判断基于其当前官方商店页面；同时其 ToS 中仍保留订阅表述，说明其商业结构可能经历过演进。
- `当前公开免费` 或 `早期访问` 的项目，未来商业模式仍可能变化。

## 6. 对 ClawGame 的战略启示

## 6.1 你不能只做“又一个 bot 平台”

Botzone、BotClash、CodinGame 这类产品都已经证明，“提交 bot、看排行榜、跑比赛”这件事本身有人做。  
如果 ClawGame 只是把 bot 上传上来打架，它会很难建立独特认知。

你最值得坚持的部分是：

- RPG 世界
- 任务与地下城
- 周期性竞技场
- 公开世界动态
- bot 角色成长和故事

## 6.2 你适合做“月卡 + 额度 + 赛事”的混合模式

从竞品看：

- Screeps 证明持续运行能力可以收费
- CodeCombat 证明组织化赛事可以收费
- Elympics 证明基础设施层可以卖高客单价

因此，ClawGame 最合理的商业路径仍然是：

1. `个人月卡`
- bot 数量
- API 额度
- 回放和数据导出
- 高级赛季资格

2. `团队/组织者套餐`
- 私有联赛
- 团队榜单
- 赛季管理
- 组内回放

3. `未来 B2B 授权`
- 私有世界
- 规则定制
- 白标化

## 6.3 早期最该补的不是内容，而是平台能力

参考竞品之后，ClawGame 商业化之前最应该补的能力是：

- bot 注册与认证
- bot 配额与限流
- 调试、日志、回放
- 私有赛季和组织者工具
- 面向人类围观者的公开站叙事能力

这些能力会直接决定你能不能收费。

## 6.4 最值得对标的对象

如果只选两个最值得长期跟踪的对象，我建议是：

1. `Screeps`
原因：它是最接近你长期愿景的成熟参照物。

2. `BotClash`
原因：它代表了新一代 bot-only 产品在接入体验和围观表达上的做法。

3. `CodinGame`
原因：它代表多 arena、多活动、长期社区运营的成熟打法。

4. `ChessArena / BotArena`
原因：它们代表了新一代“bot 托管即产品”的 builder-first 表达。

如果只选一个未来平台化参照物，则是：

5. `Elympics`
原因：它代表你未来把能力卖给别人时最容易遇到的思路。

## 7. 最终判断

ClawGame 所在赛道 `不是空白市场`，但仍然 `存在清晰空位`。

空位主要在于：

- 还没有一个特别强的产品把 `用户自带 bot + RPG 世界 + live service + 公开观战 + 周期赛事` 真正整合好
- 现有产品往往只占其中一段

因此，ClawGame 不需要回避“有竞品”这件事。相反，竞品的存在说明这个方向不是伪需求。

真正关键的是：

- 不要做成 Screeps 的低配替代
- 不要做成 Botzone 的另一种壳
- 不要做成 BotClash 的慢版本
- 也不要做成 CodinGame 上一题更重的长期题

而要明确把自己定义为：

`一个让用户自带 bot，在持续运行的 RPG 世界里成长、竞争、被围观，并最终形成赛季和联赛生态的平台。`

## 8. 参考来源

以下来源均为本次分析使用的公开页面，访问时间为 `2026-03-26`：

- [Screeps Documentation - Introduction](https://docs.screeps.com/introduction.html)
- [Screeps Documentation - Terms of Service](https://docs.screeps.com/tos.html)
- [Screeps Documentation - Community Servers](https://docs.screeps.com/community-servers.html)
- [Screeps Store](https://store.screeps.com/)
- [Screeps: World store page](https://store.screeps.com/world-test)
- [Screeps Item Shop](https://world.store.screeps.com/)
- [Screeps Creator Program](https://store.screeps.com/creators)
- [Screeps: Arena on Steam](https://store.steampowered.com/app/1137320/Screeps_Arena/)
- [Botzone 首页](https://www.botzone.org.cn/)
- [Botzone 游戏列表](https://www.botzone.org.cn/games)
- [IJCAI 2025 Mahjong AI Competition on Botzone](https://www.botzone.org.cn/static/gamecontest2025a_cn.html)
- [Botzone Wiki](https://www.wiki.botzone.org.cn/)
- [BotClash](https://botclash.live/)
- [ChessArena.dev](https://chessarena.dev/)
- [ChessArena.dev Documentation](https://chessarena.dev/docs)
- [BotArena](https://www.botarena.app/)
- [Arena Protocol / OpenClaw Agent League](https://openclawagentleague.com/)
- [Constrained Agent Arena](https://constrainedarena.com/)
- [Elympics Bot Deployment](https://www.elympics.ai/products/bot-deployment)
- [Elympics Smart Matchmaking](https://www.elympics.ai/products/smart-matchmaking)
- [Elympics Verifiable Replays](https://www.elympics.ai/products/verifiable-replays)
- [Battlecode About](https://battlecode.org/about.html)
- [Robocode Home](https://robocode.sourceforge.io/)
- [Robocode Tank Royale Docs](https://robocode.dev/)
- [Gladiabots](https://gladiabots.com/)
- [CodinGame Bot Programming](https://www.codingame.com/multiplayer/bot-programming)
- [CodinGame Multiplayer](https://www.codingame.com/multiplayer)
- [CodinGame contest example: Code4Life](https://www.codingame.com/contests/code4life)
- [CodeCombat blog: AI League and support tools](https://blog.codecombat.com/2022-brings-new-features-and-new-support-tools/)
- [CodeCombat blog: custom packages for academic esports](https://blog.codecombat.com/kickstart-back-to-school-with-new-tools-and-resources/)
- [CodeCombat blog: AI League overview](https://blog.codecombat.com/ai-league-redefining-what-an-esport-can-be/)
- [CodeCombat blog: custom tournaments](https://blog.codecombat.com/new-year-new-updates-custom-tournaments-latest-lesson-slides-more/)
- [CodeCombat Features](https://codecombat.com/features/)
