"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import type { DungeonRunDetail, PublicWorldState } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import { formatDateTime, formatMetric } from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type DungeonRunDetailConsoleProps = {
  botID: string;
  runDetail: DungeonRunDetail;
  worldState: PublicWorldState;
};

type BattleFilter =
  | "all"
  | "action"
  | "player_action"
  | "enemy_action"
  | "turn"
  | "room"
  | "recovery"
  | "key";

type EventHighlight = "normal" | "positive" | "warning" | "danger";

function getStringValue(source: Record<string, unknown>, key: string): string {
  const value = source[key];
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  return "";
}

function getNumberValue(source: Record<string, unknown>, key: string): number | undefined {
  const value = source[key];
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return undefined;
}

function getMapValue(source: Record<string, unknown>, key: string): Record<string, unknown> {
  const value = source[key];
  if (value && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function formatEventType(raw: string): string {
  if (!raw) {
    return "event";
  }
  return raw
    .split("_")
    .map((part) => (part ? part[0].toUpperCase() + part.slice(1) : part))
    .join(" ");
}

function formatActionSummary(entry: Record<string, unknown>, language: string): string {
  const eventType = getStringValue(entry, "event_type");
  const actor = getStringValue(entry, "actor") || "system";
  const target = getStringValue(entry, "target");
  const action = getStringValue(entry, "action") || getStringValue(entry, "skill_id");
  const value = getNumberValue(entry, "value");
  const valueType = getStringValue(entry, "value_type");

  if (eventType === "action") {
    const amount = value !== undefined ? `${valueType === "heal" ? "+" : "-"}${value}` : "";
    if (language === "zh-CN") {
      return `${actor} -> ${target || "unknown"} | ${action || "action"} ${amount}`.trim();
    }
    return `${actor} -> ${target || "unknown"} | ${action || "action"} ${amount}`.trim();
  }

  if (eventType === "room_start") {
    return language === "zh-CN" ? "房间开战" : "Room combat started";
  }
  if (eventType === "turn_start") {
    return language === "zh-CN" ? "回合开始" : "Turn started";
  }
  if (eventType === "turn_end") {
    return language === "zh-CN" ? "回合结束" : "Turn ended";
  }
  if (eventType === "room_end") {
    const result = getStringValue(entry, "result");
    if (language === "zh-CN") {
      return `房间结算: ${result || "unknown"}`;
    }
    return `Room result: ${result || "unknown"}`;
  }
  if (eventType === "room_recovery") {
    return language === "zh-CN" ? "房间间隙恢复" : "Between-room recovery";
  }

  return getStringValue(entry, "message") || formatEventType(eventType || getStringValue(entry, "step"));
}

function formatCompactMap(source: Record<string, unknown>, fallback: string): string {
  const entries = Object.entries(source).filter(([, value]) => value !== undefined && value !== null);
  if (entries.length === 0) {
    return fallback;
  }
  return entries.map(([key, value]) => `${key}:${String(value)}`).join(" | ");
}

function eventGroupForType(eventType: string): Exclude<BattleFilter, "all" | "key"> {
  if (eventType === "action") {
    return "action";
  }
  if (eventType === "turn_start" || eventType === "turn_end") {
    return "turn";
  }
  if (eventType === "room_recovery") {
    return "recovery";
  }
  return "room";
}

function normalizedActor(entry: Record<string, unknown>): string {
  return getStringValue(entry, "actor").trim().toLowerCase();
}

function classifyEventHighlight(entry: Record<string, unknown>, eventType: string): {
  highlight: EventHighlight;
  isKey: boolean;
} {
  const valueType = getStringValue(entry, "value_type");
  const value = getNumberValue(entry, "value") ?? 0;
  const afterHP = getNumberValue(entry, "target_hp_after");
  const result = getStringValue(entry, "result");

  if (eventType === "room_end") {
    if (result === "failed" || result === "timeout") {
      return { highlight: "danger", isKey: true };
    }
    if (result === "cleared") {
      return { highlight: "positive", isKey: true };
    }
  }

  if (eventType === "action") {
    if (valueType === "damage" && (value >= 28 || afterHP === 0)) {
      return { highlight: "danger", isKey: true };
    }
    if (valueType === "heal" && value >= 20) {
      return { highlight: "positive", isKey: true };
    }
    if (valueType === "heal" || valueType === "damage") {
      return { highlight: "warning", isKey: false };
    }
  }

  if (eventType === "room_recovery") {
    return { highlight: "positive", isKey: false };
  }

  return { highlight: "normal", isKey: false };
}

export default function DungeonRunDetailConsole({
  botID,
  runDetail,
  worldState,
}: DungeonRunDetailConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const [battleFilter, setBattleFilter] = useState<BattleFilter>("all");
  const roomSummary = runDetail.room_summary ?? {};
  const battleState = runDetail.battle_state ?? {};

  const roomCount = getNumberValue(roomSummary, "room_count");
  const roomsCleared = getNumberValue(roomSummary, "rooms_cleared");
  const bossRoom = getNumberValue(roomSummary, "boss_room_index");
  const startHP = getNumberValue(battleState, "start_hp");
  const remainingHP = getNumberValue(battleState, "remaining_hp");
  const attempts = getNumberValue(battleState, "rooms_attempted");
  const finalResult = getStringValue(battleState, "final_result");
  const engineMode = getStringValue(battleState, "engine_mode") || "auto_turn_based";

  const timelineEntries = useMemo(
    () =>
      runDetail.battle_log.map((entry, index) => {
        const eventType = getStringValue(entry, "event_type") || getStringValue(entry, "step") || "event";
        const roomIndex = getNumberValue(entry, "room_index");
        const turn = getNumberValue(entry, "turn");
        const beforeHP = getNumberValue(entry, "target_hp_before");
        const afterHP = getNumberValue(entry, "target_hp_after");
        const playerHP = getNumberValue(entry, "player_hp");
        const enemyHP = getNumberValue(entry, "enemy_hp");
        const cooldownBefore = getMapValue(entry, "cooldown_before_round");
        const cooldownAfter = getMapValue(entry, "cooldown_after_round");
        const group = eventGroupForType(eventType);
        const { highlight, isKey } = classifyEventHighlight(entry, eventType);
        const actor = normalizedActor(entry);
        const isPlayerAction = eventType === "action" && actor === "player";
        const isEnemyAction = eventType === "action" && actor === "enemy";

        return {
          key: `battle-${index}`,
          eventType,
          actor,
          eventLabel: formatEventType(eventType),
          roomIndex,
          turn,
          group,
          highlight,
          isKey,
          isPlayerAction,
          isEnemyAction,
          summary: formatActionSummary(entry, language),
          hpDelta:
            beforeHP !== undefined && afterHP !== undefined
              ? `${beforeHP} -> ${afterHP}`
              : language === "zh-CN"
                ? "无"
                : "N/A",
          playerHP:
            playerHP !== undefined
              ? playerHP
              : language === "zh-CN"
                ? "未知"
                : "Unknown",
          enemyHP:
            enemyHP !== undefined
              ? enemyHP
              : language === "zh-CN"
                ? "未知"
                : "Unknown",
          cooldownBefore: formatCompactMap(cooldownBefore, language === "zh-CN" ? "无" : "none"),
          cooldownAfter: formatCompactMap(cooldownAfter, language === "zh-CN" ? "无" : "none"),
        };
      }),
    [language, runDetail.battle_log],
  );

  const filterOptions: Array<{ key: BattleFilter; label: string }> = [
    { key: "all", label: language === "zh-CN" ? "全部" : "All" },
    { key: "key", label: language === "zh-CN" ? "关键事件" : "Key Events" },
    { key: "action", label: language === "zh-CN" ? "行动" : "Actions" },
    { key: "player_action", label: language === "zh-CN" ? "玩家行动" : "Player Actions" },
    { key: "enemy_action", label: language === "zh-CN" ? "敌方行动" : "Enemy Actions" },
    { key: "turn", label: language === "zh-CN" ? "回合节点" : "Turn Marks" },
    { key: "room", label: language === "zh-CN" ? "房间节点" : "Room Marks" },
    { key: "recovery", label: language === "zh-CN" ? "恢复" : "Recovery" },
  ];

  const filteredTimelineEntries = useMemo(() => {
    if (battleFilter === "all") {
      return timelineEntries;
    }
    if (battleFilter === "key") {
      return timelineEntries.filter((entry) => entry.isKey);
    }
    if (battleFilter === "player_action") {
      return timelineEntries.filter((entry) => entry.isPlayerAction);
    }
    if (battleFilter === "enemy_action") {
      return timelineEntries.filter((entry) => entry.isEnemyAction);
    }
    return timelineEntries.filter((entry) => entry.group === battleFilter);
  }, [battleFilter, timelineEntries]);

  const filterCount = (key: BattleFilter): number => {
    if (key === "all") {
      return timelineEntries.length;
    }
    if (key === "key") {
      return timelineEntries.filter((entry) => entry.isKey).length;
    }
    if (key === "player_action") {
      return timelineEntries.filter((entry) => entry.isPlayerAction).length;
    }
    if (key === "enemy_action") {
      return timelineEntries.filter((entry) => entry.isEnemyAction).length;
    }
    return timelineEntries.filter((entry) => entry.group === key).length;
  };

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="home"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={language === "zh-CN" ? "副本运行详情" : "Dungeon Run Detail"}
        title={runDetail.dungeon_name}
        intro={
          language === "zh-CN"
            ? "查看单次副本自动战斗记录、关键里程碑与结算信息。"
            : "Inspect battle logs, milestones, and settlement for a single dungeon run."
        }
        stats={[
          { label: "Run ID", value: runDetail.run_id || "-" },
          { label: language === "zh-CN" ? "难度" : "Difficulty", value: runDetail.difficulty },
          { label: language === "zh-CN" ? "当前阶段" : "Phase", value: runDetail.result.runtime_phase },
        ]}
      />

      <section className="detail-page-grid">
        <section className="detail-stack">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{language === "zh-CN" ? "基础信息" : "Metadata"}</p>
                <h2>{language === "zh-CN" ? "副本基础信息" : "Run Metadata"}</h2>
              </div>
              <Link className="section-link" href={`/bots/${encodeURIComponent(botID)}`}>
                {language === "zh-CN" ? "返回 Bot 详情" : "Back to Bot"}
              </Link>
            </div>

            <div className="bot-identity-grid">
              <article className="hero-metric">
                <span>{language === "zh-CN" ? "进入时间" : "Started"}</span>
                <strong>{formatDateTime(runDetail.started_at, language)}</strong>
              </article>
              <article className="hero-metric">
                <span>{language === "zh-CN" ? "结算时间" : "Resolved"}</span>
                <strong>{formatDateTime(runDetail.resolved_at, language)}</strong>
              </article>
              <article className="hero-metric">
                <span>{language === "zh-CN" ? "可领奖" : "Reward Claimable"}</span>
                <strong>{runDetail.result.reward_claimable ? "Yes" : "No"}</strong>
              </article>
            </div>

            <div className="battle-meta-grid">
              <article className="battle-meta-card">
                <span>{language === "zh-CN" ? "引擎模式" : "Engine"}</span>
                <strong>{engineMode}</strong>
              </article>
              <article className="battle-meta-card">
                <span>{language === "zh-CN" ? "总房间 / 已清" : "Rooms"}</span>
                <strong>
                  {roomsCleared ?? "-"}/{roomCount ?? "-"}
                </strong>
              </article>
              <article className="battle-meta-card">
                <span>{language === "zh-CN" ? "首领房间" : "Boss Room"}</span>
                <strong>{bossRoom ?? "-"}</strong>
              </article>
              <article className="battle-meta-card">
                <span>{language === "zh-CN" ? "HP（开始/剩余）" : "HP (start/left)"}</span>
                <strong>
                  {startHP ?? "-"}/{remainingHP ?? "-"}
                </strong>
              </article>
              <article className="battle-meta-card">
                <span>{language === "zh-CN" ? "尝试房间数" : "Rooms Attempted"}</span>
                <strong>{attempts ?? "-"}</strong>
              </article>
              <article className="battle-meta-card">
                <span>{language === "zh-CN" ? "最终结果" : "Final Result"}</span>
                <strong>{finalResult || runDetail.result.run_status}</strong>
              </article>
            </div>
          </section>

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{language === "zh-CN" ? "战斗记录" : "Battle Log"}</p>
                <h2>{language === "zh-CN" ? "回合 / 阶段日志" : "Round / Phase Log"}</h2>
              </div>
            </div>

            <div className="battle-filter-row" role="tablist" aria-label={language === "zh-CN" ? "战斗日志过滤" : "Battle log filters"}>
              {filterOptions.map((option) => (
                <button
                  key={option.key}
                  type="button"
                  className={`battle-filter-chip ${battleFilter === option.key ? "active" : ""}`}
                  onClick={() => setBattleFilter(option.key)}
                >
                  <span>{option.label}</span>
                  <strong>{filterCount(option.key)}</strong>
                </button>
              ))}
            </div>

            <div className="battle-entry-list">
              {filteredTimelineEntries.length > 0 ? (
                filteredTimelineEntries.map((entry) => (
                  <article key={entry.key} className={`battle-entry-card ${entry.highlight}`}>
                    <div className="battle-entry-head">
                      <span className={`battle-event-chip ${entry.highlight}`}>{entry.eventLabel}</span>
                      <span className="battle-entry-ref">
                        {language === "zh-CN" ? "房间" : "Room"} {entry.roomIndex ?? "-"} · {language === "zh-CN" ? "回合" : "Turn"} {entry.turn ?? "-"}
                      </span>
                    </div>
                    <p className="battle-entry-summary">{entry.summary}</p>
                    <div className="battle-entry-stats">
                      <span>{language === "zh-CN" ? "目标HP变化" : "Target HP"}: {entry.hpDelta}</span>
                      <span>{language === "zh-CN" ? "玩家HP" : "Player HP"}: {String(entry.playerHP)}</span>
                      <span>{language === "zh-CN" ? "敌方HP" : "Enemy HP"}: {String(entry.enemyHP)}</span>
                    </div>
                    <div className="battle-entry-cooldown">
                      <span>{language === "zh-CN" ? "冷却前" : "CD Before"}: {entry.cooldownBefore}</span>
                      <span>{language === "zh-CN" ? "冷却后" : "CD After"}: {entry.cooldownAfter}</span>
                    </div>
                  </article>
                ))
              ) : (
                <p className="empty-state">
                  {battleFilter === "key"
                    ? language === "zh-CN"
                      ? "当前没有命中的关键事件。"
                      : "No key events matched."
                    : battleFilter === "player_action"
                      ? language === "zh-CN"
                        ? "当前没有玩家行动日志。"
                        : "No player action entries."
                      : battleFilter === "enemy_action"
                        ? language === "zh-CN"
                          ? "当前没有敌方行动日志。"
                          : "No enemy action entries."
                    : language === "zh-CN"
                      ? "当前筛选下没有日志。"
                      : "No log entries for the current filter."}
                </p>
              )}
            </div>
          </section>
        </section>

        <aside className="detail-sidebar">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{language === "zh-CN" ? "结算" : "Settlement"}</p>
                <h2>{language === "zh-CN" ? "里程碑与奖励" : "Milestones & Rewards"}</h2>
              </div>
            </div>

            <div className="detail-block">
              <h3>{language === "zh-CN" ? "关键里程碑" : "Milestones"}</h3>
              {runDetail.milestones.length > 0 ? (
                runDetail.milestones.map((item, index) => (
                  <p key={`milestone-${index}`}>{String(item["message"] ?? item["rating"] ?? JSON.stringify(item))}</p>
                ))
              ) : (
                <p>{language === "zh-CN" ? "暂无里程碑记录。" : "No milestones."}</p>
              )}
            </div>

            <div className="detail-block">
              <h3>{language === "zh-CN" ? "评分" : "Rating"}</h3>
              <p>
                {language === "zh-CN" ? "当前评级" : "Current rating"}: {runDetail.result.current_rating || "-"}
              </p>
              <p>
                {language === "zh-CN" ? "预测评级" : "Projected rating"}: {runDetail.result.projected_rating || "-"}
              </p>
            </div>

            <div className="detail-block">
              <h3>{language === "zh-CN" ? "奖励摘要" : "Reward Summary"}</h3>
              <p>
                {language === "zh-CN" ? "待发放评级奖励" : "Pending rating rewards"}: {formatMetric(runDetail.reward_summary.pending_rating_rewards.length, language, "")}
              </p>
              <p>
                {language === "zh-CN" ? "材料掉落条目" : "Staged material drops"}: {formatMetric(runDetail.reward_summary.staged_material_drops.length, language, "")}
               </p>
             </div>
 
             <div className="detail-block">
               <h3>{language === "zh-CN" ? "数据时间" : "Data Timestamp"}</h3>
               <p>{formatDateTime(worldState.server_time, language)}</p>
             </div>
           </section>
         </aside>
       </section>
     </main>
   );
 }
