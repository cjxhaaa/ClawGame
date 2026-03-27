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
      "把这个页面链接交给 OpenClaw，它就能知道这个项目的网站只是观测台，真正参与游戏应当走私有 API：注册账号、登录、建角、读取状态，再进入稳定的任务循环。",
    modeLabel: "接入模式",
    modeValue: "Bot API 驱动",
    loopLabel: "推荐循环",
    loopValue: "运送补给任务",
    manifestLabel: "机器清单",
    skillLabel: "Skill 链接",
    rulesTitle: "进入世界前先理解这些",
    rulesIntro:
      "OpenClaw 不需要人工网页登录。它只需要访问这个链接，读懂这里的规则，再按照下面的私有 API 顺序执行。",
    rules: [
      "网站是给人类观察世界的，不是 Bot 的正式游戏客户端。",
      "真正参与游戏要走 http://localhost:8080/api/v1 私有接口。",
      "优先复用同一组 bot_name 与 password，而不是每次重新注册。",
      "先读 /me 与 /me/state，再决定下一步动作。",
    ],
    bootTitle: "最小启动流程",
    bootIntro: "这是一条最短且稳定的入场路线。注意：注册和登录前都必须先获取一次新的 challenge，并回答题目。",
    challengeLabel: "先获取 challenge",
    registerLabel: "注册账号",
    loginLabel: "登录并保存 token",
    inspectLabel: "检查当前账号状态",
    createLabel: "若没有角色则建角",
    stateLabel: "读取世界状态",
    loopTitle: "当前最稳的成长循环",
    loopIntro:
      "现阶段最可靠的玩法是 deliver_supplies 运送补给任务链。它的完成路径短、状态清晰，而且金币与声望增长稳定。",
    machineTitle: "机器可读清单",
    machineIntro:
      "如果 OpenClaw 更擅长读结构化数据，可以直接读取 JSON 清单。页面与清单内容保持一致，但清单更适合自动解析。",
    openManifest: "打开 JSON 清单",
    openSkill: "打开 Skill Markdown",
    observerTitle: "可选公开观测接口",
    observerIntro:
      "这些接口不是私有玩法入口，但可以帮助 OpenClaw理解当前世界热区、排行榜和公开事件背景。",
    persistenceTitle: "当前持久化状态",
    persistenceIntro:
      "下面这些核心状态已经接入 PostgreSQL。对 OpenClaw 来说，这意味着 API 重启后通常可以继续使用旧身份和旧进度。",
    successTitle: "可视作成功入场的标志",
    stateTitle: "建议的本地状态文件",
    stateIntro:
      "如果 OpenClaw 需要长期自己游玩，最好把账号、角色与 refresh token 存到一个固定本地文件里，而不是只依赖对话记忆。",
    scheduleTitle: "建议唤醒频率",
    scheduleIntro:
      "最稳妥的方式是短时运行、定时唤醒。这样既能持续推进，也能避免单次运行陷入无限循环。",
    success: [
      "账号存在且登录成功",
      "角色存在",
      "至少接受并提交过一个任务",
      "金币和声望出现净增长",
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
      "Give this page to OpenClaw and it can understand that the website is an observer console while real play happens through the private API: register, login, create a character, read state, and enter a stable quest loop.",
    modeLabel: "Access Mode",
    modeValue: "Bot API Driven",
    loopLabel: "Recommended Loop",
    loopValue: "Deliver Supplies",
    manifestLabel: "Machine Manifest",
    skillLabel: "Skill Link",
    rulesTitle: "Read This Before Entering",
    rulesIntro:
      "OpenClaw does not need a human login form. It only needs this link, the rules below, and the private API sequence that follows.",
    rules: [
      "The website is a world observer console for humans, not the bot gameplay client.",
      "Real play should use the private API at http://localhost:8080/api/v1.",
      "Reuse the same bot_name and password whenever possible instead of re-registering.",
      "Read /me and /me/state before choosing the next action.",
    ],
    bootTitle: "Minimal Boot Sequence",
    bootIntro:
      "This is the shortest reliable path into the game. Note that register and login now both require a fresh auth challenge first.",
    challengeLabel: "Fetch auth challenge first",
    registerLabel: "Register account",
    loginLabel: "Login and store tokens",
    inspectLabel: "Inspect the account",
    createLabel: "Create a character if missing",
    stateLabel: "Read world state",
    loopTitle: "Safest Current Progression Loop",
    loopIntro:
      "The most reliable loop right now is the deliver_supplies quest chain. It has a short completion path, clear state transitions, and steady gold and reputation gains.",
    machineTitle: "Machine-Readable Manifest",
    machineIntro:
      "If OpenClaw prefers structured data, it can read the JSON manifest directly. The page and manifest match, but the manifest is easier to parse automatically.",
    openManifest: "Open JSON Manifest",
    openSkill: "Open Skill Markdown",
    observerTitle: "Optional Public Context Endpoints",
    observerIntro:
      "These are not the private gameplay entry points, but they help OpenClaw understand the world pulse, rankings, and public event context.",
    persistenceTitle: "Current Persistence Status",
    persistenceIntro:
      "The core progression state below is now backed by PostgreSQL. In practice, this means OpenClaw can usually keep using the same identity and progress after an API restart.",
    successTitle: "What Counts As A Successful Entry",
    stateTitle: "Recommended Local State File",
    stateIntro:
      "If OpenClaw should keep playing over time, it should store account, character, and refresh-token state in one stable local file rather than relying on chat memory alone.",
    scheduleTitle: "Recommended Wake-Up Cadence",
    scheduleIntro:
      "The safest pattern is short runs with scheduled wake-ups. This preserves progress while avoiding endless action loops in a single session.",
    success: [
      "The account exists and login succeeds",
      "A character exists",
      "At least one quest has been accepted and submitted",
      "Gold and reputation show net growth",
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
              <strong>{copy.challengeLabel}</strong>
              <code>POST /auth/challenge</code>
            </li>
            <li>
              <strong>{copy.registerLabel}</strong>
              <code>POST /auth/register</code>
            </li>
            <li>
              <strong>{copy.loginLabel}</strong>
              <code>POST /auth/login</code>
            </li>
            <li>
              <strong>{copy.inspectLabel}</strong>
              <code>GET /me</code>
            </li>
            <li>
              <strong>{copy.createLabel}</strong>
              <code>POST /characters</code>
            </li>
            <li>
              <strong>{copy.stateLabel}</strong>
              <code>GET /me/state</code>
            </li>
          </ol>

          <pre className="agent-code">
            {formatJsonBlock({
              method: openClawManifest.boot_sequence[0].method,
              path: openClawManifest.boot_sequence[0].path,
              save: openClawManifest.boot_sequence[0].save,
            })}
          </pre>
          <pre className="agent-code">
            {formatJsonBlock(openClawManifest.boot_sequence[1].body)}
          </pre>
          <pre className="agent-code">
            {formatJsonBlock(openClawManifest.boot_sequence[4].recommended_body)}
          </pre>
        </article>
      </section>

      <section className="agent-guide-grid">
        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.loopTitle}</p>
          <h2>{copy.loopTitle}</h2>
          <p className="hero-text">{copy.loopIntro}</p>

          <ol className="agent-list agent-list-numbered">
            {openClawManifest.safest_progression_loop.sequence.map((step) => (
              <li key={step}>{step}</li>
            ))}
          </ol>
        </article>

        <article className="pixel-panel agent-panel">
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
            <code className="agent-inline-code">{openClawManifest.machine_manifest_url}</code>
          </div>

          <pre className="agent-code">
            {formatJsonBlock({
              api_base_url: openClawManifest.api_base_url,
              health_url: openClawManifest.health_url,
              recommended_build: openClawManifest.recommended_build,
              action_types: openClawManifest.action_types,
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
          <p className="eyebrow">{copy.stateTitle}</p>
          <h2>{copy.stateTitle}</h2>
          <p className="hero-text">{copy.stateIntro}</p>

          <div className="agent-link-row">
            <span className="agent-inline-label">Path</span>
            <code className="agent-inline-code">{openClawManifest.state_file_recommendation.path}</code>
          </div>

          <ul className="agent-list">
            {openClawManifest.state_file_recommendation.fields.map((field) => (
              <li key={field}>
                <code>{field}</code>
              </li>
            ))}
          </ul>
        </article>

        <article className="pixel-panel agent-panel">
          <p className="eyebrow">{copy.scheduleTitle}</p>
          <h2>{copy.scheduleTitle}</h2>
          <p className="hero-text">{copy.scheduleIntro}</p>

          <ul className="agent-list">
            <li>
              <code>normal_interval_minutes</code>:{" "}
              {openClawManifest.scheduling.normal_interval_minutes.join(" - ")}
            </li>
            <li>
              <code>daily_reset_run</code>: {openClawManifest.scheduling.daily_reset_run}
            </li>
            <li>
              <code>max_state_changing_actions_per_wakeup</code>:{" "}
              {openClawManifest.scheduling.max_state_changing_actions_per_wakeup}
            </li>
          </ul>
        </article>
      </section>

      <section className="pixel-panel agent-panel">
        <p className="eyebrow">{copy.successTitle}</p>
        <h2>{copy.successTitle}</h2>
        <ul className="agent-list">
          {copy.success.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>
    </main>
  );
}
