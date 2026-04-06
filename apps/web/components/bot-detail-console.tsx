"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import type {
  BotDetail,
  DungeonHistoryItem,
  EquipmentItemScore,
  PublicWorldState,
  QuestHistoryItem,
} from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  formatDateTime,
  formatMetric,
  formatRelativeTime,
  localizeClass,
  localizeRegionName,
  localizeWeapon,
} from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type BotDetailConsoleProps = {
  botID: string;
  worldState: PublicWorldState;
  detail: BotDetail;
  questHistory: QuestHistoryItem[];
  dungeonHistory: DungeonHistoryItem[];
};

type TodayTimelineEntry = {
  key: string;
  type: "quest" | "dungeon";
  title: string;
  occurredAt: string;
  subtitle: string;
  href?: string;
};

export default function BotDetailConsole({
  botID,
  worldState,
  detail,
  questHistory,
  dungeonHistory,
}: BotDetailConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const [todayFilter, setTodayFilter] = useState<"all" | "quests" | "dungeons">("all");
  const [showBackpack, setShowBackpack] = useState(false);

  const questsToday = detail.completed_quests_today;
  const dungeonsToday = detail.dungeon_runs_today;
  const seasonLevel =
    pickNumber(detail.stats_snapshot.season_level) ?? pickNumber(detail.character_summary.season_level);
  const seasonExp =
    pickNumber(detail.stats_snapshot.season_exp) ?? pickNumber(detail.character_summary.season_exp);
  const seasonExpToNext =
    pickNumber(detail.stats_snapshot.season_exp_to_next) ??
    pickNumber(detail.character_summary.season_exp_to_next);

  const stats = useMemo(
    () => [
      { label: "HP", value: formatMetric(detail.stats_snapshot.max_hp, language, "") },
      { label: language === "zh-CN" ? "物攻" : "PATK", value: formatMetric(detail.stats_snapshot.physical_attack, language, "") },
      { label: language === "zh-CN" ? "物防" : "PDEF", value: formatMetric(detail.stats_snapshot.physical_defense, language, "") },
      { label: language === "zh-CN" ? "魔攻" : "MATK", value: formatMetric(detail.stats_snapshot.magic_attack, language, "") },
      { label: language === "zh-CN" ? "魔防" : "MDEF", value: formatMetric(detail.stats_snapshot.magic_defense, language, "") },
      { label: language === "zh-CN" ? "速度" : "SPD", value: formatMetric(detail.stats_snapshot.speed, language, "") },
      { label: language === "zh-CN" ? "治疗" : "HEAL", value: formatMetric(detail.stats_snapshot.healing_power, language, "") },
    ],
    [detail.stats_snapshot, language],
  );

  const itemScoreByID = useMemo(() => {
    const map = new Map<string, EquipmentItemScore>();
    for (const item of detail.equipment_item_scores) {
      map.set(item.item_id, item);
    }
    return map;
  }, [detail.equipment_item_scores]);

  const equippedViews = useMemo(
    () =>
      detail.equipment.equipped.map((item, index) =>
        toEquipmentView(item, index, language, itemScoreByID),
      ),
    [detail.equipment.equipped, itemScoreByID, language],
  );

  const backpackViews = useMemo(
    () =>
      detail.equipment.inventory.map((item, index) =>
        toEquipmentView(item, index, language, itemScoreByID),
      ),
    [detail.equipment.inventory, itemScoreByID, language],
  );

  const equippedBySlot = useMemo(() => {
    const map = new Map<string, EquipmentView>();
    for (const item of equippedViews) {
      if (!map.has(item.slot)) {
        map.set(item.slot, item);
      }
    }
    return map;
  }, [equippedViews]);

  const totalEquipBonus = useMemo(() => {
    const totals = new Map<string, number>();
    for (const item of equippedViews) {
      for (const entry of item.statsEntries) {
        totals.set(entry.key, (totals.get(entry.key) ?? 0) + entry.value);
      }
    }

    return Array.from(totals.entries())
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, value]) => `${localizeStatKey(key, language)} +${value}`);
  }, [equippedViews, language]);

  const todayTimeline = useMemo(() => {
    const questEntries: TodayTimelineEntry[] = questsToday.map((quest) => ({
      key: `quest-${quest.quest_id}-${quest.submitted_at}`,
      type: "quest" as const,
      title: quest.quest_name,
      occurredAt: quest.submitted_at,
      subtitle: language === "zh-CN" ? "任务完成" : "Quest Completed",
    }));

    const dungeonEntries: TodayTimelineEntry[] = dungeonsToday.map((run) => ({
      key: `dungeon-${run.run_id}`,
      type: "dungeon" as const,
      title: run.dungeon_name,
      occurredAt: run.resolved_at,
      subtitle: language === "zh-CN" ? "副本结算" : "Dungeon Cleared",
      href: `/bots/${encodeURIComponent(botID)}/dungeon-runs/${encodeURIComponent(run.run_id)}`,
    }));

    return [...questEntries, ...dungeonEntries].sort((left, right) => {
      const rightTime = Date.parse(right.occurredAt);
      const leftTime = Date.parse(left.occurredAt);

      if (Number.isNaN(rightTime) || Number.isNaN(leftTime)) {
        return right.occurredAt.localeCompare(left.occurredAt);
      }

      return rightTime - leftTime;
    });
  }, [botID, dungeonsToday, language, questsToday]);

  const visibleTodayTimeline = useMemo(() => {
    if (todayFilter === "all") {
      return todayTimeline;
    }

    return todayTimeline.filter((entry) => (todayFilter === "quests" ? entry.type === "quest" : entry.type === "dungeon"));
  }, [todayFilter, todayTimeline]);

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="home"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={language === "zh-CN" ? "Bot 观察档案" : "Bot Observer File"}
        title={detail.character_summary.name}
        intro={
          language === "zh-CN"
            ? "查看 Bot 当前状态、装备背包、今日任务与副本战斗，并跟踪最近一周成长轨迹。"
            : "Inspect current status, equipment, backpack, today's quest/dungeon activities, and 7-day growth history."
        }
        stats={[
          {
            label: language === "zh-CN" ? "战斗力" : "Combat Power",
            value: formatMetric(detail.combat_power.panel_power_score, language, ""),
          },
          {
            label: language === "zh-CN" ? "当前区域" : "Current Region",
            value: localizeRegionName(
              detail.character_summary.location_region_id,
              detail.character_summary.location_region_id,
              language,
            ),
          },
          { label: language === "zh-CN" ? "声望" : "Reputation", value: formatMetric(detail.character_summary.reputation, language, "") },
          { label: language === "zh-CN" ? "金币" : "Gold", value: formatMetric(detail.character_summary.gold, language, "") },
        ]}
      />

      <section className="detail-stack single-column-grid">
        <section className="pixel-panel detail-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{language === "zh-CN" ? "角色总览" : "Profile"}</p>
              <h2>{language === "zh-CN" ? "角色信息 / 成长进度 / 属性快照" : "Identity / Progress / Stats"}</h2>
            </div>
          </div>

          <section className="detail-block profile-summary-block">
            <h3>{language === "zh-CN" ? "角色ID" : "Character ID"}</h3>
            <p className="profile-id-row">{detail.character_summary.character_id}</p>

            <div className="profile-kv-grid" style={{ marginBottom: 12 }}>
              <p className="profile-kv-row">
                <span>{language === "zh-CN" ? "总战斗力" : "Panel Combat Power"}</span>
                <strong>{formatMetric(detail.combat_power.panel_power_score, language, "")}</strong>
              </p>
              <p className="profile-kv-row">
                <span>{language === "zh-CN" ? "基础成长分" : "Base Growth"}</span>
                <strong>{formatMetric(detail.combat_power.base_growth_score, language, "")}</strong>
              </p>
              <p className="profile-kv-row">
                <span>{language === "zh-CN" ? "装备总分" : "Equipment Score"}</span>
                <strong>{formatMetric(detail.combat_power.equipment_score, language, "")}</strong>
              </p>
              <p className="profile-kv-row">
                <span>{language === "zh-CN" ? "构筑修正" : "Build Modifier"}</span>
                <strong>
                  {detail.combat_power.build_modifier_score >= 0 ? "+" : ""}
                  {formatMetric(detail.combat_power.build_modifier_score, language, "")}
                </strong>
              </p>
              <p className="profile-kv-row">
                <span>{language === "zh-CN" ? "竞技场预估" : "Arena Preview"}</span>
                <strong>{detail.combat_power.arena_preview.estimated_win_rate_band}</strong>
              </p>
              <p className="profile-kv-row">
                <span>{language === "zh-CN" ? "建议副本" : "Suggested Dungeon"}</span>
                <strong>
                  {detail.combat_power.dungeon_previews[0]
                    ? `${detail.combat_power.dungeon_previews[0].dungeon_name} (${detail.combat_power.dungeon_previews[0].estimated_clear_band})`
                    : language === "zh-CN"
                      ? "暂无"
                      : "N/A"}
                </strong>
              </p>
            </div>

            <div className="profile-main-grid">
              <section className="profile-side-block">
                <h3>{language === "zh-CN" ? "角色信息 / 成长进度" : "Identity / Growth"}</h3>
                <div className="profile-kv-grid">
                  <p className="profile-kv-row">
                    <span>{language === "zh-CN" ? "职业" : "Class"}</span>
                    <strong>{localizeClass(detail.character_summary.class, language)}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{language === "zh-CN" ? "武器流派" : "Weapon"}</span>
                    <strong>{localizeWeapon(detail.character_summary.weapon_style, language)}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{language === "zh-CN" ? "状态" : "Status"}</span>
                    <strong>{detail.character_summary.status || (language === "zh-CN" ? "未知" : "Unknown")}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{language === "zh-CN" ? "等级" : "Level"}</span>
                    <strong>{seasonLevel ?? "--"}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{language === "zh-CN" ? "当前经验" : "Current EXP"}</span>
                    <strong>{seasonExp ?? "--"}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{language === "zh-CN" ? "距下一级" : "To Next Level"}</span>
                    <strong>{seasonExpToNext ?? "--"}</strong>
                  </p>
                </div>
              </section>

              <section className="profile-side-block">
                <h3>{language === "zh-CN" ? "属性快照" : "Stats Snapshot"}</h3>
                <div className="profile-kv-grid compact-stats-grid">
                  {stats.map((item) => (
                    <p key={item.label} className="profile-kv-row">
                      <span>{item.label}</span>
                      <strong>{item.value}</strong>
                    </p>
                  ))}
                </div>
              </section>
            </div>
          </section>
        </section>

        <section className="pixel-panel detail-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{language === "zh-CN" ? "装备栏" : "Equipment Slots"}</p>
              <h2>{language === "zh-CN" ? "槽位常驻显示" : "Persistent Gear Slots"}</h2>
            </div>
            <button className="section-link" type="button" onClick={() => setShowBackpack((v) => !v)}>
              {showBackpack
                ? language === "zh-CN"
                  ? "收起背包"
                  : "Hide Backpack"
                : language === "zh-CN"
                  ? "打开背包"
                  : "Open Backpack"}
            </button>
          </div>

          <div className="detail-block equipment-total-block">
            <h3>{language === "zh-CN" ? "装备总加成" : "Total Equipment Bonus"}</h3>
            <p>
              {totalEquipBonus.length > 0
                ? totalEquipBonus.join(" · ")
                : language === "zh-CN"
                  ? "当前没有装备属性加成。"
                  : "No equipment bonus yet."}
            </p>
          </div>

          <div className="gear-slot-grid">
            {EQUIPMENT_SLOTS.map((slotDef) => {
              const item = equippedBySlot.get(slotDef.key);
              return (
                <article key={slotDef.key} className={`gear-slot-card ${item ? "filled" : "empty"}`}>
                  <div className="gear-slot-head">
                    <span className="gear-slot-icon">{slotDef.badge}</span>
                    <strong>{slotDef.label[language === "zh-CN" ? "zh" : "en"]}</strong>
                  </div>
                  <p className="gear-slot-name">
                    {item
                      ? `${item.name} · ${language === "zh-CN" ? "战力" : "Power"} ${item.powerScore}`
                      : language === "zh-CN"
                        ? "该槽位暂无装备"
                        : "Empty slot"}
                  </p>

                  <div className="gear-tooltip">
                    {item ? (
                      <>
                        <strong>{item.name}</strong>
                        <p>{item.meta}</p>
                        <p>{item.extra}</p>
                      </>
                    ) : (
                      <p>{language === "zh-CN" ? "空槽位，等待装备。" : "Empty slot, waiting for gear."}</p>
                    )}
                  </div>
                </article>
              );
            })}
          </div>

          {showBackpack ? (
            <section className="detail-block backpack-block">
              <h3>{language === "zh-CN" ? "背包物品" : "Backpack Items"}</h3>
              {backpackViews.length > 0 ? (
                <div className="backpack-grid">
                  {backpackViews.map((item) => (
                    <article key={item.key} className="backpack-item-card">
                      <strong>{item.name}</strong>
                      <span>{item.slotLabel}</span>
                      <div className="gear-tooltip">
                        <strong>{item.name}</strong>
                        <p>{item.meta}</p>
                        <p>{item.extra}</p>
                      </div>
                    </article>
                  ))}
                </div>
              ) : (
                <p>{language === "zh-CN" ? "背包为空。" : "Backpack is empty."}</p>
              )}
            </section>
          ) : null}
        </section>

        <section className="pixel-panel detail-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{language === "zh-CN" ? "行动与限制" : "Action & Limits"}</p>
              <h2>{language === "zh-CN" ? "今日限制与当前任务" : "Daily Limits & Active Quests"}</h2>
            </div>
          </div>
          <div className="detail-split-grid">
            <section className="detail-block">
              <h3>{language === "zh-CN" ? "每日限制" : "Daily Limits"}</h3>
              <p>
                {language === "zh-CN" ? "任务提交" : "Quest completion"}: {detail.daily_limits.quest_completion_used}/{detail.daily_limits.quest_completion_cap}
              </p>
              <p>
                {language === "zh-CN" ? "副本领奖" : "Dungeon claim"}: {detail.daily_limits.dungeon_entry_used}/{detail.daily_limits.dungeon_entry_cap}
              </p>
              <p>
                {language === "zh-CN" ? "重置时间" : "Reset at"}: {formatDateTime(detail.daily_limits.daily_reset_at, language)}
              </p>
            </section>
            <section className="detail-block">
              <h3>{language === "zh-CN" ? "当前任务" : "Active Quests"}</h3>
              {detail.active_quests.length > 0 ? (
                detail.active_quests.slice(0, 6).map((quest, index) => (
                  <p key={`active-quest-${index}`}>{String(quest["title"] ?? quest["quest_id"] ?? `#${index + 1}`)}</p>
                ))
              ) : (
                <p>{language === "zh-CN" ? "当前没有已接任务。" : "No active quests."}</p>
              )}
            </section>
          </div>
        </section>

        <section className="pixel-panel detail-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{language === "zh-CN" ? "今日记录" : "Today"}</p>
              <h2>{language === "zh-CN" ? "行动时间线" : "Action Timeline"}</h2>
            </div>
          </div>

          <div className="today-summary-grid">
            <article className="today-summary-card">
              <span>{language === "zh-CN" ? "全部事件" : "All Events"}</span>
              <strong>{todayTimeline.length}</strong>
            </article>
            <article className="today-summary-card">
              <span>{language === "zh-CN" ? "任务完成" : "Quests Completed"}</span>
              <strong>{questsToday.length}</strong>
            </article>
            <article className="today-summary-card">
              <span>{language === "zh-CN" ? "副本结算" : "Dungeon Clears"}</span>
              <strong>{dungeonsToday.length}</strong>
            </article>
          </div>

          <div className="board-tab-row today-filter-row">
            <button
              type="button"
              className={`board-tab ${todayFilter === "all" ? "active" : ""}`}
              onClick={() => setTodayFilter("all")}
            >
              {language === "zh-CN" ? "全部" : "All"}
            </button>
            <button
              type="button"
              className={`board-tab ${todayFilter === "quests" ? "active" : ""}`}
              onClick={() => setTodayFilter("quests")}
            >
              {language === "zh-CN" ? "今日任务完成" : "Completed Quests Today"}
            </button>
            <button
              type="button"
              className={`board-tab ${todayFilter === "dungeons" ? "active" : ""}`}
              onClick={() => setTodayFilter("dungeons")}
            >
              {language === "zh-CN" ? "今日副本战斗" : "Dungeon Combat Today"}
            </button>
          </div>

          {visibleTodayTimeline.length > 0 ? (
            <div className="today-timeline-list">
              {visibleTodayTimeline.map((entry) => {
                const heading = (
                  <>
                    <div className="today-entry-title-row">
                      <span className={`today-type-badge ${entry.type}`}>
                        {entry.type === "quest" ? (language === "zh-CN" ? "任务" : "Quest") : language === "zh-CN" ? "副本" : "Dungeon"}
                      </span>
                      <strong>{entry.title}</strong>
                    </div>
                    <p className="today-entry-meta">
                      {entry.subtitle} · {formatDateTime(entry.occurredAt, language)} · {formatRelativeTime(entry.occurredAt, language)}
                    </p>
                  </>
                );

                return (
                  <article key={entry.key} className="today-entry-card">
                    <span className={`today-entry-marker ${entry.type}`} aria-hidden="true" />
                    <div className="today-entry-body">
                      {entry.href ? (
                        <Link className="today-entry-link" href={entry.href}>
                          {heading}
                        </Link>
                      ) : (
                        heading
                      )}
                    </div>
                  </article>
                );
              })}
            </div>
          ) : (
            <p className="empty-state">
              {todayFilter === "quests"
                ? language === "zh-CN"
                  ? "今日暂无已完成任务。"
                  : "No completed quests today."
                : todayFilter === "dungeons"
                  ? language === "zh-CN"
                    ? "今日暂无副本战斗记录。"
                    : "No dungeon runs today."
                  : language === "zh-CN"
                    ? "今日暂无公开行动记录。"
                    : "No public action records today."}
            </p>
          )}
        </section>

        <section className="pixel-panel detail-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{language === "zh-CN" ? "历史轨迹" : "History"}</p>
              <h2>{language === "zh-CN" ? "最近 7 天成长" : "Latest 7 Days"}</h2>
            </div>
          </div>

          <div className="detail-split-grid">
            <section className="detail-block">
              <h3>{language === "zh-CN" ? "任务历史" : "Quest History"}</h3>
              {questHistory.length > 0 ? (
                questHistory.slice(0, 10).map((quest) => (
                  <p key={`${quest.quest_id}-${quest.submitted_at}`}>
                    {quest.quest_name} · {formatRelativeTime(quest.submitted_at, language)}
                  </p>
                ))
              ) : (
                <p>{language === "zh-CN" ? "暂无任务历史。" : "No quest history."}</p>
              )}
            </section>

            <section className="detail-block">
              <h3>{language === "zh-CN" ? "副本历史" : "Dungeon History"}</h3>
              {dungeonHistory.length > 0 ? (
                dungeonHistory.slice(0, 10).map((run) => (
                  <Link
                    key={run.run_id}
                    className="inline-link"
                    href={`/bots/${encodeURIComponent(botID)}/dungeon-runs/${encodeURIComponent(run.run_id)}`}
                  >
                    {run.dungeon_name} · {formatRelativeTime(run.resolved_at, language)}
                  </Link>
                ))
              ) : (
                <p>{language === "zh-CN" ? "暂无副本历史。" : "No dungeon history."}</p>
              )}
            </section>
          </div>

          <div className="detail-split-grid">
            <section className="detail-block">
              <h3>{language === "zh-CN" ? "最近事件" : "Recent Events"}</h3>
              {detail.recent_events.length > 0 ? (
                detail.recent_events.slice(0, 8).map((event) => (
                  <p key={event.event_id}>
                    {event.summary} · {formatRelativeTime(event.occurred_at, language)}
                  </p>
                ))
              ) : (
                <p>{language === "zh-CN" ? "暂无最近事件。" : "No recent events."}</p>
              )}
            </section>

            <section className="detail-block">
              <h3>{language === "zh-CN" ? "竞技场历史" : "Arena History"}</h3>
              {detail.arena_history.length > 0 ? (
                detail.arena_history.slice(0, 6).map((entry, index) => (
                  <p key={`arena-${index}`}>{String(entry["summary"] ?? entry["result"] ?? `#${index + 1}`)}</p>
                ))
              ) : (
                <p>{language === "zh-CN" ? "暂无竞技场记录。" : "No arena history."}</p>
              )}
            </section>
          </div>

          <div className="detail-block">
            <h3>{language === "zh-CN" ? "数据时间" : "Data Timestamp"}</h3>
            <p>{formatDateTime(worldState.server_time, language)}</p>
          </div>
        </section>
      </section>
    </main>
  );
}

type EquipmentView = {
  key: string;
  name: string;
  slot: string;
  slotLabel: string;
  meta: string;
  extra: string;
  powerScore: number;
  deltaVsEquipped: number;
  statsEntries: Array<{ key: string; value: number }>;
};

const EQUIPMENT_SLOTS: Array<{
  key: string;
  badge: string;
  label: { zh: string; en: string };
}> = [
  { key: "head", badge: "HD", label: { zh: "头部", en: "Head" } },
  { key: "chest", badge: "CH", label: { zh: "胸甲", en: "Chest" } },
  { key: "necklace", badge: "NK", label: { zh: "项链", en: "Necklace" } },
  { key: "ring", badge: "RG", label: { zh: "戒指", en: "Ring" } },
  { key: "boots", badge: "BT", label: { zh: "靴子", en: "Boots" } },
  { key: "weapon", badge: "WP", label: { zh: "武器", en: "Weapon" } },
];

function pickNumber(value: unknown): number | undefined {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }

  return undefined;
}

function toEquipmentView(
  item: Record<string, unknown>,
  index: number,
  language: string,
  itemScoreByID: Map<string, EquipmentItemScore>,
) {
  const itemID = String(item.item_id ?? item.catalog_id ?? `item-${index}`);
  const name = String(item.name ?? (language === "zh-CN" ? "未知物品" : "Unknown item"));
  const slot = String(item.slot ?? "-");
  const slotLabel = localizeSlot(slot, language);
  const rarity = String(item.rarity ?? "-");
  const enhancement = typeof item.enhancement_level === "number" ? item.enhancement_level : 0;
  const durability = typeof item.durability === "number" ? item.durability : 0;
  const state = String(item.state ?? "-");
  const statsEntries = toStatsEntries(item.stats);
  const stats = statsEntries.map((entry) => `${localizeStatKey(entry.key, language)}+${entry.value}`).join(", ");
  const score = itemScoreByID.get(itemID);
  const powerScore = score?.power_score ?? 0;
  const deltaVsEquipped = score?.delta_vs_equipped ?? 0;

  return {
    key: `${itemID}-${index}`,
    name,
    slot,
    slotLabel,
    statsEntries,
    powerScore,
    deltaVsEquipped,
    meta:
      language === "zh-CN"
        ? `槽位: ${slotLabel} · 稀有度: ${rarity} · 强化: +${enhancement} · 战力: ${powerScore}`
        : `Slot: ${slotLabel} · Rarity: ${rarity} · Enhance: +${enhancement} · Power: ${powerScore}`,
    extra:
      language === "zh-CN"
        ? `耐久: ${durability} · 状态: ${state} · 相对当前: ${deltaVsEquipped >= 0 ? "+" : ""}${deltaVsEquipped}${stats ? ` · 属性: ${stats}` : ""}`
        : `Durability: ${durability} · State: ${state} · Delta vs equipped: ${deltaVsEquipped >= 0 ? "+" : ""}${deltaVsEquipped}${stats ? ` · Stats: ${stats}` : ""}`,
  };
}

function toStatsEntries(value: unknown): Array<{ key: string; value: number }> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return [];
  }

  return Object.entries(value as Record<string, unknown>)
    .filter(([, statValue]) => typeof statValue === "number")
    .map(([statKey, statValue]) => ({ key: statKey, value: Number(statValue) }));
}

function localizeSlot(slot: string, language: string): string {
  const table: Record<string, { zh: string; en: string }> = {
    head: { zh: "头部", en: "Head" },
    chest: { zh: "胸甲", en: "Chest" },
    necklace: { zh: "项链", en: "Necklace" },
    boots: { zh: "靴子", en: "Boots" },
    ring: { zh: "戒指", en: "Ring" },
    weapon: { zh: "武器", en: "Weapon" },
  };

  const key = String(slot || "").toLowerCase();
  const fallback = language === "zh-CN" ? "未知槽位" : "Unknown slot";
  return table[key]?.[language === "zh-CN" ? "zh" : "en"] ?? (key || fallback);
}

function localizeStatKey(statKey: string, language: string): string {
  const table: Record<string, { zh: string; en: string }> = {
    max_hp: { zh: "生命", en: "HP" },
    max_mp: { zh: "法力", en: "MP" },
    physical_attack: { zh: "物攻", en: "PATK" },
    physical_defense: { zh: "物防", en: "PDEF" },
    magic_attack: { zh: "魔攻", en: "MATK" },
    magic_defense: { zh: "魔防", en: "MDEF" },
    speed: { zh: "速度", en: "SPD" },
    healing_power: { zh: "治疗", en: "HEAL" },
  };

  const normalized = String(statKey || "").toLowerCase();
  return table[normalized]?.[language === "zh-CN" ? "zh" : "en"] ?? normalized;
}
