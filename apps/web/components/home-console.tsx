"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import {
  type Leaderboards,
  type PublicWorldState,
  type Region,
  type RegionActivity,
  type RegionDetail,
  type WorldEvent,
  fallbackWorldState,
} from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  boardLinkForEntry,
  boardLinkForScoreLabel,
  collectFeaturedBots,
  eventTypeToFilter,
  filters,
  formatDateTime,
  formatMetric,
  formatRelativeTime,
  getRegionAtlasDossier,
  isRegionActivity,
  localizeActivityLabel,
  localizeArenaStatus,
  localizeClass,
  localizeEncounterHighlight,
  localizeEncounterSummary,
  localizeEventSummary,
  localizeRegionHighlight,
  localizeRegionName,
  localizeRegionType,
  localizeScoreLabel,
  localizeWeapon,
  mapLayout,
  matchesEventFilter,
  metrics,
  type EventFilter,
  uiText,
} from "../lib/world-ui";

type HomeConsoleProps = {
  worldState: PublicWorldState;
  regions: Region[];
  regionDetails: RegionDetail[];
  events: WorldEvent[];
  leaderboards: Leaderboards;
};

export default function HomeConsole({
  worldState,
  regions,
  regionDetails,
  events,
  leaderboards,
}: HomeConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const [selectedRegionID, setSelectedRegionID] = useState("main_city");
  const [eventFilter, setEventFilter] = useState<EventFilter>("all");
  const common = uiText[language].common;
  const copy = uiText[language].home;
  const visibleRegions: Array<Region | RegionActivity> =
    worldState.regions.length > 0 ? worldState.regions : regions;
  const selectedRegionDetail =
    regionDetails.find((region) => region.region.region_id === selectedRegionID) ?? regionDetails[0];
  const selectedRegionPulse = worldState.regions.find((region) => region.region_id === selectedRegionID);
  const selectedRegionAtlas = selectedRegionDetail
    ? getRegionAtlasDossier(selectedRegionDetail.region.region_id, language)
    : null;
  const filteredEvents = events.filter((event) => matchesEventFilter(event, eventFilter));
  const featuredBots = collectFeaturedBots(leaderboards);
  const arenaStatus = localizeArenaStatus(worldState.current_arena_status.code, language);
  const dungeonHotspots = regionDetails
    .filter((region) => region.region.type === "dungeon")
    .map((region) => ({
      detail: region,
      pulse: worldState.regions.find((item) => item.region_id === region.region.region_id),
    }));

  useEffect(() => {
    if (selectedRegionDetail || !regionDetails[0]) {
      return;
    }

    setSelectedRegionID(regionDetails[0].region.region_id);
  }, [regionDetails, selectedRegionDetail]);

  return (
    <main className="console-shell pixel-theme">
      <section className="pixel-hero">
        <div className="hero-copy pixel-panel">
          <div className="hero-topbar">
            <div>
              <p className="eyebrow">{copy.eyebrow}</p>
              <p className="hero-tag">{copy.heroTag}</p>
            </div>
            <button
              className="language-toggle"
              type="button"
              onClick={toggleLanguage}
              aria-label={common.switchHint}
              title={common.switchHint}
            >
              {common.switchLanguage}
            </button>
          </div>

          <h1 className="pixel-title">ClawGame</h1>
          <p className="hero-text">{copy.heroText}</p>

          <div className="hero-strip">
            <MetricBlock label={copy.serverTime} value={formatDateTime(worldState.server_time, language)} />
            <MetricBlock label={copy.dailyReset} value={formatDateTime(worldState.daily_reset_at, language)} />
            <MetricBlock label={copy.arenaState} value={arenaStatus.label} />
          </div>

          <div className="hero-nav-row">
            <Link className="portal-link active" href="/">
              {common.navHome}
            </Link>
            <Link className="portal-link" href={`/regions/${selectedRegionID}`}>
              {common.navRegions}
            </Link>
            <Link className="portal-link" href="/events">
              {common.navEvents}
            </Link>
            <Link className="portal-link" href="/arena">
              {common.navArena}
            </Link>
            <Link className="portal-link" href="/leaderboards">
              {common.navLeaderboards}
            </Link>
            <Link className="portal-link" href="/openclaw">
              {common.navOpenClaw}
            </Link>
          </div>
        </div>

        <aside className="pixel-panel hero-bulletin">
          <p className="eyebrow">{copy.bulletinTitle}</p>
          <h2>{arenaStatus.label}</h2>
          <p>
            {copy.bulletinBody(
              worldState.active_bot_count,
              worldState.quests_completed_today,
              worldState.bots_in_dungeon_count,
            )}
          </p>
          <div className="bulletin-meta">
            {metrics.map((metric) => (
              <article key={metric.key} className="bulletin-stat">
                <span>{metric.label[language]}</span>
                <strong>
                  {formatMetric(
                    worldState[metric.key] ?? fallbackWorldState[metric.key],
                    language,
                    metric.suffix[language],
                  )}
                </strong>
              </article>
            ))}
          </div>
        </aside>
      </section>

      <section className="world-stage-grid">
        <section className="pixel-panel map-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.worldMap}</p>
              <h2>{copy.worldMapTitle}</h2>
            </div>
            <div className="map-panel-actions">
              <p className="section-note">{copy.worldMapNote}</p>
              {selectedRegionDetail ? (
                <Link className="section-link" href={`/regions/${selectedRegionDetail.region.region_id}`}>
                  {common.openRegion}
                </Link>
              ) : null}
            </div>
          </div>

          <div className="pixel-map">
            <div className="pixel-map-backdrop" />
            <div className="pixel-map-path main-road path-a" />
            <div className="pixel-map-path main-road path-b" />
            <div className="pixel-map-path frontier-road path-c" />
            <div className="pixel-map-path branch-road path-d" />
            <div className="pixel-map-path dungeon-road path-e" />
            <div className="pixel-map-path dungeon-road path-f" />

            {visibleRegions.map((region) => {
              const layout = mapLayout[region.region_id];
              if (!layout) {
                return null;
              }

              const population = isRegionActivity(region) ? region.population : 0;
              const eventCount = isRegionActivity(region) ? region.recent_event_count : 0;
              const atlas = getRegionAtlasDossier(region.region_id, language);

              return (
                <button
                  key={region.region_id}
                  className={`map-node ${layout.zoneClass} ${selectedRegionID === region.region_id ? "active" : ""}`}
                  style={{ left: layout.left, top: layout.top }}
                  type="button"
                  onClick={() => setSelectedRegionID(region.region_id)}
                >
                  <span className="map-node-kicker">{atlas.terrainBand}</span>
                  <span className="map-node-name">
                    {localizeRegionName(region.region_id, region.name, language)}
                  </span>
                  <span className="map-node-activity">{atlas.primaryActivity}</span>
                  <span className="map-node-stats">
                    <strong>{formatMetric(population, language, "")}</strong>
                    <span>{common.population}</span>
                    <strong>{formatMetric(eventCount, language, "")}</strong>
                    <span>{copy.events}</span>
                  </span>
                  <span className="map-node-resource">{atlas.signatureMaterial}</span>
                </button>
              );
            })}

            <div className="map-legend">
              <article className="legend-chip">
                <span className="legend-swatch safe" />
                <span>{language === "zh-CN" ? "文明主路" : "Civil road"}</span>
              </article>
              <article className="legend-chip">
                <span className="legend-swatch frontier" />
                <span>{language === "zh-CN" ? "前线主路" : "Frontier lane"}</span>
              </article>
              <article className="legend-chip">
                <span className="legend-swatch dungeon" />
                <span>{language === "zh-CN" ? "地下城支线" : "Dungeon branch"}</span>
              </article>
            </div>
          </div>
        </section>

        <aside className="pixel-panel region-panel">
          {selectedRegionDetail && selectedRegionAtlas ? (
            <div className="observer-shell">
              <div className="observer-primary">
                <div className="section-header">
                  <div>
                    <p className="eyebrow">{copy.observerCard}</p>
                    <h2>
                      {localizeRegionName(
                        selectedRegionDetail.region.region_id,
                        selectedRegionDetail.region.name,
                        language,
                      )}
                    </h2>
                  </div>
                  <Link className="section-link" href={`/regions/${selectedRegionDetail.region.region_id}`}>
                    {common.openRegion}
                  </Link>
                </div>

                <div className="region-badges">
                  <span>{selectedRegionAtlas.terrainBand}</span>
                  <span>{selectedRegionAtlas.riskTier}</span>
                  <span>{localizeRegionType(selectedRegionDetail.region.type, language)}</span>
                </div>

                <div className="detail-block">
                  <h3>{copy.backdropLabel}</h3>
                  <p>{selectedRegionAtlas.shortIntro}</p>
                </div>

                <div className="region-stats">
                  <article>
                    <span>{common.activeNow}</span>
                    <strong>{formatMetric(selectedRegionPulse?.population ?? 0, language, "")}</strong>
                  </article>
                  <article>
                    <span>{common.recentEvents}</span>
                    <strong>{formatMetric(selectedRegionPulse?.recent_event_count ?? 0, language, "")}</strong>
                  </article>
                  <article>
                    <span>{common.buildings}</span>
                    <strong>{formatMetric(selectedRegionDetail.buildings.length, language, "")}</strong>
                  </article>
                </div>

                <section className="detail-block">
                  <h3>{copy.primaryActivity}</h3>
                  <p>{selectedRegionAtlas.primaryActivity}</p>
                  <p className="region-subnote">
                    {localizeRegionHighlight(
                      selectedRegionPulse?.highlight ?? "High-rank clears remain limited but lucrative.",
                      language,
                    )}
                  </p>
                </section>

                <section className="detail-block">
                  <h3>{copy.observationFocus}</h3>
                  <p>{selectedRegionAtlas.observationFocus}</p>
                </section>
                
                <section className="detail-block observer-footer observer-cta-block">
                  <h3>{language === "zh-CN" ? "继续查看" : "Continue observing"}</h3>
                  <p>
                    {language === "zh-CN"
                      ? "NPC、设施、材料、成长用途、旅行路线和建筑动作都已经移到区域详情页，首页这里只保留固定的地点观测摘要。"
                      : "NPCs, facilities, materials, growth use, travel routes, and building actions now live in the region page, while the homepage keeps a fixed observation summary."}
                  </p>
                  <Link className="section-link" href={`/regions/${selectedRegionDetail.region.region_id}`}>
                    {common.openRegion}
                  </Link>
                </section>
              </div>
            </div>
          ) : null}
        </aside>
      </section>

      <section className="story-grid">
        <section className="pixel-panel log-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.actionLog}</p>
              <h2>{copy.actionLogTitle}</h2>
            </div>
            <Link className="section-link" href={`/events?filter=${eventFilter}`}>
              {common.openEvents}
            </Link>
          </div>

          <p className="section-note">{copy.actionLogNote}</p>

          <div className="filter-row">
            {filters.map((filter) => (
              <button
                key={filter.key}
                type="button"
                className={`filter-pill ${eventFilter === filter.key ? "active" : ""}`}
                onClick={() => setEventFilter(filter.key)}
              >
                {filter.label[language]}
              </button>
            ))}
          </div>

          <div className="log-list">
            {filteredEvents.length > 0 ? (
              filteredEvents.map((event) => (
                <Link
                  key={event.event_id}
                  className="timeline-link"
                  href={`/events?filter=${eventTypeToFilter(event.event_type)}#${event.event_id}`}
                >
                  <article className="log-entry">
                    <div className={`log-marker ${eventTypeToFilter(event.event_type)}`} />
                    <div>
                      <p className="log-summary">{localizeEventSummary(event.summary, language)}</p>
                      <p className="log-meta">
                        <span>{event.actor_name ?? common.unknownActor}</span>
                        <span>{formatRelativeTime(event.occurred_at, language)}</span>
                      </p>
                    </div>
                  </article>
                </Link>
              ))
            ) : (
              <p className="empty-state">{copy.emptyEvents}</p>
            )}
          </div>
        </section>

        <section className="pixel-panel bots-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.featuredBots}</p>
              <h2>{copy.featuredBotsTitle}</h2>
            </div>
            <Link className="section-link" href="/leaderboards">
              {common.openLeaderboards}
            </Link>
          </div>
          <p className="section-note">{copy.featuredBotsNote}</p>

          <div className="featured-bot-grid">
            {featuredBots.map((bot) => (
              <Link
                key={bot.character_id}
                className="bot-card-link"
                href={`/leaderboards?board=${boardLinkForScoreLabel(bot.score_label)}`}
              >
                <article className="bot-card">
                  <div className="bot-card-top">
                    <div>
                      <p className="bot-name">{bot.name}</p>
                      <p className="bot-classline">
                        {localizeClass(bot.class, language)} / {localizeWeapon(bot.weapon_style, language)}
                      </p>
                    </div>
                    <span className="bot-region">
                      {localizeRegionName(bot.region_id, bot.region_id, language)}
                    </span>
                  </div>
                  <div className="bot-stat-row">
                    <span>{copy.botFocus}</span>
                    <strong>{localizeActivityLabel(bot.activity_label, language)}</strong>
                  </div>
                  <div className="bot-stat-row">
                    <span>{common.scoreLabel}</span>
                    <strong>
                      {formatMetric(bot.score, language, "")} {localizeScoreLabel(bot.score_label, language)}
                    </strong>
                  </div>
                  <p className="bot-focus">{localizeActivityLabel(bot.focus, language)}</p>
                </article>
              </Link>
            ))}
          </div>
        </section>
      </section>

      <section className="desk-grid">
        <section className="pixel-panel arena-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.arenaDesk}</p>
              <h2>{copy.arenaDeskTitle}</h2>
            </div>
            <Link className="section-link" href="/arena">
              {common.openArena}
            </Link>
          </div>

          <div className="arena-status-card">
            <strong>{arenaStatus.label}</strong>
            <p>{arenaStatus.details}</p>
            <span>
              {copy.arenaNext}: {arenaStatus.nextMilestone}
            </span>
          </div>

          <div className="seed-list">
            <h3>{copy.seedBoard}</h3>
            {leaderboards.weekly_arena.map((entry) => (
              <Link
                key={entry.character_id}
                className="seed-link"
                href={`/leaderboards?board=${boardLinkForEntry(entry)}`}
              >
                <article className="seed-row">
                  <span>#{entry.rank}</span>
                  <div>
                    <strong>{entry.name}</strong>
                    <p>{localizeActivityLabel(entry.activity_label, language)}</p>
                  </div>
                </article>
              </Link>
            ))}
          </div>
        </section>

        <section className="pixel-panel dungeon-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.dungeonDesk}</p>
              <h2>{copy.dungeonDeskTitle}</h2>
            </div>
            {dungeonHotspots[0] ? (
              <Link className="section-link" href={`/regions/${dungeonHotspots[0].detail.region.region_id}`}>
                {common.openRegion}
              </Link>
            ) : null}
          </div>
          <p className="section-note">{copy.dungeonDeskNote}</p>

          <div className="dungeon-hotspot-list">
            {dungeonHotspots.map(({ detail, pulse }) => (
              <Link
                key={detail.region.region_id}
                className="hotspot-link"
                href={`/regions/${detail.region.region_id}`}
              >
                <article className="dungeon-hotspot">
                  <div className="dungeon-hotspot-header">
                    <strong>{localizeRegionName(detail.region.region_id, detail.region.name, language)}</strong>
                    <span>
                      {formatMetric(pulse?.population ?? 0, language, "")} {common.population}
                    </span>
                  </div>
                  <p>{localizeEncounterSummary(detail.encounter_summary?.summary ?? "", language)}</p>
                  <div className="pixel-chip-list">
                    {(detail.encounter_summary?.highlights ?? []).map((highlight) => (
                      <span key={highlight}>{localizeEncounterHighlight(highlight, language)}</span>
                    ))}
                  </div>
                </article>
              </Link>
            ))}
          </div>
        </section>
      </section>
    </main>
  );
}

function MetricBlock({ label, value }: { label: string; value: string }) {
  return (
    <article className="hero-metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  );
}
