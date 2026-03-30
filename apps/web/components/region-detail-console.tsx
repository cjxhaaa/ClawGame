"use client";

import Link from "next/link";
import { useState } from "react";

import type { PublicWorldState, Region, RegionDetail, WorldEvent } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  eventTypeToFilter,
  formatMetric,
  formatRelativeTime,
  getRegionAtlasDossier,
  localizeActionName,
  localizeEventSummary,
  localizeRegionDescription,
  localizeRegionName,
  localizeRegionType,
  localizeBuildingName,
  localizeRank,
  metricSuffix,
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
  const [openPanel, setOpenPanel] = useState<"dossier" | "actions" | null>(null);
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
          storyLead: "区域叙述与当前动态",
          storySummary: "先理解这个地点在世界中的职责，再看当前 Bot 正在这里推进什么。",
          systemsTitle: "世界中的功能",
          systemsNote: "把这个区域的 NPC、设施、材料与成长用途放在同一张观察面板里。",
          npcTitle: "关键 NPC",
          facilityTitle: "设施功能",
          materialTitle: "主要产出",
          growthTitle: "成长用途",
          operationsTitle: "行动与线路",
          operationsNote: "这里展示 Bot 在此地可执行的公共动作，以及离开这片区域后的去向。",
          buildingActionsTitle: "建筑与动作",
          routeMatrixTitle: "旅行网络",
          signalNote: "最近的公开事件能帮助你判断这片区域现在究竟热不热、危险高不高。",
          atlasRailTitle: "继续观察其它地点",
          atlasRailNote: "从这里横向切换到其它地点，像翻一本像素地图册，而不是跳进一个独立侧栏。",
          dossierLabel: "地点档案",
          signatureMaterial: "代表材料",
          signalStreamTitle: "动态情报流",
          signalStreamNote: "这里只保留真实事件，并让它持续滚动，像一条正在刷新的区域情报带。",
          openDossier: "打开完整地点档案",
          openOperations: "打开建筑与动作",
          closePanel: "关闭面板",
          recentEventsTitle: "滚动情报流",
          recentEventsNote: "按时间顺序持续滚动最近发生的公开动作；当区域数据较少时自动循环补齐，悬浮时暂停。",
          openRoutes: "旅行网络",
          feedLane: "动态情报流",
          routeSidebarTitle: "旅行网络",
          routeSidebarNote: "鼠标移动到这里时展开路线卡片，直接导航到下一个地点。",
          noTravelHint: "暂无额外旅行路线信息。",
          browseWorld: "继续观察其它地点",
          travelDockHint: "悬浮路线卡",
          routeDeckNote: "每张卡都代表一个下一跳地点，包含风险、用途与旅行代价。",
          routeTraffic: "本地活跃",
        }
      : {
          storyLead: "Region story and current pressure",
          storySummary: "Understand what this place does in the world first, then read what bots are actively pushing here.",
          systemsTitle: "What this region provides",
          systemsNote: "NPCs, facilities, materials, and growth value are grouped into one readable observer panel.",
          npcTitle: "Key NPCs",
          facilityTitle: "Facilities",
          materialTitle: "Primary outputs",
          growthTitle: "Growth use",
          operationsTitle: "Actions and routes",
          operationsNote: "This panel shows what public actions bots can take here and where they branch next.",
          buildingActionsTitle: "Buildings and actions",
          routeMatrixTitle: "Travel network",
          signalNote: "Recent public events make it easier to judge whether this region is hot, risky, or simply busy.",
          atlasRailTitle: "Continue to other places",
          atlasRailNote: "Move sideways through the world like a pixel field guide instead of jumping to a separate sidebar.",
          dossierLabel: "Place dossier",
          signatureMaterial: "Signature material",
          signalStreamTitle: "Signal stream",
          signalStreamNote: "This lane is reserved for real event rows only, and keeps moving like a live regional intel tape.",
          openDossier: "Open full dossier",
          openOperations: "Open buildings and actions",
          closePanel: "Close panel",
          recentEventsTitle: "Rolling intel feed",
          recentEventsNote: "Recent public actions scroll in time order; when local data is short, the feed loops to stay alive and pauses on hover.",
          openRoutes: "Travel network",
          feedLane: "Signal stream",
          routeSidebarTitle: "Travel network",
          routeSidebarNote: "Hover here to expand the route card and jump to the next place.",
          noTravelHint: "No extra travel routes are available yet.",
          browseWorld: "Continue to other places",
          travelDockHint: "Hover route card",
          routeDeckNote: "Each card represents a destination hop with risk, purpose, and travel cost.",
          routeTraffic: "Local activity",
        };
  const latestEvent = localEvents[0];
  const eventFilter = latestEvent ? eventTypeToFilter(latestEvent.event_type) : "all";
  const eventEntries = localEvents.map((event) => ({
    key: `event-${event.event_id}`,
    href: `/events/${event.event_id}?filter=${eventTypeToFilter(event.event_type)}&region=${regionDetail.region.region_id}&focus=${event.event_id}`,
    summary: localizeEventSummary(event.summary, language),
    actorCharacterID: event.actor_character_id,
    actor: event.actor_name ?? common.unknownActor,
    time: formatRelativeTime(event.occurred_at, language),
    marker: eventTypeToFilter(event.event_type),
  }));
  const loopSeed =
    eventEntries.length > 0
      ? Array.from({ length: Math.max(4, eventEntries.length) }, (_, index) => {
          const event = eventEntries[index % eventEntries.length];

          return {
            ...event,
            loopKey: `${event.key}-seed-${index}`,
          };
        })
      : [];
  const marqueeEvents = loopSeed.length > 0 ? [...loopSeed, ...loopSeed] : [];
  const travelDestinations = regionDetail.travel_options.map((route) => {
    const destinationAtlas = getRegionAtlasDossier(route.region_id, language);
    const destinationPulse = worldState.regions.find((item) => item.region_id === route.region_id);

    return {
      ...route,
      atlas: destinationAtlas,
      population: destinationPulse?.population ?? 0,
    };
  });

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
        <section className="pixel-panel detail-panel region-story-panel">
          <div className="region-story-grid">
            <div className="region-story-main">
              <div className="region-story-copy">
                <p className="eyebrow">{layoutCopy.storyLead}</p>
                <h2 className="region-story-title">
                  {localizeRegionName(regionDetail.region.region_id, regionDetail.region.name, language)}
                </h2>
                <p className="region-story-lead">{storyLeadCopy}</p>

                <div className="region-badges">
                  <span>{localizeRegionType(regionDetail.region.type, language)}</span>
                  <span>
                    {common.minRank} {localizeRank(regionDetail.region.min_rank, language)}
                  </span>
                  <span>
                    {common.travelCost}{" "}
                    {formatMetric(regionDetail.region.travel_cost_gold, language, metricSuffix(language))}
                  </span>
                </div>
              </div>

              <div className="region-stats region-stats-compact">
                <article>
                  <span>{common.activeNow}</span>
                  <strong>{formatMetric(pulse?.population ?? 0, language, "")}</strong>
                </article>
                <article>
                  <span>{common.buildings}</span>
                  <strong>{formatMetric(regionDetail.buildings.length, language, "")}</strong>
                </article>
                <article>
                  <span>{homeCopy.terrainBand}</span>
                  <strong>{atlas.terrainBand}</strong>
                </article>
                <article>
                  <span>{homeCopy.riskTier}</span>
                  <strong>{atlas.riskTier}</strong>
                </article>
              </div>
            </div>

            <aside className="pixel-panel region-live-panel">
                <div className="section-header">
                  <div>
                    <p className="eyebrow">{layoutCopy.feedLane}</p>
                    <h2>{layoutCopy.recentEventsTitle}</h2>
                  </div>
                <Link
                  className="section-link"
                  href={
                    regionDetail.region.region_id
                      ? `/events${eventFilter === "all" ? "" : `?filter=${eventFilter}`}${
                          eventFilter === "all" ? "?" : "&"
                        }region=${regionDetail.region.region_id}`
                      : `/events${eventFilter === "all" ? "" : `?filter=${eventFilter}`}`
                  }
                >
                  {uiText[language].common.openEvents}
                </Link>
              </div>

              <p className="section-note">{layoutCopy.signalStreamNote}</p>
              <p className="region-subnote">{layoutCopy.recentEventsNote}</p>

              {eventEntries.length > 0 ? (
                  <div className="region-event-scroll">
                    <div className="region-event-track animated">
                      {marqueeEvents.map((entry, index) => (
                        <article key={`${entry.loopKey}-${index}`} className="log-entry region-feed-entry">
                          <div className={`log-marker ${entry.marker}`} />
                          <div>
                            <Link className="feed-summary-link" href={entry.href}>
                              {entry.summary}
                            </Link>
                            <p className="log-meta">
                              {entry.actorCharacterID ? (
                                <Link className="inline-link" href={`/bots/${entry.actorCharacterID}`}>
                                  {entry.actor}
                                </Link>
                              ) : (
                                <span>{entry.actor}</span>
                              )}
                              <span>{entry.time}</span>
                            </p>
                          </div>
                        </article>
                      ))}
                    </div>
                  </div>
              ) : (
                <p className="empty-state">{copy.noRegionEvents}</p>
              )}
            </aside>
          </div>

          <div className="region-story-footer">
            <div className="region-hero-actions">
              <button type="button" className="section-link" onClick={() => setOpenPanel("dossier")}>
                {layoutCopy.openDossier}
              </button>
              <button type="button" className="section-link" onClick={() => setOpenPanel("actions")}>
                {layoutCopy.openOperations}
              </button>
            </div>

            <div className="region-travel-hover" tabIndex={0}>
              <div className="region-travel-trigger">
                <span>{layoutCopy.openRoutes}</span>
                <small>{layoutCopy.travelDockHint}</small>
              </div>

              <section className="pixel-panel region-travel-popover">
                <div className="section-header">
                  <div>
                    <p className="eyebrow">{layoutCopy.openRoutes}</p>
                    <h2>{layoutCopy.routeSidebarTitle}</h2>
                  </div>
                </div>

                <p className="section-note">{layoutCopy.routeSidebarNote}</p>

                <p className="region-subnote">{layoutCopy.routeDeckNote}</p>

                {travelDestinations.length > 0 ? (
                  <div className="region-route-deck">
                    {travelDestinations.map((route) => (
                      <Link
                        key={`${regionDetail.region.region_id}-${route.region_id}`}
                        className="route-card-link region-route-card"
                        href={`/regions/${route.region_id}`}
                      >
                        <article>
                          <p className="region-atlas-overline">
                            {route.atlas.terrainBand} / {route.atlas.riskTier}
                          </p>
                          <strong>{localizeRegionName(route.region_id, route.name, language)}</strong>
                          <p className="region-route-story">{route.atlas.primaryActivity}</p>
                          <p className="region-route-summary">{route.atlas.shortIntro}</p>
                          <div className="region-route-meta">
                            <span>
                              {localizeRank(route.requires_rank, language)} /{" "}
                              {formatMetric(route.travel_cost_gold, language, metricSuffix(language))}
                            </span>
                            <span>
                              {layoutCopy.routeTraffic}: {formatMetric(route.population, language, "")}
                            </span>
                          </div>
                        </article>
                      </Link>
                    ))}
                  </div>
                ) : (
                  <p className="empty-state">{layoutCopy.noTravelHint}</p>
                )}
              </section>
            </div>
          </div>
        </section>
      </section>

      {openPanel ? (
        <div className="region-modal-overlay" onClick={() => setOpenPanel(null)}>
          <section
            className={`pixel-panel region-modal-panel ${openPanel === "actions" ? "operations-modal" : ""}`}
            onClick={(event) => event.stopPropagation()}
          >
            <div className="region-modal-header">
              <div>
                <p className="eyebrow">
                  {openPanel === "dossier" ? layoutCopy.dossierLabel : layoutCopy.openOperations}
                </p>
                <h2>{openPanel === "dossier" ? layoutCopy.systemsTitle : layoutCopy.openOperations}</h2>
              </div>
              <button type="button" className="section-link" onClick={() => setOpenPanel(null)}>
                {layoutCopy.closePanel}
              </button>
            </div>

            <p className="section-note">
              {openPanel === "dossier" ? layoutCopy.systemsNote : layoutCopy.operationsNote}
            </p>

            {openPanel === "dossier" ? (
              <div className="region-modal-grid">
                <section className="region-meta-block">
                  <h3>{layoutCopy.npcTitle}</h3>
                  {atlas.notableNpcs.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.notableNpcs.map((npc) => (
                        <span key={npc}>{npc}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}
                </section>

                <section className="region-meta-block">
                  <h3>{layoutCopy.facilityTitle}</h3>
                  {atlas.facilities.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.facilities.map((facility) => (
                        <span key={facility}>{facility}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}
                </section>

                <section className="region-meta-block">
                  <h3>{layoutCopy.materialTitle}</h3>
                  {atlas.materials.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.materials.map((material) => (
                        <span key={material}>{material}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}
                </section>

                <section className="region-meta-block">
                  <h3>{layoutCopy.growthTitle}</h3>
                  {atlas.growthUses.length > 0 ? (
                    <div className="pixel-chip-list compact-chip-list">
                      {atlas.growthUses.map((item) => (
                        <span key={item}>{item}</span>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noHighlights}</p>
                  )}
                  <p className="region-subnote">
                    {layoutCopy.signatureMaterial}: <strong>{atlas.signatureMaterial}</strong>
                  </p>
                  {atlas.linkedRegionId ? (
                    <p className="region-linkout">
                      <span>{atlas.linkedRegionLabel ?? common.linkedRegion}</span>
                      <Link className="inline-link" href={`/regions/${atlas.linkedRegionId}`}>
                        {localizeRegionName(atlas.linkedRegionId, atlas.linkedRegionId, language)}
                      </Link>
                    </p>
                  ) : null}
                </section>
              </div>
            ) : (
              <div className="region-modal-grid single-column-grid">
                <section className="region-operations-column">
                  <h3>{layoutCopy.buildingActionsTitle}</h3>
                  {regionDetail.buildings.length > 0 ? (
                    <div className="building-stack">
                      {regionDetail.buildings.map((building) => (
                        <article key={building.building_id} className="building-card modal-building-card">
                          <strong>{localizeBuildingName(building, language)}</strong>
                          <div className="pixel-chip-list compact-chip-list">
                            {building.actions.map((action) => (
                              <span key={`${building.building_id}-${action}`}>
                                {localizeActionName(action, language)}
                              </span>
                            ))}
                          </div>
                        </article>
                      ))}
                    </div>
                  ) : (
                    <p className="empty-state">{common.noBuildings}</p>
                  )}
                </section>
              </div>
            )}
          </section>
        </div>
      ) : null}
    </main>
  );
}
