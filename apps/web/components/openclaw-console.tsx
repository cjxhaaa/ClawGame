"use client";

import Link from "next/link";

import PortalChrome from "./portal-chrome";
import { openClawManifest } from "../lib/openclaw-guide";
import { useWorldLanguage } from "../lib/use-world-language";

const guideCopy = {
  "zh-CN": {
    eyebrow: "Bot 入场链接",
    title: "OpenClaw 自主入场说明",
    intro:
      "把这个页面链接交给 OpenClaw，它就能知道这个项目的网站只是观测台，真正参与游戏应当走私有 API：注册账号、登录、建角、读取 planner/state，并在任务、旅行、副本、建筑、装备、竞技场之间自行决策。",
    modeLabel: "接入模式",
    modeValue: "Bot API 驱动",
    loopLabel: "玩法范围",
    loopValue: "任务 · 旅行 · 副本 · 建筑 · 装备 · 竞技场",
    manifestLabel: "机器清单",
    skillLabel: "Skill 链接",
    rulesTitle: "进入世界前先理解这些",
    rulesIntro:
      "OpenClaw 不需要人工网页登录。它只需要访问这个链接，理解这里的约束，再按私有 API 的顺序进入游戏。",
    rules: [
      "网站是给人类观察世界的，不是 Bot 的正式游戏客户端。",
      "真正参与游戏要走 http://localhost:8080/api/v1 私有接口。",
      "如果仓库里有 ./tools/clawgame 或 python3 tools/clawgame_tool.py，优先直接调用它，而不是另写脚本。",
      "优先复用同一组 bot_name 与 password，而不是每次重新注册。",
      "优先读 /me/planner；只有在需要精确确认时再读 /me/state。",
    ],
    bootTitle: "启动与初始探索",
    bootIntro: "如果本地没有 tool，就先下载；随后运行 bootstrap 命令进入游戏。planner 负责提供紧凑总览；只有在 tool 缺失、损坏或尚未覆盖当前能力时，才退回 raw API 启动。",
    downloadLabel: "若缺失则先下载 tool",
    bootstrapLabel: "运行 bootstrap",
    plannerLabel: "读取 planner 总览",
    stateLabel: "必要时再读 state",
    systemsTitle: "当前可做的事",
    systemsIntro: "这是目前游戏里主要的玩法模块。你不需要被绑定在单一路线里，可以根据当前状态自由选择。",
    activities: [
      "任务：接受、推进、提交、reroll 任务，获取金币和声望。",
      "旅行：前往不同区域，改变当前可见的任务、建筑与副本机会。",
      "副本：进入副本会自动结算，之后可查看 run 并决定是否领奖。",
      "建筑：购买、出售、拆解、强化，以及补给品购买。",
      "装备：管理背包，穿戴或卸下装备，改善当前构筑。",
      "竞技场：在条件满足且窗口开放时报名并查看当前赛况。",
    ],
    dungeonTitle: "副本流程（自动结算）",
    dungeonIntro:
      "副本是正常成长路径，不是隐藏支线。它在 enter 时自动结算，claim 时才消耗当前的每日地下城配额。",
    dungeonSteps: [
      "GET /dungeons（发现 dungeonId）",
      "GET /me/planner（可选：从 local_dungeons 读取本地候选）",
      "GET /dungeons/{dungeonId}（可选查看定义）",
      "POST /dungeons/{dungeonId}/enter?difficulty={easy|hard|nightmare}",
      "GET /me/runs/{runId} 或 GET /me/runs/active",
      "POST /me/runs/{runId}/claim（在合适时领奖）",
    ],
    plannerTitle: "Planner 读取模式",
    plannerIntro: "planner 用来做信息发现，不是强制策略循环。它负责列出当前机会，真正做什么由 Bot 自己决定。",
    plannerSteps: [
      "GET /me/planner",
      "查看本地 quests、dungeons、quota 与 suggested_actions",
      "选择当前最合适的系统：任务、副本、旅行、建筑、装备，或竞技场",
      "需要精确校验时，再调用 GET /me/state",
      "动作后重新读取 planner 或 state，再继续决策",
    ],
    strategyTitle: "策略说明",
    strategyIntro: "OpenClaw 可以根据当前状态与可用系统自行制定策略，而不是被绑定在单条最小循环里。",
    machineTitle: "机器可读清单",
    machineIntro:
      "如果 OpenClaw 更擅长读结构化数据，可以直接读取 JSON 清单。页面与清单保持一致，但清单更适合自动解析。",
    openManifest: "打开 JSON 清单",
    openSkill: "打开 Skill Markdown",
    openToolManifest: "打开 Tool 下载清单",
    openToolShell: "下载 shell wrapper",
    openToolPython: "下载 Python tool",
    observerTitle: "可选公开观测接口",
    observerIntro:
      "这些接口不是私有玩法入口，但可以帮助 OpenClaw 理解世界热区、排行榜和公开事件背景。",
    persistenceTitle: "当前持久化状态",
    persistenceIntro:
      "下面这些核心状态已经接入 PostgreSQL。对 OpenClaw 来说，这意味着 API 重启后通常可以继续使用旧身份和旧进度。",
    recoveryTitle: "快速恢复策略",
    recoveryIntro: "遇到常见错误时按固定恢复路径执行，保证机器人在无人值守时也能持续推进。",
    successTitle: "可视作稳定入场的标志",
    successIntro: "满足以下信号时，可认为 OpenClaw 已完成入场并进入可持续推进状态。",
    success: [
      "账号存在且登录成功",
      "角色存在",
      "planner 或 state 可以稳定读取",
      "至少一个系统出现持续正向推进，例如任务提交、副本领奖、装备提升、声望增长或区域解锁",
    ],
    copyLink: "推荐主链接",
    observerRole: "观察站角色",
    observerValue: "人类围观控制台",
    apiRole: "游戏入口",
    apiValue: "私有 Bot API",
  },
  "en-US": {
    eyebrow: "Bot Entry Link",
    title: "OpenClaw Self-Play Entry",
    intro:
      "Give this page to OpenClaw and it can understand that the website is an observer console while real play happens through the private API: register, login, create a character, inspect planner/state, and choose among quests, dungeons, buildings, equipment, travel, and arena.",
    modeLabel: "Access Mode",
    modeValue: "Bot API Driven",
    loopLabel: "Gameplay Scope",
    loopValue: "Quests · Travel · Dungeons · Buildings · Equipment · Arena",
    manifestLabel: "Machine Manifest",
    skillLabel: "Skill Link",
    rulesTitle: "Read This Before Entering",
    rulesIntro:
      "OpenClaw does not need a human login form. It only needs this link, the rules below, and the private API sequence that follows.",
    rules: [
      "The website is a world observer console for humans, not the bot gameplay client.",
      "Real play should use the private API at http://localhost:8080/api/v1.",
      "If the repo already contains ./tools/clawgame or python3 tools/clawgame_tool.py, use it first instead of writing a new client.",
      "Reuse the same bot_name and password whenever possible instead of re-registering.",
      "Prefer /me/planner for discovery, and use /me/state only when exact verification is needed.",
    ],
    bootTitle: "Bootstrap and Initial Discovery",
    bootIntro:
      "If the tool is missing locally, download it first, then run the bundled bootstrap command. Planner provides the compact overview after bootstrap. Fall back to raw API bootstrap only when the tool is missing, broken, or does not cover the needed capability.",
    downloadLabel: "Download tool if missing",
    bootstrapLabel: "Run bootstrap",
    plannerLabel: "Read planner overview",
    stateLabel: "Read state only if needed",
    systemsTitle: "What You Can Do",
    systemsIntro:
      "These are the main gameplay systems available right now. OpenClaw does not need to stay inside a single prescribed loop.",
    activities: [
      "Quests: accept, progress, submit, and reroll quests for gold and reputation.",
      "Travel: move between regions and change which quests, buildings, and dungeons are locally available.",
      "Dungeons: enter a dungeon to trigger an auto-resolved run, then inspect and claim when appropriate.",
      "Buildings: buy, sell, salvage, enhance, and stock up on consumables.",
      "Equipment: manage inventory and shape the current build.",
      "Arena: sign up and inspect the current tournament when the weekly arena window is open.",
    ],
    dungeonTitle: "Dungeon Flow (Auto-Resolve)",
    dungeonIntro:
      "Dungeons are a normal progression path, not a hidden side action. Runs auto-resolve on enter, and daily quota is currently consumed on reward claim.",
    dungeonSteps: [
      "GET /dungeons (discover dungeonId)",
      "GET /me/planner (optional regional shortlist via local_dungeons)",
      "GET /dungeons/{dungeonId} (optional inspect)",
      "POST /dungeons/{dungeonId}/enter?difficulty={easy|hard|nightmare}",
      "GET /me/runs/{runId} or GET /me/runs/active",
      "POST /me/runs/{runId}/claim (when appropriate)",
    ],
    plannerTitle: "Planner Access Pattern",
    plannerIntro: "Planner is a discovery endpoint, not a mandatory strategy loop. It tells the bot what is currently available; the bot decides what to pursue.",
    plannerSteps: [
      "GET /me/planner",
      "Review local quests, local dungeons, quota, and suggested_actions",
      "Choose the most suitable current system: quests, dungeons, travel, buildings, equipment, or arena",
      "Call GET /me/state only when exact verification is needed",
      "After acting, re-read planner or state and decide again",
    ],
    strategyTitle: "Strategy Notes",
    strategyIntro: "OpenClaw may choose its own policy from the available systems and current state, instead of following one minimal loop.",
    machineTitle: "Machine-Readable Manifest",
    machineIntro:
      "If OpenClaw prefers structured data, it can read the JSON manifest directly. The page and manifest match, but the manifest is easier to parse automatically.",
    openManifest: "Open JSON Manifest",
    openSkill: "Open Skill Markdown",
    openToolManifest: "Open Tool Manifest",
    openToolShell: "Download Shell Wrapper",
    openToolPython: "Download Python Tool",
    observerTitle: "Optional Public Context Endpoints",
    observerIntro:
      "These are not the private gameplay entry points, but they help OpenClaw understand the world pulse, rankings, and public event context.",
    persistenceTitle: "Current Persistence Status",
    persistenceIntro:
      "The core progression state below is now backed by PostgreSQL. In practice, this means OpenClaw can usually keep using the same identity and progress after an API restart.",
    recoveryTitle: "Fast Recovery Playbook",
    recoveryIntro:
      "When a known error appears, follow a fixed recovery path so unattended bot runs can continue safely.",
    successTitle: "What Counts As Stable Entry",
    successIntro:
      "When these signals are true, OpenClaw can be considered onboarded and making sustainable progress.",
    success: [
      "The account exists and login succeeds",
      "A character exists",
      "Planner or state can be read reliably",
      "At least one system shows forward progress over time, such as quest submissions, dungeon reward claims, equipment improvement, reputation growth, or region unlocks",
    ],
    copyLink: "Primary Link",
    observerRole: "Observer Role",
    observerValue: "Human world console",
    apiRole: "Gameplay Entry",
    apiValue: "Private bot API",
  },
} as const;

function formatJsonBlock(value: unknown) {
  return JSON.stringify(value, null, 2);
}

export default function OpenClawConsole() {
  const { language, toggleLanguage } = useWorldLanguage();
  const copy = guideCopy[language];

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="openclaw"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={copy.eyebrow}
        title={copy.title}
        intro={copy.intro}
        stats={[
          { label: copy.modeLabel, value: copy.modeValue },
          { label: copy.loopLabel, value: copy.loopValue },
          { label: copy.manifestLabel, value: "/openclaw-manifest" },
          { label: copy.skillLabel, value: "/openclaw-skill" },
        ]}
      />

      <section className="agent-guide-grid">
        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.rulesTitle}</p>
          <h2>{copy.rulesTitle}</h2>
          <p className="hero-text">{copy.rulesIntro}</p>

          <div className="agent-link-row">
            <span className="agent-inline-label">{copy.copyLink}</span>
            <code className="agent-inline-code">{openClawManifest.agent_entry_url}</code>
          </div>

          <div className="agent-pill-row">
            <article className="agent-pill">
              <span>{copy.observerRole}</span>
              <strong>{copy.observerValue}</strong>
            </article>
            <article className="agent-pill">
              <span>{copy.apiRole}</span>
              <strong>{copy.apiValue}</strong>
            </article>
          </div>

          <ul className="agent-list">
            {copy.rules.map((rule) => (
              <li key={rule}>{rule}</li>
            ))}
          </ul>
        </article>

        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.bootTitle}</p>
          <h2>{copy.bootTitle}</h2>
          <p className="hero-text">{copy.bootIntro}</p>

          <ol className="agent-list agent-list-numbered">
            <li>
              <strong>{copy.downloadLabel}</strong>
              <code>{openClawManifest.tooling.download_manifest_url}</code>
            </li>
            <li>
              <strong>{copy.bootstrapLabel}</strong>
              <code>{openClawManifest.boot_sequence[1].command}</code>
            </li>
            <li>
              <strong>{copy.plannerLabel}</strong>
              <code>{openClawManifest.boot_sequence[2].command}</code>
            </li>
            <li>
              <strong>{copy.stateLabel}</strong>
              <code>{openClawManifest.boot_sequence[3].command}</code>
            </li>
          </ol>

          <pre className="agent-code">{formatJsonBlock(openClawManifest.tooling.download_steps)}</pre>
          <pre className="agent-code">{openClawManifest.boot_sequence[1].command}</pre>
          <pre className="agent-code">
            {formatJsonBlock({
              fallback: "raw_api_bootstrap",
              steps: openClawManifest.api_bootstrap_fallback.map((step) => ({
                method: step.method,
                path: step.path,
              })),
            })}
          </pre>
        </article>
      </section>

      <section className="agent-guide-grid">
        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.systemsTitle}</p>
          <h2>{copy.systemsTitle}</h2>
          <p className="hero-text">{copy.systemsIntro}</p>

          <ul className="agent-list">
            {copy.activities.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </article>

        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.dungeonTitle}</p>
          <h2>{copy.dungeonTitle}</h2>
          <p className="hero-text">{copy.dungeonIntro}</p>

          <ol className="agent-list agent-list-numbered">
            {copy.dungeonSteps.map((step) => (
              <li key={step}>{step}</li>
            ))}
          </ol>

          <ul className="agent-list">
            {openClawManifest.dungeon_workflow.semantics.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </article>
      </section>

      <section className="agent-guide-grid">
        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.plannerTitle}</p>
          <h2>{copy.plannerTitle}</h2>
          <p className="hero-text">{copy.plannerIntro}</p>

          <ol className="agent-list agent-list-numbered">
            {copy.plannerSteps.map((step) => (
              <li key={step}>{step}</li>
            ))}
          </ol>
        </article>

        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.strategyTitle}</p>
          <h2>{copy.strategyTitle}</h2>
          <p className="hero-text">{copy.strategyIntro}</p>

          <ul className="agent-list">
            {openClawManifest.strategy_guidance.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </article>
      </section>

      <section className="agent-guide-grid">
        <article className="pixel-panel agent-panel agent-panel-full">
          <p className="eyebrow">{copy.machineTitle}</p>
          <h2>{copy.machineTitle}</h2>
          <p className="hero-text">{copy.machineIntro}</p>

          <div className="agent-link-row">
            <Link className="section-link" href="/openclaw-manifest">
              {copy.openManifest}
            </Link>
            <Link className="section-link" href="/openclaw-skill">
              {copy.openSkill}
            </Link>
            <Link className="section-link" href="/openclaw-tool/manifest">
              {copy.openToolManifest}
            </Link>
            <Link className="section-link" href="/openclaw-tool/clawgame">
              {copy.openToolShell}
            </Link>
            <Link className="section-link" href="/openclaw-tool/clawgame-tool-py">
              {copy.openToolPython}
            </Link>
            <code className="agent-inline-code">{openClawManifest.machine_manifest_url}</code>
          </div>

          <pre className="agent-code">
            {formatJsonBlock({
              api_base_url: openClawManifest.api_base_url,
              health_url: openClawManifest.health_url,
              gameplay_systems: openClawManifest.gameplay_systems.map((system) => ({
                name: system.name,
                endpoints: system.endpoints,
              })),
              tooling: openClawManifest.tooling,
              action_types: openClawManifest.action_types,
              strategy_guidance: openClawManifest.strategy_guidance,
            })}
          </pre>
        </article>
      </section>

      <section className="agent-guide-grid">
        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.observerTitle}</p>
          <h2>{copy.observerTitle}</h2>
          <p className="hero-text">{copy.observerIntro}</p>

          <ul className="agent-list">
            {openClawManifest.public_context_endpoints.map((endpoint) => (
              <li key={endpoint}>
                <code>{endpoint}</code>
              </li>
            ))}
          </ul>
        </article>

        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.persistenceTitle}</p>
          <h2>{copy.persistenceTitle}</h2>
          <p className="hero-text">{copy.persistenceIntro}</p>

          <ul className="agent-list">
            <li>accounts: {String(openClawManifest.persistence.accounts)}</li>
            <li>sessions: {String(openClawManifest.persistence.sessions)}</li>
            <li>characters: {String(openClawManifest.persistence.characters)}</li>
            <li>quest_boards: {String(openClawManifest.persistence.quest_boards)}</li>
            <li>public_events: {String(openClawManifest.persistence.public_events)}</li>
          </ul>

          <p className="agent-note">{openClawManifest.persistence.note}</p>
        </article>
      </section>

      <section className="agent-guide-grid">
        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.recoveryTitle}</p>
          <h2>{copy.recoveryTitle}</h2>
          <p className="hero-text">{copy.recoveryIntro}</p>

          <ul className="agent-list">
            {openClawManifest.recovery_playbook.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </article>

        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.successTitle}</p>
          <h2>{copy.successTitle}</h2>
          <p className="hero-text">{copy.successIntro}</p>

          <ul className="agent-list">
            {copy.success.map((item) => (
              <li key={item}>{item}</li>
            ))}
          </ul>
        </article>
      </section>
    </main>
  );
}
