"use client";

import Link from "next/link";

import type { BotDetail, DungeonRunDetail, QuestHistoryItem, WorldEvent } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  getRegionAtlasDossier,
  eventTypeToFilter,
  filters,
  formatDateTime,
  formatMetric,
  formatRelativeTime,
  localizeClass,
  localizeEventSummary,
  localizeRegionName,
  localizeWeapon,
  uiText,
} from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type EventDetailConsoleProps = {
  event: WorldEvent;
  actorDetail?: BotDetail | null;
  questRecord?: QuestHistoryItem | null;
  runDetail?: DungeonRunDetail | null;
  returnFeedHref: string;
  returnRegionHref?: string | null;
};

export default function EventDetailConsole({
  event,
  actorDetail,
  questRecord,
  runDetail,
  returnFeedHref,
  returnRegionHref,
}: EventDetailConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const common = uiText[language].common;
  const filterKey = eventTypeToFilter(event.event_type);
  const filterLabel = filters.find((item) => item.key === filterKey)?.label[language] ?? filterKey;
  const payload = event.payload ?? {};
  const questID = getString(payload, "quest_id");
  const questTitle = getString(payload, "quest_title");
  const runID = getString(payload, "run_id");
  const dungeonID = getString(payload, "dungeon_id");
  const fromRegionID = getString(payload, "from_region_id");
  const rewardGold = getNumber(payload, "reward_gold");
  const rewardReputation = getNumber(payload, "reward_reputation");
  const travelCostGold = getNumber(payload, "travel_cost_gold");
  const rating = getString(payload, "rating") || getString(payload, "current_rating");
  const rewardItems = getStringArray(payload, "reward_item_catalog_ids").concat(getStringArray(payload, "reward_items"));
  const materialDrops = getObjectArray(payload, "material_drops");
  const historyRun = runID ? actorDetail?.dungeon_history_7d.find((item) => item.run_id === runID) ?? null : null;
  const displayDungeonName = runDetail?.dungeon_name || historyRun?.dungeon_name || dungeonID || copyUnknown(language);
  const displayDifficulty = runDetail?.difficulty || copyUnknown(language);
  const displayRunStatus = runDetail?.result.run_status || historyRun?.result || copyUnknown(language);
  const displayRating = runDetail?.result.current_rating || rating || historyRun?.reward_summary?.rating || "-";
  const displayRewardGold = rewardGold ?? historyRun?.reward_summary?.gold;
  const battlePreview = (runDetail?.battle_log ?? []).slice(0, 10);
  const payloadEntries = Object.entries(payload).filter(([, value]) => value !== undefined && value !== null);
  const actorName = event.actor_name ?? common.unknownActor;
  const regionName = event.region_id ? localizeRegionName(event.region_id, event.region_id, language) : copyUnknown(language);
  const regionAtlas = event.region_id ? getRegionAtlasDossier(event.region_id, language) : null;
  const actorRole = actorDetail
    ? `${localizeClass(actorDetail.character_summary.class, language)} · ${localizeWeapon(
        actorDetail.character_summary.weapon_style,
        language,
      )}`
    : copyUnknown(language);
  const actorReputation = actorDetail
    ? formatMetric(actorDetail.character_summary.reputation, language, "")
    : copyUnknown(language);
  const questStatus = questStatusLabel(event.event_type, language, buildQuestCopy(language));
  const displayRunStatusLabel = localizeRunStatus(displayRunStatus, language);
  const visibilityLabel = localizeVisibility(event.visibility ?? "public", language);
  const pageHeadline = buildEventHeadline({
    language,
    actorName,
    regionName,
    event,
    questName: questRecord?.quest_name ?? questTitle,
    dungeonName: displayDungeonName,
  });
  const archiveRows = [
    { label: language === "zh-CN" ? "事件编号" : "Event ID", value: event.event_id },
    { label: language === "zh-CN" ? "归档类型" : "Record type", value: event.event_type },
    { label: language === "zh-CN" ? "公开范围" : "Visibility", value: visibilityLabel },
    { label: language === "zh-CN" ? "副本运行" : "Run ID", value: runID || "-" },
    { label: language === "zh-CN" ? "任务编号" : "Quest ID", value: questID || "-" },
  ].filter((row) => row.value && row.value !== "-");

  const copy =
    language === "zh-CN"
      ? {
          eyebrow: "世界公报",
          title: "单事件纪要",
          intro: "这里把一条公开事件整理成可阅读的世界公报，优先呈现人物、地点、结果与后续影响，而不是接口字段本身。",
          eventRecord: "事件公报",
          eventFacts: "现场纪要",
          eventResult: "结果判定",
          actorDesk: "主角档案",
          relatedDesk: "继续追踪",
          questResult: "委托结算",
          dungeonResult: "远征战报",
          battleIntel: "战况摘录",
          rawPayload: "档案附注",
          archiveIndex: "档案索引",
          archiveOpen: "展开技术附注",
          backToFeed: "返回事件流",
          backToRegion: "返回原区域",
          openBot: "查看人物详情",
          openRun: "查看完整副本记录",
          occurredAt: "发生时间",
          eventType: "事件类别",
          visibility: "公开范围",
          actor: "行动者",
          region: "地点",
          eventID: "事件 ID",
          questName: "任务名称",
          questStatus: "任务状态",
          rewardGold: "金币奖励",
          rewardReputation: "声望奖励",
          relatedDungeon: "关联副本",
          runId: "副本运行 ID",
          dungeonName: "副本",
          difficulty: "难度",
          runStatus: "运行状态",
          rating: "评级",
          rewardClaimable: "可领奖",
          pendingRewards: "待发放奖励",
          materialDrops: "材料掉落",
          travelFrom: "出发地点",
          travelTo: "到达地点",
          travelCost: "旅行费用",
          noPayload: "这条事件没有额外附注。",
          noBattle: "公开渠道只放出了结算，没有逐回合战报。",
          yes: "是",
          no: "否",
          unknown: "未知",
          accepted: "已接取",
          completed: "已完成",
          submitted: "已提交",
          genericResult: "当前能公开展示的事实已经整理在下方，如需技术附注，可展开底部档案。",
          actorIdentity: "人物身份",
          actorPosition: "当前驻地",
          actorRank: "当前声望",
          sceneTitle: "场景背景",
          sceneNote: "事件发生地的公开档案会影响人们如何理解这条消息。",
          headlineTitle: "事件纪要",
          witnessTitle: "围观者能读到什么",
          witnessNote: "这部分把时间、人物、地点和结果整理成可阅读的公报口吻。",
          rewardItems: "战利品",
          materialDropsTitle: "材料",
          resultHeadline: "本次结果",
          techNote: "技术附注保留给调试和追档，不参与主叙事。",
          travelRecord: "行程记录",
          battleNoteLead: "公开战况",
        }
      : {
          eyebrow: "World Bulletin",
          title: "Single Event Report",
          intro: "This page reshapes one public event into a readable in-world bulletin, foregrounding actor, place, outcome, and aftermath instead of raw interface fields.",
          eventRecord: "Event bulletin",
          eventFacts: "Scene memo",
          eventResult: "Verdict",
          actorDesk: "Actor dossier",
          relatedDesk: "Continue tracking",
          questResult: "Quest settlement",
          dungeonResult: "Expedition report",
          battleIntel: "Battle extracts",
          rawPayload: "Archive appendix",
          archiveIndex: "Archive index",
          archiveOpen: "Open technical appendix",
          backToFeed: "Back to feed",
          backToRegion: "Back to region",
          openBot: "Open bot detail",
          openRun: "Open full dungeon record",
          occurredAt: "Occurred",
          eventType: "Event type",
          visibility: "Visibility",
          actor: "Actor",
          region: "Region",
          eventID: "Event ID",
          questName: "Quest name",
          questStatus: "Quest status",
          rewardGold: "Gold reward",
          rewardReputation: "Reputation reward",
          relatedDungeon: "Linked dungeon",
          runId: "Run ID",
          dungeonName: "Dungeon",
          difficulty: "Difficulty",
          runStatus: "Run status",
          rating: "Rating",
          rewardClaimable: "Reward claimable",
          pendingRewards: "Pending rewards",
          materialDrops: "Material drops",
          travelFrom: "From",
          travelTo: "To",
          travelCost: "Travel cost",
          noPayload: "No extra appendix is attached to this event.",
          noBattle: "Public channels only exposed the settlement, not the turn-by-turn combat log.",
          yes: "Yes",
          no: "No",
          unknown: "Unknown",
          accepted: "Accepted",
          completed: "Completed",
          submitted: "Submitted",
          genericResult: "The currently public-facing facts are organized below. Technical appendix stays hidden unless needed.",
          actorIdentity: "Identity",
          actorPosition: "Current station",
          actorRank: "Current reputation",
          sceneTitle: "Scene backdrop",
          sceneNote: "The place dossier shapes how this event is read by outside observers.",
          headlineTitle: "Event chronicle",
          witnessTitle: "What observers can read",
          witnessNote: "This section retells the public facts as an in-world bulletin.",
          rewardItems: "Spoils",
          materialDropsTitle: "Materials",
          resultHeadline: "Outcome",
          techNote: "The appendix is preserved for debugging and traceability, not as the main reading mode.",
          travelRecord: "Travel record",
          battleNoteLead: "Public combat note",
        };

  const heroStats = [
    { label: copy.actor, value: actorName },
    { label: copy.region, value: regionName },
    { label: copy.resultHeadline, value: buildOutcomeLabel(event, displayRunStatusLabel, questStatus, language) },
  ];

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="events"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={copy.eyebrow}
        title={pageHeadline}
        intro={buildHeadlineIntro({
          language,
          actorName,
          regionName,
          regionAtlasIntro: regionAtlas?.shortIntro,
        })}
        stats={heroStats}
      />

      <section className="detail-page-grid">
        <section className="detail-stack">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.eventRecord}</p>
                <h2>{copy.eventRecord}</h2>
              </div>
            </div>

            <div className="event-link-row">
              <Link className="inline-link" href={returnFeedHref}>
                {copy.backToFeed}
              </Link>
              {returnRegionHref ? (
                <Link className="inline-link" href={returnRegionHref}>
                  {copy.backToRegion}
                </Link>
              ) : null}
            </div>

            <div className="detail-split-grid">
              <section className="detail-block">
                <h3>{copy.headlineTitle}</h3>
                <p className="event-story-lead">
                  {buildEventChronicle({
                    language,
                    actorName,
                    regionName,
                    event,
                    questName: questRecord?.quest_name ?? questTitle,
                    questStatus,
                    displayDungeonName,
                    displayRunStatus: displayRunStatusLabel,
                    displayRating,
                    displayRewardGold,
                    rewardReputation,
                    fromRegionName: fromRegionID
                      ? localizeRegionName(fromRegionID, fromRegionID, language)
                      : copy.unknown,
                    toRegionName: regionName,
                    travelCostGold,
                  })}
                </p>
                <p className="event-story-note">{copy.genericResult}</p>
              </section>

              <section className="detail-block">
                <h3>{copy.eventFacts}</h3>
                <div className="profile-kv-grid event-dossier-grid">
                  <p className="profile-kv-row">
                    <span>{copy.occurredAt}</span>
                    <strong>{formatDateTime(event.occurred_at, language)}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{copy.eventType}</span>
                    <strong>{filterLabel}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{copy.actor}</span>
                    <strong>{actorName}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{copy.region}</span>
                    <strong>{regionName}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{copy.visibility}</span>
                    <strong>{visibilityLabel}</strong>
                  </p>
                  <p className="profile-kv-row">
                    <span>{copy.resultHeadline}</span>
                    <strong>{buildOutcomeLabel(event, displayRunStatusLabel, questStatus, language)}</strong>
                  </p>
                </div>
              </section>
            </div>
          </section>

          {regionAtlas ? (
            <section className="pixel-panel detail-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{copy.sceneTitle}</p>
                  <h2>{copy.sceneTitle}</h2>
                </div>
              </div>

              <div className="detail-split-grid">
                <section className="detail-block">
                  <h3>{regionName}</h3>
                  <p className="event-story-lead">{regionAtlas.shortIntro}</p>
                  <p className="event-story-note">{copy.sceneNote}</p>
                </section>
                <section className="detail-block">
                  <h3>{copy.witnessTitle}</h3>
                  <div className="event-chip-row">
                    <span className="event-pill">{regionAtlas.terrainBand}</span>
                    <span className="event-pill">{regionAtlas.riskTier}</span>
                    <span className="event-pill">{regionAtlas.primaryActivity}</span>
                  </div>
                  <p className="event-story-note">{copy.witnessNote}</p>
                </section>
              </div>
            </section>
          ) : null}

          {isQuestEvent(event.event_type) ? (
            <section className="pixel-panel detail-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{copy.questResult}</p>
                  <h2>{copy.questResult}</h2>
                </div>
              </div>

              <div className="today-summary-grid">
                <article className="today-summary-card">
                  <span>{copy.questStatus}</span>
                  <strong>{questStatus}</strong>
                </article>
                <article className="today-summary-card">
                  <span>{copy.rewardGold}</span>
                  <strong>{rewardGold !== undefined ? formatMetric(rewardGold, language, language === "zh-CN" ? " 金" : "g") : "-"}</strong>
                </article>
                <article className="today-summary-card">
                  <span>{copy.rewardReputation}</span>
                  <strong>{rewardReputation !== undefined ? formatMetric(rewardReputation, language, "") : "-"}</strong>
                </article>
              </div>

              <div className="detail-split-grid">
                <section className="detail-block">
                  <h3>{copy.questName}</h3>
                  <p>{questRecord?.quest_name ?? questTitle ?? copy.unknown}</p>
                  <p className="region-subnote">
                    {copy.questStatus}: <strong>{questStatus}</strong>
                  </p>
                  {dungeonID ? <p className="region-subnote">{copy.relatedDungeon}: {dungeonID}</p> : null}
                </section>

                <section className="detail-block">
                  <h3>{copy.eventResult}</h3>
                  {questRecord?.reward_summary ? (
                    <>
                      <p>
                        {copy.rewardGold}: {formatMetric(questRecord.reward_summary.gold ?? 0, language, language === "zh-CN" ? " 金" : "g")}
                      </p>
                      <p>
                        {copy.rewardReputation}: {formatMetric(questRecord.reward_summary.reputation ?? 0, language, "")}
                      </p>
                    </>
                  ) : (
                    <p>{localizeEventSummary(event.summary, language)}</p>
                  )}
                </section>
              </div>
            </section>
          ) : null}

          {isDungeonEvent(event.event_type) || runDetail ? (
            <section className="pixel-panel detail-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{copy.dungeonResult}</p>
                  <h2>{copy.dungeonResult}</h2>
                </div>
                {event.actor_character_id && runID && runDetail ? (
                  <Link
                    className="section-link"
                    href={`/bots/${encodeURIComponent(event.actor_character_id)}/dungeon-runs/${encodeURIComponent(runID)}`}
                  >
                    {copy.openRun}
                  </Link>
                ) : null}
              </div>

              <div className="battle-meta-grid">
                <article className="battle-meta-card">
                  <span>{copy.dungeonName}</span>
                  <strong>{displayDungeonName}</strong>
                </article>
                <article className="battle-meta-card">
                  <span>{copy.difficulty}</span>
                  <strong>{displayDifficulty}</strong>
                </article>
                <article className="battle-meta-card">
                  <span>{copy.runStatus}</span>
                  <strong>{displayRunStatusLabel}</strong>
                </article>
                <article className="battle-meta-card">
                  <span>{copy.rating}</span>
                  <strong>{displayRating}</strong>
                </article>
                <article className="battle-meta-card">
                  <span>{copy.rewardClaimable}</span>
                  <strong>{runDetail ? (runDetail.result.reward_claimable ? copy.yes : copy.no) : copy.unknown}</strong>
                </article>
                <article className="battle-meta-card">
                  <span>{copy.runId}</span>
                  <strong>{runID || runDetail?.run_id || "-"}</strong>
                </article>
              </div>

              <div className="detail-split-grid">
                <section className="detail-block">
                  <h3>{copy.eventResult}</h3>
                  {displayRewardGold !== undefined ? (
                    <p>
                      {copy.rewardGold}: {formatMetric(displayRewardGold, language, language === "zh-CN" ? " 金" : "g")}
                    </p>
                  ) : (
                    <p>{localizeEventSummary(event.summary, language)}</p>
                  )}
                  <p>
                    {copy.pendingRewards}: {formatMetric(runDetail?.reward_summary.pending_rating_rewards.length ?? 0, language, "")}
                  </p>
                  <p>
                    {copy.materialDrops}: {formatMetric((runDetail?.reward_summary.staged_material_drops.length ?? 0) + materialDrops.length, language, "")}
                  </p>
                </section>

                <section className="detail-block">
                  <h3>{copy.rewardItems}</h3>
                  {rewardItems.length > 0 ? (
                    <div className="event-chip-row">
                      {rewardItems.map((item) => (
                        <span key={item} className="event-pill">
                          {item}
                        </span>
                      ))}
                    </div>
                  ) : null}
                  {rewardItems.length === 0 ? <p>{copy.noPayload}</p> : null}
                  {materialDrops.length > 0 ? (
                    <>
                      <p className="region-subnote">{copy.materialDropsTitle}</p>
                      <div className="event-chip-row">
                        {materialDrops.map((item, index) => (
                          <span key={`${runID || event.event_id}-drop-${index}`} className="event-pill">
                            {compactDrop(item)}
                          </span>
                        ))}
                      </div>
                    </>
                  ) : null}
                </section>
              </div>

              <section className="detail-block">
                <h3>{copy.battleIntel}</h3>
                {battlePreview.length > 0 ? (
                  <div className="battle-entry-list">
                    {battlePreview.map((entry, index) => (
                      <article key={`${runDetail?.run_id ?? runID}-battle-${index}`} className="battle-entry-card">
                        <div className="battle-entry-head">
                          <span className="battle-event-chip">
                            {String(entry["event_type"] ?? entry["step"] ?? "event")}
                          </span>
                          <span className="battle-entry-ref">
                            {compactRef(entry)}
                          </span>
                        </div>
                        <p className="battle-entry-summary">{summarizeBattleEntry(entry, language)}</p>
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className="event-story-note">{copy.noBattle}</p>
                )}
              </section>
            </section>
          ) : null}

          {event.event_type.startsWith("travel") ? (
            <section className="pixel-panel detail-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{copy.travelRecord}</p>
                  <h2>{copy.travelRecord}</h2>
                </div>
              </div>

              <div className="detail-split-grid">
                <section className="detail-block">
                  <h3>{copy.headlineTitle}</h3>
                  <p className="event-story-lead">
                    {buildTravelChronicle({
                      language,
                      actorName,
                      fromRegionName: fromRegionID ? localizeRegionName(fromRegionID, fromRegionID, language) : copy.unknown,
                      toRegionName: regionName,
                      travelCostGold,
                    })}
                  </p>
                </section>
                <section className="detail-block">
                  <h3>{copy.eventFacts}</h3>
                  <p>{copy.travelFrom}: {fromRegionID ? localizeRegionName(fromRegionID, fromRegionID, language) : copy.unknown}</p>
                  <p>{copy.travelTo}: {regionName}</p>
                  <p className="region-subnote">
                    {copy.travelCost}:{" "}
                    {travelCostGold !== undefined
                      ? formatMetric(travelCostGold, language, language === "zh-CN" ? " 金" : "g")
                      : "-"}
                  </p>
                </section>
              </div>
            </section>
          ) : null}

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.rawPayload}</p>
                <h2>{copy.rawPayload}</h2>
              </div>
            </div>

            <p className="section-note">{copy.techNote}</p>

            <details className="event-archive-panel">
              <summary>{copy.archiveOpen}</summary>
              <div className="detail-split-grid">
                <section className="detail-block">
                  <h3>{copy.archiveIndex}</h3>
                  <div className="profile-kv-grid event-dossier-grid">
                    {archiveRows.length > 0 ? (
                      archiveRows.map((row) => (
                        <p key={`${row.label}-${row.value}`} className="profile-kv-row">
                          <span>{row.label}</span>
                          <strong>{row.value}</strong>
                        </p>
                      ))
                    ) : (
                      <p className="empty-state">{copy.noPayload}</p>
                    )}
                  </div>
                </section>
                <section className="detail-block">
                  <h3>{copy.rawPayload}</h3>
                  {payloadEntries.length > 0 ? (
                    <pre className="agent-code event-payload-block">{JSON.stringify(payload, null, 2)}</pre>
                  ) : (
                    <p className="empty-state">{copy.noPayload}</p>
                  )}
                </section>
              </div>
            </details>
          </section>
        </section>

        <aside className="detail-sidebar">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.actorDesk}</p>
                <h2>{copy.actorDesk}</h2>
              </div>
            </div>

            <div className="detail-block">
              <p>
                <strong>{actorName}</strong>
              </p>
              {actorDetail ? (
                <>
                  <p>{copy.actorIdentity}: {actorRole}</p>
                  <p>{copy.actorRank}: {actorReputation}</p>
                  <p>
                    {copy.actorPosition}:{" "}
                    {localizeRegionName(
                      actorDetail.character_summary.location_region_id,
                      actorDetail.character_summary.location_region_id,
                      language,
                    )}
                  </p>
                  <Link className="inline-link" href={`/bots/${actorDetail.character_summary.character_id}`}>
                    {copy.openBot}
                  </Link>
                </>
              ) : (
                <p>{copy.unknown}</p>
              )}
            </div>
          </section>

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.relatedDesk}</p>
                <h2>{copy.relatedDesk}</h2>
              </div>
            </div>

            <div className="atlas-list">
              <Link className="atlas-link" href={returnFeedHref}>
                <strong>{copy.backToFeed}</strong>
                <span>{filterLabel}</span>
              </Link>
              {returnRegionHref && event.region_id ? (
                <Link className="atlas-link" href={returnRegionHref}>
                  <strong>{copy.backToRegion}</strong>
                  <span>{localizeRegionName(event.region_id, event.region_id, language)}</span>
                </Link>
              ) : null}
            </div>
          </section>
        </aside>
      </section>
    </main>
  );
}

function getString(source: Record<string, unknown>, key: string) {
  const value = source[key];
  if (typeof value === "string") {
    return value;
  }
  if (typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  return "";
}

function getNumber(source: Record<string, unknown>, key: string) {
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

function getStringArray(source: Record<string, unknown>, key: string) {
  const value = source[key];
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((item): item is string => typeof item === "string");
}

function getObjectArray(source: Record<string, unknown>, key: string) {
  const value = source[key];
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((item): item is Record<string, unknown> => Boolean(item) && typeof item === "object" && !Array.isArray(item));
}

function compactRecord(record: Record<string, unknown>) {
  return Object.entries(record)
    .filter(([, value]) => value !== undefined && value !== null)
    .map(([key, value]) => `${key}:${String(value)}`)
    .join(" | ");
}

function compactDrop(record: Record<string, unknown>) {
  const key = getString(record, "material_key") || getString(record, "item") || "drop";
  const quantity = getNumber(record, "quantity");
  return quantity !== undefined ? `${key} x${quantity}` : key;
}

function compactRef(record: Record<string, unknown>) {
  const roomIndex = getNumber(record, "room_index");
  const turn = getNumber(record, "turn");
  const parts = [];
  if (roomIndex !== undefined) parts.push(`room ${roomIndex}`);
  if (turn !== undefined) parts.push(`turn ${turn}`);
  return parts.join(" · ") || "timeline";
}

function summarizeBattleEntry(record: Record<string, unknown>, language: "zh-CN" | "en-US") {
  const eventType = getString(record, "event_type") || getString(record, "step");
  const action = getString(record, "action") || getString(record, "skill_id");
  const actor = getString(record, "actor");
  const target = getString(record, "target");
  const value = getNumber(record, "value");
  const valueType = getString(record, "value_type");
  const result = getString(record, "result");
  const message = getString(record, "message");

  if (message) {
    return message;
  }

  if (eventType === "action") {
    const amount = value !== undefined ? `${valueType === "heal" ? "+" : "-"}${value}` : "";
    return language === "zh-CN"
      ? `${actor || "unknown"} 对 ${target || "unknown"} 使用 ${action || "action"} ${amount}`.trim()
      : `${actor || "unknown"} used ${action || "action"} on ${target || "unknown"} ${amount}`.trim();
  }

  if (eventType === "room_end") {
    return language === "zh-CN" ? `房间结算：${result || "unknown"}` : `Room result: ${result || "unknown"}`;
  }

  return compactRecord(record);
}

function isQuestEvent(eventType: string) {
  return eventType.startsWith("quest");
}

function isDungeonEvent(eventType: string) {
  return eventType.startsWith("dungeon");
}

function questStatusLabel(eventType: string, language: "zh-CN" | "en-US", copy: Record<string, string>) {
  if (eventType === "quest.accepted") return copy.accepted;
  if (eventType === "quest.completed") return copy.completed;
  if (eventType === "quest.submitted") return copy.submitted;
  return language === "zh-CN" ? "任务事件" : "Quest event";
}

function copyUnknown(language: "zh-CN" | "en-US") {
  return language === "zh-CN" ? "未知" : "Unknown";
}

function buildQuestCopy(language: "zh-CN" | "en-US") {
  return language === "zh-CN"
    ? { accepted: "已接取", completed: "已完成", submitted: "已提交" }
    : { accepted: "Accepted", completed: "Completed", submitted: "Submitted" };
}

function buildOutcomeLabel(
  event: WorldEvent,
  dungeonStatus: string,
  questStatus: string,
  language: "zh-CN" | "en-US",
) {
  if (isQuestEvent(event.event_type)) return questStatus;
  if (isDungeonEvent(event.event_type)) return dungeonStatus || copyUnknown(language);
  if (event.event_type.startsWith("travel")) return language === "zh-CN" ? "已抵达" : "Arrived";
  return language === "zh-CN" ? "已记录" : "Recorded";
}

function buildHeadlineIntro(args: {
  language: "zh-CN" | "en-US";
  actorName: string;
  regionName: string;
  regionAtlasIntro?: string;
}) {
  if (args.language === "zh-CN") {
    return `${args.actorName} 的这条公开记录发生在 ${args.regionName}。${args.regionAtlasIntro ?? "围观者能读到人物、地点、结果与后续影响。"}`
      .trim();
  }

  return `${args.actorName}'s public record surfaced in ${args.regionName}. ${
    args.regionAtlasIntro ?? "Observers can read the actor, place, result, and immediate aftermath here."
  }`.trim();
}

function buildEventHeadline(args: {
  language: "zh-CN" | "en-US";
  actorName: string;
  regionName: string;
  event: WorldEvent;
  questName?: string;
  dungeonName?: string;
}) {
  const questName = args.questName || copyUnknown(args.language);
  const dungeonName = args.dungeonName || copyUnknown(args.language);

  if (args.event.event_type === "quest.accepted") {
    return args.language === "zh-CN"
      ? `${args.actorName} 在 ${args.regionName} 接下了「${questName}」`
      : `${args.actorName} accepted "${questName}" in ${args.regionName}`;
  }
  if (args.event.event_type === "quest.completed") {
    return args.language === "zh-CN"
      ? `${args.actorName} 在 ${args.regionName} 完成了「${questName}」`
      : `${args.actorName} completed "${questName}" in ${args.regionName}`;
  }
  if (args.event.event_type === "quest.submitted") {
    return args.language === "zh-CN"
      ? `${args.actorName} 在 ${args.regionName} 交付了「${questName}」`
      : `${args.actorName} turned in "${questName}" in ${args.regionName}`;
  }
  if (args.event.event_type === "dungeon.entered") {
    return args.language === "zh-CN"
      ? `${args.actorName} 进入了 ${dungeonName}`
      : `${args.actorName} entered ${dungeonName}`;
  }
  if (args.event.event_type === "dungeon.cleared") {
    return args.language === "zh-CN"
      ? `${args.actorName} 通关了 ${dungeonName}`
      : `${args.actorName} cleared ${dungeonName}`;
  }
  if (args.event.event_type === "dungeon.loot_granted") {
    return args.language === "zh-CN"
      ? `${args.actorName} 完成了 ${dungeonName} 的奖励结算`
      : `${args.actorName} settled rewards from ${dungeonName}`;
  }
  if (args.event.event_type.startsWith("travel")) {
    return args.language === "zh-CN"
      ? `${args.actorName} 抵达了 ${args.regionName}`
      : `${args.actorName} arrived in ${args.regionName}`;
  }

  return args.language === "zh-CN"
    ? `${args.actorName} 在 ${args.regionName} 留下了一条新记录`
    : `${args.actorName} left a new record in ${args.regionName}`;
}

function buildEventChronicle(args: {
  language: "zh-CN" | "en-US";
  actorName: string;
  regionName: string;
  event: WorldEvent;
  questName?: string;
  questStatus: string;
  displayDungeonName: string;
  displayRunStatus: string;
  displayRating: string;
  displayRewardGold?: number;
  rewardReputation?: number;
  fromRegionName: string;
  toRegionName: string;
  travelCostGold?: number;
}) {
  if (isQuestEvent(args.event.event_type)) {
    return buildQuestChronicle({
      language: args.language,
      actorName: args.actorName,
      regionName: args.regionName,
      questName: args.questName || copyUnknown(args.language),
      questStatus: args.questStatus,
      rewardGold: args.displayRewardGold,
      rewardReputation: args.rewardReputation,
    });
  }

  if (isDungeonEvent(args.event.event_type)) {
    return buildDungeonChronicle({
      language: args.language,
      actorName: args.actorName,
      regionName: args.regionName,
      dungeonName: args.displayDungeonName,
      runStatus: args.displayRunStatus,
      rating: args.displayRating,
      rewardGold: args.displayRewardGold,
      rewardItems: [],
      materialDropsCount: 0,
    });
  }

  if (args.event.event_type.startsWith("travel")) {
    return buildTravelChronicle({
      language: args.language,
      actorName: args.actorName,
      fromRegionName: args.fromRegionName,
      toRegionName: args.toRegionName,
      travelCostGold: args.travelCostGold,
    });
  }

  return args.language === "zh-CN"
    ? `${args.actorName} 在 ${args.regionName} 留下了一条新的公开记录。围观者目前能确认的主信息，是这次行动已经完成并被写入世界日志。`
    : `${args.actorName} left a fresh public record in ${args.regionName}. Observers can confirm that the action has already resolved and entered the world log.`;
}

function buildQuestChronicle(args: {
  language: "zh-CN" | "en-US";
  actorName: string;
  regionName: string;
  questName: string;
  questStatus: string;
  rewardGold?: number;
  rewardReputation?: number;
}) {
  if (args.language === "zh-CN") {
    const rewardBits = [
      args.rewardGold !== undefined ? `${args.rewardGold} 金` : "",
      args.rewardReputation !== undefined ? `${args.rewardReputation} 声望` : "",
    ].filter(Boolean);
    return `${args.actorName} 在 ${args.regionName} 对委托「${args.questName}」完成了${args.questStatus}。${
      rewardBits.length > 0 ? `公开结算显示，此次回报为 ${rewardBits.join("、")}。` : "当前公开记录没有额外奖励细节。"
    }`;
  }

  const rewardBits = [
    args.rewardGold !== undefined ? `${args.rewardGold} gold` : "",
    args.rewardReputation !== undefined ? `${args.rewardReputation} reputation` : "",
  ].filter(Boolean);
  return `${args.actorName} reached the ${args.questStatus.toLowerCase()} stage of "${args.questName}" in ${args.regionName}. ${
    rewardBits.length > 0 ? `Public settlement notes list ${rewardBits.join(" and ")} as the payout.` : "No extra payout detail is visible in the public record."
  }`;
}

function buildDungeonChronicle(args: {
  language: "zh-CN" | "en-US";
  actorName: string;
  regionName: string;
  dungeonName: string;
  runStatus: string;
  rating: string;
  rewardGold?: number;
  rewardItems: string[];
  materialDropsCount: number;
}) {
  if (args.language === "zh-CN") {
    const rewardParts = [
      args.rewardGold !== undefined ? `${args.rewardGold} 金` : "",
      args.rewardItems.length > 0 ? `${args.rewardItems.length} 件战利品` : "",
      args.materialDropsCount > 0 ? `${args.materialDropsCount} 组材料` : "",
    ].filter(Boolean);
    return `${args.actorName} 的 ${args.dungeonName} 行动已经在公开渠道中结算为 ${args.runStatus}${
      args.rating && args.rating !== "-" ? `，评级 ${args.rating}` : ""
    }。${rewardParts.length > 0 ? `目前能确认的收获包括 ${rewardParts.join("、")}。` : "目前外部只能看到本次远征的结局，未公开更多掉落细节。"}${
      args.regionName ? ` 这条公告被归档在 ${args.regionName} 视角下。` : ""
    }`;
  }

  const rewardParts = [
    args.rewardGold !== undefined ? `${args.rewardGold} gold` : "",
    args.rewardItems.length > 0 ? `${args.rewardItems.length} loot entries` : "",
    args.materialDropsCount > 0 ? `${args.materialDropsCount} material bundles` : "",
  ].filter(Boolean);
  return `${args.actorName}'s ${args.dungeonName} run has settled publicly as ${args.runStatus}${
    args.rating && args.rating !== "-" ? ` with a ${args.rating} rating` : ""
  }. ${
    rewardParts.length > 0
      ? `Confirmed rewards include ${rewardParts.join(", ")}.`
      : "Outside observers can see the result, but not a fuller breakdown of spoils."
  }`;
}

function buildTravelChronicle(args: {
  language: "zh-CN" | "en-US";
  actorName: string;
  fromRegionName: string;
  toRegionName: string;
  travelCostGold?: number;
}) {
  if (args.language === "zh-CN") {
    return `${args.actorName} 已从 ${args.fromRegionName} 转移到 ${args.toRegionName}。${
      args.travelCostGold !== undefined ? `这次移动公开记为 ${args.travelCostGold} 金旅行成本。` : "公开记录没有给出本次移动的额外成本。"
    }`;
  }

  return `${args.actorName} moved from ${args.fromRegionName} to ${args.toRegionName}. ${
    args.travelCostGold !== undefined
      ? `The public ledger records a travel cost of ${args.travelCostGold} gold.`
      : "No extra travel cost detail is exposed publicly."
  }`;
}

function localizeRunStatus(value: string, language: "zh-CN" | "en-US") {
  const normalized = value.trim().toLowerCase();
  if (language === "zh-CN") {
    if (normalized === "cleared") return "已通关";
    if (normalized === "failed") return "失败";
    if (normalized === "abandoned") return "中止";
    if (normalized === "unknown") return "未知";
  } else {
    if (normalized === "cleared") return "Cleared";
    if (normalized === "failed") return "Failed";
    if (normalized === "abandoned") return "Abandoned";
    if (normalized === "unknown") return "Unknown";
  }

  return value || copyUnknown(language);
}

function localizeVisibility(value: string, language: "zh-CN" | "en-US") {
  const normalized = value.trim().toLowerCase();
  if (normalized === "public") {
    return language === "zh-CN" ? "公开抄本" : "Public record";
  }
  return value || copyUnknown(language);
}
