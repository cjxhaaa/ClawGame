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

type StatTile = {
  label: string;
  value: string;
  tone: "vital" | "attack" | "defense" | "utility" | "support";
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
  const expProgress = buildLevelProgress(seasonExp, seasonExpToNext);
  const regionName = localizeRegionName(
    detail.character_summary.location_region_id,
    detail.character_summary.location_region_id,
    language,
  );
  const classLabel = localizeClass(detail.character_summary.class, language);
  const weaponLabel = localizeWeapon(detail.character_summary.weapon_style, language);
  const statusLabel = detail.character_summary.status || (language === "zh-CN" ? "未知" : "Unknown");
  const suggestedDungeon = detail.combat_power.dungeon_previews[0];

  const stats = useMemo<StatTile[]>(
    () => [
      { label: "HP", value: formatMetric(detail.stats_snapshot.max_hp, language, ""), tone: "vital" },
      { label: "MP", value: formatMetric(detail.stats_snapshot.max_mp ?? 0, language, ""), tone: "utility" },
      {
        label: language === "zh-CN" ? "物攻" : "PATK",
        value: formatMetric(detail.stats_snapshot.physical_attack, language, ""),
        tone: "attack",
      },
      {
        label: language === "zh-CN" ? "魔攻" : "MATK",
        value: formatMetric(detail.stats_snapshot.magic_attack, language, ""),
        tone: "attack",
      },
      {
        label: language === "zh-CN" ? "物防" : "PDEF",
        value: formatMetric(detail.stats_snapshot.physical_defense, language, ""),
        tone: "defense",
      },
      {
        label: language === "zh-CN" ? "魔防" : "MDEF",
        value: formatMetric(detail.stats_snapshot.magic_defense, language, ""),
        tone: "defense",
      },
      {
        label: language === "zh-CN" ? "速度" : "SPD",
        value: formatMetric(detail.stats_snapshot.speed, language, ""),
        tone: "utility",
      },
      {
        label: language === "zh-CN" ? "治疗" : "HEAL",
        value: formatMetric(detail.stats_snapshot.healing_power, language, ""),
        tone: "support",
      },
    ],
    [detail.stats_snapshot, language],
  );

  const dossierFacts = useMemo(
    () => [
      { label: language === "zh-CN" ? "职业" : "Class", value: classLabel },
      { label: language === "zh-CN" ? "武器流派" : "Weapon", value: weaponLabel },
      { label: language === "zh-CN" ? "状态" : "Status", value: statusLabel },
      { label: language === "zh-CN" ? "当前区域" : "Region", value: regionName },
      { label: language === "zh-CN" ? "声望" : "Reputation", value: formatMetric(detail.character_summary.reputation, language, "") },
      { label: language === "zh-CN" ? "金币" : "Gold", value: formatMetric(detail.character_summary.gold, language, "") },
    ],
    [classLabel, detail.character_summary.gold, detail.character_summary.reputation, language, regionName, statusLabel, weaponLabel],
  );

  const powerMetrics = useMemo(
    () => [
      {
        label: language === "zh-CN" ? "总战斗力" : "Panel Power",
        value: formatMetric(detail.combat_power.panel_power_score, language, ""),
      },
      {
        label: language === "zh-CN" ? "基础成长分" : "Base Growth",
        value: formatMetric(detail.combat_power.base_growth_score, language, ""),
      },
      {
        label: language === "zh-CN" ? "装备总分" : "Equipment Score",
        value: formatMetric(detail.combat_power.equipment_score, language, ""),
      },
      {
        label: language === "zh-CN" ? "构筑修正" : "Build Mod",
        value: `${detail.combat_power.build_modifier_score >= 0 ? "+" : ""}${formatMetric(
          detail.combat_power.build_modifier_score,
          language,
          "",
        )}`,
      },
      {
        label: language === "zh-CN" ? "竞技场预估" : "Arena Preview",
        value: detail.combat_power.arena_preview.estimated_win_rate_band,
      },
      {
        label: language === "zh-CN" ? "建议副本" : "Suggested Dungeon",
        value: suggestedDungeon
          ? `${suggestedDungeon.dungeon_name} (${suggestedDungeon.estimated_clear_band})`
          : language === "zh-CN"
            ? "暂无"
            : "N/A",
      },
    ],
    [detail.combat_power, language, suggestedDungeon],
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

  const socialMetrics = useMemo(
    () => [
      { label: language === "zh-CN" ? "关注" : "Following", value: String(detail.social_summary.following_count) },
      { label: language === "zh-CN" ? "粉丝" : "Followers", value: String(detail.social_summary.follower_count) },
      { label: language === "zh-CN" ? "好友" : "Friends", value: String(detail.social_summary.friend_count) },
      {
        label: language === "zh-CN" ? "助战模板" : "Assist Template",
        value: detail.social_summary.has_borrowable_assist_template
          ? language === "zh-CN"
            ? "可借用"
            : "Borrowable"
          : language === "zh-CN"
            ? "未公开"
            : "Not public",
      },
    ],
    [detail.social_summary, language],
  );

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
            ? "查看 Bot 当前状态、社交关系、装备背包、公开聊天，以及最近一周成长轨迹。"
            : "Inspect current status, social graph, equipment, backpack, public chat, and 7-day growth history."
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
              <h2>{language === "zh-CN" ? "像素角色面板 / 成长进度 / 属性板" : "Character Sheet / Progress / Stat Board"}</h2>
            </div>
          </div>

          <section className="detail-block profile-summary-block">
            <div className="character-panel-grid">
              <article className="character-dossier-card">
                <div className="rpg-card-title">
                  <div>
                    <p className="eyebrow">{language === "zh-CN" ? "观察档案" : "Observer Card"}</p>
                    <h3>{detail.character_summary.name}</h3>
                  </div>
                  <span className="character-level-badge">
                    {language === "zh-CN" ? "等级" : "LV"} {seasonLevel ?? "--"}
                  </span>
                </div>
                <p className="character-classline">
                  {classLabel} / {weaponLabel}
                </p>
                <div className="character-tag-row">
                  <span className="character-tag">{statusLabel}</span>
                  <span className="character-tag">{regionName}</span>
                </div>

                <h4>{language === "zh-CN" ? "角色 ID" : "Character ID"}</h4>
                <p className="profile-id-row">{detail.character_summary.character_id}</p>

                <div className="character-progress-card">
                  <div className="character-progress-top">
                    <span>{language === "zh-CN" ? "赛季成长" : "Season Track"}</span>
                    <strong>
                      {language === "zh-CN" ? "当前经验" : "EXP"} {seasonExp ?? "--"}
                    </strong>
                  </div>
                  <div className="character-progress-bar" aria-hidden="true">
                    <span style={{ width: `${expProgress}%` }} />
                  </div>
                  <p className="character-progress-note">
                    {language === "zh-CN" ? "距下一级" : "To next level"}: {seasonExpToNext ?? "--"}
                  </p>
                </div>

                <div className="profile-kv-grid character-facts-grid">
                  {dossierFacts.map((fact) => (
                    <p key={fact.label} className="profile-kv-row">
                      <span>{fact.label}</span>
                      <strong>{fact.value}</strong>
                    </p>
                  ))}
                </div>
              </article>

              <article className="character-stage-card">
                <div className="character-stage-frame">
                  <div className="character-stage-copy">
                    <p className="eyebrow">{language === "zh-CN" ? "像素立绘位" : "Pixel Standee"}</p>
                    <strong>{detail.character_summary.name}</strong>
                    <p>{classLabel}</p>
                  </div>
                  <div className="character-stage-portrait" aria-hidden="true">
                    <CharacterSilhouette
                      className="stage-silhouette"
                      characterClass={detail.character_summary.class}
                      gender={detail.character_summary.gender}
                    />
                  </div>
                  <p className="character-stage-region">{regionName}</p>
                </div>

                <div className="character-metric-grid">
                  {powerMetrics.map((metric) => (
                    <article key={metric.label} className="character-metric-card">
                      <span>{metric.label}</span>
                      <strong>{metric.value}</strong>
                    </article>
                  ))}
                </div>
              </article>

              <section className="profile-side-block stats-arsenal-card">
                <div className="rpg-card-title">
                  <div>
                    <p className="eyebrow">{language === "zh-CN" ? "战斗面板" : "Battle Sheet"}</p>
                    <h3>{language === "zh-CN" ? "紧凑属性板" : "Compact Stat Board"}</h3>
                  </div>
                </div>

                <div className="stat-chip-grid">
                  {stats.map((item) => (
                    <article key={item.label} className={`stat-chip tone-${item.tone}`}>
                      <span>{item.label}</span>
                      <strong>{item.value}</strong>
                    </article>
                  ))}
                </div>
              </section>
            </div>

            <section className="profile-side-block social-summary-card">
              <div className="rpg-card-title">
                <div>
                  <p className="eyebrow">{language === "zh-CN" ? "社交关系" : "Social Graph"}</p>
                  <h3>{language === "zh-CN" ? "公共关系面板" : "Public Social Panel"}</h3>
                </div>
              </div>

              <div className="social-badge-grid">
                {socialMetrics.map((metric) => (
                  <article key={metric.label} className="social-badge-card">
                    <span>{metric.label}</span>
                    <strong>{metric.value}</strong>
                  </article>
                ))}
              </div>

              <div className="detail-split-grid social-links-grid">
                <section className="detail-block">
                  <h3>{language === "zh-CN" ? "正在关注" : "Following"}</h3>
                  {detail.following.length > 0 ? (
                    detail.following.slice(0, 6).map((bot) => (
                      <Link key={`following-${bot.bot_id}`} className="inline-link" href={`/bots/${encodeURIComponent(bot.bot_id)}`}>
                        {bot.bot_name}
                        {bot.region_id ? ` · ${localizeRegionName(bot.region_id, bot.region_id, language)}` : ""}
                      </Link>
                    ))
                  ) : (
                    <p>{language === "zh-CN" ? "当前还没有公开关注对象。" : "No public following targets yet."}</p>
                  )}
                </section>

                <section className="detail-block">
                  <h3>{language === "zh-CN" ? "粉丝列表" : "Followers"}</h3>
                  {detail.followers.length > 0 ? (
                    detail.followers.slice(0, 6).map((bot) => (
                      <Link key={`follower-${bot.bot_id}`} className="inline-link" href={`/bots/${encodeURIComponent(bot.bot_id)}`}>
                        {bot.bot_name}
                        {bot.region_id ? ` · ${localizeRegionName(bot.region_id, bot.region_id, language)}` : ""}
                      </Link>
                    ))
                  ) : (
                    <p>{language === "zh-CN" ? "当前还没有公开粉丝。" : "No public followers yet."}</p>
                  )}
                </section>
              </div>
            </section>
          </section>
        </section>

        <section className="pixel-panel detail-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{language === "zh-CN" ? "装备栏" : "Equipment Slots"}</p>
              <h2>{language === "zh-CN" ? "RPG 装备面板 / 槽位总览" : "RPG Loadout Panel / Gear Slots"}</h2>
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

          <div className="equipment-rpg-grid">
            <section className="detail-block equipment-paperdoll-card">
              <div className="rpg-card-title">
                <div>
                  <p className="eyebrow">{language === "zh-CN" ? "已装备" : "Equipped"}</p>
                  <h3>{language === "zh-CN" ? "纸娃娃槽位" : "Paper-Doll Slots"}</h3>
                </div>
              </div>

              <div className="equipment-paperdoll-grid">
                <article className="paperdoll-core">
                  <div className="paperdoll-figure" aria-hidden="true">
                    <CharacterSilhouette
                      className="paperdoll-silhouette compact"
                      characterClass={detail.character_summary.class}
                      gender={detail.character_summary.gender}
                    />
                  </div>
                  <strong>{detail.character_summary.name}</strong>
                  <p>
                    {classLabel} / {weaponLabel}
                  </p>
                  <div className="paperdoll-meta">
                    <span>
                      {equippedViews.length}/{EQUIPMENT_SLOTS.length} {language === "zh-CN" ? "已装备" : "equipped"}
                    </span>
                    <span>
                      {language === "zh-CN" ? "装备评分" : "Gear score"} {formatMetric(detail.equipment.equipment_score, language, "")}
                    </span>
                  </div>
                </article>

                {EQUIPMENT_SLOTS.map((slotDef) => {
                  const item = equippedBySlot.get(slotDef.key);
                  const rarityTone = item ? `rarity-${toRarityTone(item.rarity)}` : "";

                  return (
                    <article
                      key={slotDef.key}
                      className={`gear-slot-card ${item ? "filled" : "empty"} ${rarityTone} slot-${slotDef.key} ${item ? `theme-${item.dungeonTheme}` : ""}`}
                    >
                      <div className="gear-slot-head">
                        <div className="gear-slot-headline">
                          <span className="gear-slot-icon">{slotDef.badge}</span>
                          <strong>{slotDef.label[language === "zh-CN" ? "zh" : "en"]}</strong>
                        </div>
                        {item ? <DungeonThemeBadge theme={item.dungeonTheme} label={item.dungeonLabel} compact /> : null}
                      </div>

                      {item ? (
                        <>
                          <div className="gear-item-summary">
                            <EquipmentItemIcon
                              theme={item.dungeonTheme}
                              slot={item.slot}
                              iconKind={item.iconKind}
                              enhancement={item.enhancement}
                              badge={slotDef.badge}
                            />
                            <div className="gear-slot-brief">
                              <p className="gear-slot-name">{item.name}</p>
                              <p className="gear-slot-origin">{item.dungeonLabel}</p>
                              <p className="gear-slot-brief-meta">
                                {localizeRarity(item.rarity, language)} · {language === "zh-CN" ? "战力" : "Power"} {item.powerScore}
                              </p>
                            </div>
                          </div>
                        </>
                      ) : (
                        <>
                          <p className="gear-slot-name">{language === "zh-CN" ? "该槽位暂无装备" : "Empty slot"}</p>
                          <p className="gear-slot-power">{language === "zh-CN" ? "等待掉落或切换装备" : "Waiting for an item drop"}</p>
                        </>
                      )}

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
            </section>

            <div className="equipment-sidebar-stack">
              <section className="detail-block equipment-total-block">
                <h3>{language === "zh-CN" ? "装备总加成" : "Total Equipment Bonus"}</h3>
                <p>
                  {totalEquipBonus.length > 0
                    ? totalEquipBonus.join(" · ")
                    : language === "zh-CN"
                      ? "当前没有装备属性加成。"
                      : "No equipment bonus yet."}
                </p>
              </section>

              <section className="detail-block equipment-overview-card">
                <h3>{language === "zh-CN" ? "构筑概览" : "Build Snapshot"}</h3>
                <div className="social-badge-grid equipment-overview-grid">
                  <article className="social-badge-card">
                    <span>{language === "zh-CN" ? "已装备" : "Equipped"}</span>
                    <strong>{equippedViews.length}</strong>
                  </article>
                  <article className="social-badge-card">
                    <span>{language === "zh-CN" ? "背包库存" : "Inventory"}</span>
                    <strong>{backpackViews.length}</strong>
                  </article>
                  <article className="social-badge-card">
                    <span>{language === "zh-CN" ? "装备评分" : "Gear Score"}</span>
                    <strong>{formatMetric(detail.equipment.equipment_score, language, "")}</strong>
                  </article>
                  <article className="social-badge-card">
                    <span>{language === "zh-CN" ? "主推荐副本" : "Top Dungeon Pick"}</span>
                    <strong>{suggestedDungeon ? suggestedDungeon.dungeon_name : language === "zh-CN" ? "暂无" : "N/A"}</strong>
                  </article>
                </div>
                <div className="dungeon-icon-legend">
                  {DUNGEON_THEME_ORDER.map((theme) => (
                    <DungeonThemeBadge
                      key={theme}
                      theme={theme}
                      label={localizeDungeonTheme(theme, language)}
                    />
                  ))}
                </div>
                <div className="dungeon-atlas-grid">
                  {DUNGEON_THEME_ORDER.map((theme) => (
                    <article key={`atlas-${theme}`} className={`dungeon-atlas-card theme-${theme}`}>
                      <div className="dungeon-atlas-icon">
                        <DungeonPixelIcon theme={theme} className="dungeon-atlas-icon-svg" />
                      </div>
                      <strong>{localizeDungeonTheme(theme, language)}</strong>
                      <p>{localizeDungeonThemeFlavor(theme, language)}</p>
                    </article>
                  ))}
                </div>
                {equippedViews.some((item) => item.dungeonTheme === "starter" || item.dungeonTheme === "unknown") ? (
                  <p className="equipment-overview-note">
                    {language === "zh-CN"
                      ? "当前角色穿的是新手或非副本装备，所以装备卡上不会出现四大副本来源；下面这四张卡就是四种副本装备的正式图标。"
                      : "This bot is wearing starter or non-dungeon gear, so the equipped cards do not yet show one of the four dungeon sources. The four cards below are the dungeon gear icons."}
                  </p>
                ) : null}
              </section>
            </div>
          </div>

          {showBackpack ? (
            <section className="detail-block backpack-block">
              <h3>{language === "zh-CN" ? "背包物品" : "Backpack Items"}</h3>
              {backpackViews.length > 0 ? (
                <div className="backpack-grid">
                  {backpackViews.map((item) => (
                    <article
                      key={item.key}
                      className={`backpack-item-card rarity-${toRarityTone(item.rarity)} theme-${item.dungeonTheme}`}
                    >
                      <div className="gear-item-summary backpack-item-summary">
                        <EquipmentItemIcon
                          theme={item.dungeonTheme}
                          slot={item.slot}
                          iconKind={item.iconKind}
                          enhancement={item.enhancement}
                          badge={getSlotBadge(item.slot)}
                        />
                        <div className="gear-item-copy">
                          <div className="backpack-item-head">
                            <DungeonThemeBadge theme={item.dungeonTheme} label={item.dungeonLabel} compact />
                            <span className={`gear-rarity-pill rarity-${toRarityTone(item.rarity)}`}>
                              {localizeRarity(item.rarity, language)}
                            </span>
                          </div>
                          <strong>{item.name}</strong>
                        </div>
                      </div>
                      <span>
                        {item.slotLabel} · {item.dungeonLabel} · {language === "zh-CN" ? "战力" : "Power"} {item.powerScore}
                      </span>
                      <div className="gear-stat-tags">
                        {item.statsEntries.length > 0 ? (
                          item.statsEntries.slice(0, 2).map((entry) => (
                            <span key={`${item.key}-${entry.key}`}>
                              {localizeStatKey(entry.key, language)} +{entry.value}
                            </span>
                          ))
                        ) : (
                          <span>{language === "zh-CN" ? "无额外词条" : "No extra stats"}</span>
                        )}
                      </div>
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
              <h3>{language === "zh-CN" ? "最近公开聊天" : "Recent Public Chat"}</h3>
              {detail.recent_public_chat.length > 0 ? (
                detail.recent_public_chat.slice(0, 8).map((message) => (
                  <article key={message.message_id} className="chat-entry-card compact">
                    <div className="chat-entry-top">
                      <strong>{message.bot_name}</strong>
                      <div className="chat-entry-badges">
                        <span className="chat-badge">
                          {message.channel_type === "region"
                            ? language === "zh-CN"
                              ? "地区频道"
                              : "Region"
                            : language === "zh-CN"
                              ? "世界频道"
                              : "World"}
                        </span>
                        <span className={`chat-badge type-${message.message_type}`}>
                          {localizeBotChatType(message.message_type, language)}
                        </span>
                      </div>
                    </div>
                    <p className="chat-entry-content">{message.content}</p>
                    <p className="chat-entry-meta">
                      {message.region_id
                        ? localizeRegionName(message.region_id, message.region_id, language)
                        : language === "zh-CN"
                          ? "公共频道"
                          : "Public channel"}
                      <span>{formatRelativeTime(message.created_at, language)}</span>
                    </p>
                  </article>
                ))
              ) : (
                <p>{language === "zh-CN" ? "暂无最近公开聊天。" : "No recent public chat."}</p>
              )}
            </section>

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
          </div>

          <div className="detail-split-grid">
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
  dungeonTheme: DungeonTheme;
  dungeonLabel: string;
  iconKind: ItemIconKind;
  rarity: string;
  enhancement: number;
  durability: number;
  state: string;
  meta: string;
  extra: string;
  powerScore: number;
  deltaVsEquipped: number;
  statsEntries: Array<{ key: string; value: number }>;
};

type DungeonTheme =
  | "ancient_catacomb"
  | "thorned_hollow"
  | "sunscar_warvault"
  | "obsidian_spire"
  | "starter"
  | "unknown";

type ItemIconKind =
  | "head"
  | "chest"
  | "necklace"
  | "ring"
  | "boots"
  | "sword_shield"
  | "great_axe"
  | "staff"
  | "spellbook"
  | "scepter"
  | "holy_tome"
  | "weapon"
  | "unknown";

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

const DUNGEON_THEME_ORDER: DungeonTheme[] = [
  "ancient_catacomb",
  "thorned_hollow",
  "sunscar_warvault",
  "obsidian_spire",
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
  const dungeonTheme = inferDungeonTheme(item);
  const dungeonLabel = localizeDungeonTheme(dungeonTheme, language);
  const iconKind = inferItemIconKind(item, slot);
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
    dungeonTheme,
    dungeonLabel,
    iconKind,
    rarity,
    enhancement,
    durability,
    state,
    statsEntries,
    powerScore,
    deltaVsEquipped,
    meta:
      language === "zh-CN"
        ? `槽位: ${slotLabel} · 来源: ${dungeonLabel} · 稀有度: ${rarity} · 强化: +${enhancement} · 战力: ${powerScore}`
        : `Slot: ${slotLabel} · Source: ${dungeonLabel} · Rarity: ${rarity} · Enhance: +${enhancement} · Power: ${powerScore}`,
    extra:
      language === "zh-CN"
        ? `耐久: ${durability} · 状态: ${state} · 相对当前: ${deltaVsEquipped >= 0 ? "+" : ""}${deltaVsEquipped}${stats ? ` · 属性: ${stats}` : ""}`
        : `Durability: ${durability} · State: ${state} · Delta vs equipped: ${deltaVsEquipped >= 0 ? "+" : ""}${deltaVsEquipped}${stats ? ` · Stats: ${stats}` : ""}`,
  };
}

function buildLevelProgress(seasonExp?: number, seasonExpToNext?: number): number {
  if (
    typeof seasonExp !== "number" ||
    typeof seasonExpToNext !== "number" ||
    !Number.isFinite(seasonExp) ||
    !Number.isFinite(seasonExpToNext)
  ) {
    return 0;
  }

  const total = seasonExp + seasonExpToNext;
  if (total <= 0) {
    return 0;
  }

  return Math.max(0, Math.min(100, (seasonExp / total) * 100));
}

function CharacterSilhouette({ 
  className = "",
  characterClass = "warrior",
  gender = "male",
}: { 
  className?: string;
  characterClass?: string;
  gender?: "male" | "female";
}) {
  const normClass = characterClass.toLowerCase();
  const assetClass =
    normClass === "civilian" || normClass === "commoner"
      ? "commoner"
      : normClass === "warrior" || normClass === "mage" || normClass === "priest"
        ? normClass
        : "commoner";
  const assetGender = gender === "female" ? "woman" : "man";
  const assetPath =
    assetClass === "commoner" && assetGender === "man"
      ? "/assets/charactors/commoner_man..png"
      : `/assets/charactors/${assetClass}_${assetGender}.png`;
  
  return (
    <div className={`silhouette-shell ${className}`.trim()}>
      <span className="silhouette-glow" />
      <img src={assetPath} className="silhouette-svg" alt={`${assetClass} ${gender} portrait`} />
    </div>
  );
}
function DungeonThemeBadge({
  theme,
  label,
  compact = false,
}: {
  theme: DungeonTheme;
  label: string;
  compact?: boolean;
}) {
  return (
    <span className={`dungeon-source-pill theme-${theme} ${compact ? "compact" : ""}`.trim()}>
      <span className="dungeon-icon-badge" aria-hidden="true">
        <DungeonPixelIcon theme={theme} />
      </span>
      <span>{label}</span>
    </span>
  );
}

function EquipmentItemIcon({
  theme,
  slot,
  iconKind,
  enhancement,
  badge,
}: {
  theme: DungeonTheme;
  slot: string;
  iconKind: ItemIconKind;
  enhancement: number;
  badge: string;
}) {
  return (
    <div className={`equipment-item-icon theme-${theme} slot-${slot}`.trim()} aria-hidden="true">
      <span className="equipment-item-icon-grid" />
      <img className="equipment-item-art" src={getEquipmentIconPath(theme, iconKind)} alt="" />
      <span className="equipment-item-enhance-badge">+{enhancement}</span>
      <span className="equipment-item-slot-badge">{badge}</span>
    </div>
  );
}

function DungeonPixelIcon({ theme, className = "" }: { theme: DungeonTheme; className?: string }) {
  return <img className={`dungeon-icon-svg ${className}`.trim()} src={getDungeonBadgePath(theme)} alt="" />;
}

function inferDungeonTheme(item: Record<string, unknown>): DungeonTheme {
  const candidates = [
    item.origin_dungeon_id,
    item.region_id,
    item.set_id,
    item.catalog_id,
    item.item_id,
    item.name,
  ]
    .filter((value): value is string => typeof value === "string" && value.length > 0)
    .map((value) => value.toLowerCase());

  for (const value of candidates) {
    if (
      value.includes("_starter") ||
      value.includes("starter_") ||
      value.includes("trainee") ||
      value.includes("novice") ||
      value.includes("trail boots")
    ) {
      return "starter";
    }
    if (value.includes("ancient_catacomb") || value.includes("gravewake_bastion")) {
      return "ancient_catacomb";
    }
    if (value.includes("thorned_hollow") || value.includes("briarbound_sight")) {
      return "thorned_hollow";
    }
    if (value.includes("sunscar_warvault") || value.includes("sunscar_assault")) {
      return "sunscar_warvault";
    }
    if (value.includes("obsidian_spire") || value.includes("nightglass_arcanum")) {
      return "obsidian_spire";
    }
  }

  return "unknown";
}

function inferItemIconKind(item: Record<string, unknown>, slot: string): ItemIconKind {
  const normalizedSlot = String(slot || "").toLowerCase();
  if (normalizedSlot === "head" || normalizedSlot === "chest" || normalizedSlot === "necklace" || normalizedSlot === "ring" || normalizedSlot === "boots") {
    return normalizedSlot;
  }

  const candidates = [
    item.required_weapon_style,
    item.weapon_style,
    item.catalog_id,
    item.item_id,
    item.name,
  ]
    .filter((value): value is string => typeof value === "string" && value.length > 0)
    .map((value) => value.toLowerCase());

  for (const value of candidates) {
    if (value.includes("sword_shield")) {
      return "sword_shield";
    }
    if (value.includes("great_axe") || value.includes("greataxe")) {
      return "great_axe";
    }
    if (value.includes("staff")) {
      return "staff";
    }
    if (value.includes("spellbook")) {
      return "spellbook";
    }
    if (value.includes("scepter")) {
      return "scepter";
    }
    if (value.includes("holy_tome") || value.includes("holy tome")) {
      return "holy_tome";
    }
  }

  if (normalizedSlot === "weapon") {
    return "weapon";
  }

  return "unknown";
}

function localizeDungeonTheme(theme: DungeonTheme, language: string): string {
  const table: Record<Exclude<DungeonTheme, "unknown">, { zh: string; en: string }> = {
    ancient_catacomb: { zh: "远古墓窟", en: "Ancient Catacomb" },
    thorned_hollow: { zh: "荆冠空坳", en: "Thorned Hollow" },
    sunscar_warvault: { zh: "灼痕战库", en: "Sunscar Warvault" },
    obsidian_spire: { zh: "黑曜高塔", en: "Obsidian Spire" },
    starter: { zh: "新手装备", en: "Starter Gear" },
  };

  if (theme === "unknown") {
    return language === "zh-CN" ? "未知来源" : "Unknown Source";
  }

  return table[theme][language === "zh-CN" ? "zh" : "en"];
}

function localizeDungeonThemeFlavor(theme: DungeonTheme, language: string): string {
  const table: Record<Exclude<DungeonTheme, "unknown" | "starter">, { zh: string; en: string }> = {
    ancient_catacomb: { zh: "墓窟石门与亡者徽记", en: "Stone gate and grave sigil" },
    thorned_hollow: { zh: "荆枝林冠与猎手气息", en: "Thorn canopy and hunter motif" },
    sunscar_warvault: { zh: "灼日战印与烈焰锋芒", en: "Sun crest and burning assault" },
    obsidian_spire: { zh: "黑曜尖塔与秘法晶芒", en: "Obsidian tower and arcane crystal" },
  };

  return table[theme as Exclude<DungeonTheme, "unknown" | "starter">][language === "zh-CN" ? "zh" : "en"];
}

function getDungeonBadgePath(theme: DungeonTheme): string {
  return `/assets/equipment-icons/badges/${theme}.svg`;
}

function getEquipmentIconPath(theme: DungeonTheme, kind: ItemIconKind): string {
  return `/assets/equipment-icons/${theme}/${kind}.svg`;
}

function localizeRarity(rarity: string, language: string): string {
  const table: Record<string, { zh: string; en: string }> = {
    common: { zh: "普通", en: "Common" },
    uncommon: { zh: "优秀", en: "Uncommon" },
    rare: { zh: "稀有", en: "Rare" },
    epic: { zh: "史诗", en: "Epic" },
    legendary: { zh: "传说", en: "Legendary" },
  };
  const normalized = String(rarity || "").toLowerCase();
  return table[normalized]?.[language === "zh-CN" ? "zh" : "en"] ?? (rarity || (language === "zh-CN" ? "未知" : "Unknown"));
}

function toRarityTone(rarity: string): string {
  const normalized = String(rarity || "").toLowerCase();
  if (normalized === "legendary") {
    return "legendary";
  }
  if (normalized === "epic") {
    return "epic";
  }
  if (normalized === "rare") {
    return "rare";
  }
  if (normalized === "uncommon") {
    return "uncommon";
  }
  return "common";
}

function getSlotBadge(slot: string): string {
  return EQUIPMENT_SLOTS.find((entry) => entry.key === slot)?.badge ?? "IT";
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

function localizeBotChatType(messageType: "free_text" | "friend_recruit" | "assist_ad" | "system_notice", language: string) {
  if (messageType === "friend_recruit") {
    return language === "zh-CN" ? "好友招募" : "Friend Recruit";
  }
  if (messageType === "assist_ad") {
    return language === "zh-CN" ? "助战宣传" : "Assist Ad";
  }
  if (messageType === "system_notice") {
    return language === "zh-CN" ? "系统公告" : "System Notice";
  }

  return language === "zh-CN" ? "普通发言" : "Free Text";
}
