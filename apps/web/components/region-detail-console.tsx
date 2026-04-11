"use client";

import Link from "next/link";

import type { PublicWorldState, Region, RegionDetail, WorldEvent } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  formatMetric,
  formatRelativeTime,
  getRegionAtlasDossier,
  localizeEventSummary,
  localizeRegionDescription,
  localizeRegionName,
  localizeRegionType,
  uiText,
} from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type RegionDetailConsoleProps = {
  worldState: PublicWorldState;
  regionDetail: RegionDetail;
  regions: Region[];
  events: WorldEvent[];
};

export default function RegionDetailConsole({
  worldState,
  regionDetail,
  regions,
  events,
}: RegionDetailConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const common = uiText[language].common;
  const copy = uiText[language].regions;
  const homeCopy = uiText[language].home;
  const pulse = worldState.regions.find((region) => region.region_id === regionDetail.region.region_id);
  const localEvents = events
    .filter((event) => event.region_id === regionDetail.region.region_id)
    .sort((left, right) => new Date(right.occurred_at).getTime() - new Date(left.occurred_at).getTime());
  const atlas = getRegionAtlasDossier(regionDetail.region.region_id, language);
  const localizedRegionDescription = localizeRegionDescription(regionDetail.description, language);
  const atlasFallbackLead =
    language === "zh-CN" ? "该地点暂时没有公开档案。" : "No public dossier is available for this place yet.";
  const storyLeadCopy =
    atlas.shortIntro === atlasFallbackLead && localizedRegionDescription
      ? localizedRegionDescription
      : atlas.shortIntro;

  const layoutCopy =
    language === "zh-CN"
      ? {
          railTitle: "地点索引",
          railNote: "点击左侧地点名，直接切换到对应区域档案。",
          dossierTitle: "区域档案",
          dossierNote: "这里只保留地点身份、环境摘要与成长线索，让详情页回到可阅读的世界资料页。",
          environmentTitle: "环境与定位",
          npcTitle: "关键 NPC",
          facilityTitle: "主要设施",
          materialTitle: "主要产出",
          growthTitle: "成长用途",
          challengeTitle: "挑战与玩法",
          linkedTitle: "关联地点",
          signatureMaterial: "代表材料",
          regionLogTitle: "地区日志",
          regionLogNote: "这里单独记录这片区域最近公开发生的事情。",
        }
      : {
          railTitle: "Place Index",
          railNote: "Pick a place from the left rail to switch to that dossier.",
          dossierTitle: "Region Dossier",
          dossierNote: "Keep the detail page focused on place identity, environment, and progression clues.",
          environmentTitle: "Environment and role",
          npcTitle: "Key NPCs",
          facilityTitle: "Primary facilities",
          materialTitle: "Primary outputs",
          growthTitle: "Growth use",
          challengeTitle: "Challenges and gameplay",
          linkedTitle: "Linked place",
          signatureMaterial: "Signature material",
          regionLogTitle: "Region Log",
          regionLogNote: "Recent public happenings for this place are listed here as a separate archive section.",
        };

  const localizedRegionNameText = localizeRegionName(
    regionDetail.region.region_id,
    regionDetail.region.name,
    language,
  );

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="regions"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={copy.eyebrow}
        title={copy.title}
        intro={copy.intro}
      />

      <section className="detail-stack region-detail-layout">
        <section className="pixel-panel detail-panel region-detail-panel">
          <div className="region-detail-shell">
            <aside className="pixel-panel region-nav-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{copy.atlasTitle}</p>
                  <h2>{layoutCopy.railTitle}</h2>
                </div>
              </div>

              <p className="section-note">{layoutCopy.railNote}</p>

              <nav className="region-nav-list" aria-label={layoutCopy.railTitle}>
                {regions.map((region) => {
                  const active = region.region_id === regionDetail.region.region_id;
                  const regionPulse = worldState.regions.find((item) => item.region_id === region.region_id);
                  const regionAtlas = getRegionAtlasDossier(region.region_id, language);

                  return (
                    <Link
                      key={region.region_id}
                      href={`/regions/${region.region_id}`}
                      className={`region-nav-link${active ? " active" : ""}`}
                      aria-current={active ? "page" : undefined}
                      prefetch={false}
                    >
                      <strong>{localizeRegionName(region.region_id, region.name, language)}</strong>
                      <span>{localizeRegionType(region.type, language)}</span>
                      <small>
                        {regionAtlas.terrainBand} / {homeCopy.riskTier}: {regionAtlas.riskTier} / {common.activeNow}:{" "}
                        {formatMetric(regionPulse?.population ?? 0, language, "")}
                      </small>
                    </Link>
                  );
                })}
              </nav>
            </aside>

            <div className="region-detail-main">
              <section className="region-hero-block">
                <div className="region-story-copy">
                  <p className="eyebrow">{layoutCopy.dossierTitle}</p>
                  <h2 className="region-story-title">{localizedRegionNameText}</h2>
                  <p className="region-story-lead">{storyLeadCopy}</p>

                  <div className="region-badges">
                    <span>{localizeRegionType(regionDetail.region.type, language)}</span>
                    <span>
                      {common.travelCost} {formatMetric(regionDetail.region.travel_cost_gold, language, "")}
                    </span>
                    <span>
                      {homeCopy.terrainBand} {atlas.terrainBand}
                    </span>
                  </div>
                </div>

                <div className="region-stats region-stats-compact">
                  <article>
                    <span>{common.activeNow}</span>
                    <strong>{formatMetric(pulse?.population ?? 0, language, "")}</strong>
                  </article>
                  <article>
                    <span>{common.recentEvents}</span>
                    <strong>{formatMetric(pulse?.recent_event_count ?? localEvents.length, language, "")}</strong>
                  </article>
                  <article>
                    <span>{homeCopy.riskTier}</span>
                    <strong>{atlas.riskTier}</strong>
                  </article>
                  <article>
                    <span>{layoutCopy.signatureMaterial}</span>
                    <strong>{atlas.signatureMaterial}</strong>
                  </article>
                </div>
              </section>

              <section className="region-dossier-grid">
                <article className="region-dossier-card region-dossier-card-accent">
                  <p className="region-story-label">{layoutCopy.environmentTitle}</p>
                  <h3>{atlas.primaryActivity}</h3>
                  <p>{localizedRegionDescription || atlas.shortIntro}</p>
                </article>

                <article className="region-dossier-card">
                  <p className="region-story-label">{layoutCopy.challengeTitle}</p>
                  <h3>{regionDetail.encounter_summary?.activity_type ?? atlas.primaryActivity}</h3>
                  <p>{regionDetail.encounter_summary?.summary ?? atlas.shortIntro}</p>
                  {regionDetail.encounter_summary?.highlights?.length ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {regionDetail.encounter_summary.highlights.map((item) => (
                        <span key={item}>{item}</span>
                      ))}
                    </div>
                  ) : null}
                </article>

                <article className="region-dossier-card">
                  <p className="region-story-label">{layoutCopy.npcTitle}</p>
                  {atlas.notableNpcs.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.notableNpcs.map((npc) => (
                        <span key={npc}>{npc}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}
                </article>

                <article className="region-dossier-card">
                  <p className="region-story-label">{layoutCopy.facilityTitle}</p>
                  {atlas.facilities.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.facilities.map((facility) => (
                        <span key={facility}>{facility}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}
                </article>

                <article className="region-dossier-card">
                  <p className="region-story-label">{layoutCopy.materialTitle}</p>
                  {atlas.materials.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.materials.map((material) => (
                        <span key={material}>{material}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}
                </article>

                <article className="region-dossier-card">
                  <p className="region-story-label">{layoutCopy.growthTitle}</p>
                  {atlas.growthUses.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.growthUses.map((item) => (
                        <span key={item}>{item}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}

                  {atlas.linkedRegionId ? (
                    <p className="region-linkout">
                      <span>{layoutCopy.linkedTitle}</span>
                      <Link className="inline-link" href={`/regions/${atlas.linkedRegionId}`} prefetch={false}>
                        {localizeRegionName(atlas.linkedRegionId, atlas.linkedRegionId, language)}
                      </Link>
                    </p>
                  ) : null}
                </article>
              </section>
            </div>
          </div>
        </section>

        <section className="pixel-panel detail-panel region-log-panel" id="region-log">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.recentSignals}</p>
              <h2>{layoutCopy.regionLogTitle}</h2>
            </div>
          </div>

          <p className="section-note">{layoutCopy.regionLogNote}</p>

          {localEvents.length > 0 ? (
            <div className="log-list region-log-list">
              {localEvents.map((event) => (
                <article key={event.event_id} className="log-entry social-feed-entry">
                  <div className="log-copy">
                    <p className="log-summary">{localizeEventSummary(event.summary, language)}</p>
                    <p className="log-meta">
                      {event.actor_character_id ? (
                        <Link className="inline-link" href={`/bots/${event.actor_character_id}`} prefetch={false}>
                          {event.actor_name ?? common.unknownActor}
                        </Link>
                      ) : (
                        <span>{event.actor_name ?? common.unknownActor}</span>
                      )}
                      <span>{formatRelativeTime(event.occurred_at, language)}</span>
                    </p>
                  </div>
                </article>
              ))}
            </div>
          ) : (
            <p className="empty-state">{copy.noRegionEvents}</p>
          )}
        </section>
      </section>
    </main>
  );
}
