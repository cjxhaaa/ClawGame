"use client";

import Link from "next/link";
import { useMemo, useState } from "react";

import type { PublicWorldState, Region, RegionDetail, WorldEvent } from "../lib/public-api";
import SiteFrame from "./site-frame";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  filters,
  formatDateTime,
  formatMetric,
  formatRelativeTime,
  getRegionAtlasDossier,
  localizeEventSummary,
  localizeRegionName,
  localizeRegionType,
  matchesEventFilter,
  toEventFilter,
  type Language,
} from "../lib/world-ui";

type RegionsIndexConsoleProps = {
  worldState: PublicWorldState;
  regions: Region[];
  regionDetails: RegionDetail[];
  recentEvents: WorldEvent[];
  initialEventFilter?: string;
};

const copy: Record<
  Language,
  {
    eyebrow: string;
    title: string;
    intro: string;
    all: string;
    withDungeon: string;
    searchPlaceholder: string;
    risk: string;
    stage: string;
    regionType: string;
    activity: string;
    openDetail: string;
    noResults: string;
    chronicle: string;
    chronicleTitle: string;
    chronicleIntro: string;
    chronicleEmpty: string;
  }
> = {
  "zh-CN": {
    eyebrow: "区域索引",
    title: "Regions 目录与地点档案",
    intro: "先像翻 wiki 索引一样找到区域，再进入每个区域的完整档案，查看副本、设施、产出与旅行连接。",
    all: "全部区域",
    withDungeon: "仅看有关联副本",
    searchPlaceholder: "搜索区域名",
    risk: "风险",
    stage: "成长阶段",
    regionType: "类型",
    activity: "当前活跃",
    openDetail: "查看档案",
    noResults: "当前筛选下没有区域。",
    chronicle: "世界档案流",
    chronicleTitle: "并入区域档案的世界日志",
    chronicleIntro: "全局公共事件保留在 Regions 体系里，先看世界正在发生什么，再顺着事件进入对应区域档案。",
    chronicleEmpty: "当前还没有可展示的公开世界事件。",
  },
  "en-US": {
    eyebrow: "Region Index",
    title: "Regions Directory and Place Dossiers",
    intro: "Find a region through a readable world index first, then enter its full dossier for dungeons, facilities, outputs, and travel links.",
    all: "All Regions",
    withDungeon: "Has Dungeon",
    searchPlaceholder: "Search region name",
    risk: "Risk",
    stage: "Stage",
    regionType: "Type",
    activity: "Active Now",
    openDetail: "Open Dossier",
    noResults: "No regions match the current filters.",
    chronicle: "World Chronicle",
    chronicleTitle: "World log merged into the Regions archive flow",
    chronicleIntro:
      "Global public events now live inside the Regions flow, so you can read the world signal first and then move into the matching regional dossier.",
    chronicleEmpty: "There are no public world events to show right now.",
  },
};

export default function RegionsIndexConsole({
  worldState,
  regions,
  regionDetails,
  recentEvents,
  initialEventFilter,
}: RegionsIndexConsoleProps) {
  const { language } = useWorldLanguage();
  const text = copy[language];
  const [query, setQuery] = useState("");
  const [dungeonOnly, setDungeonOnly] = useState(false);
  const [selectedType, setSelectedType] = useState("all");
  const [eventFilter, setEventFilter] = useState(toEventFilter(initialEventFilter));

  const types = useMemo(() => ["all", ...new Set(regions.map((region) => region.type))], [regions]);

  const cards = useMemo(
    () =>
      regionDetails.map((detail) => {
        const atlas = getRegionAtlasDossier(detail.region.region_id, language);
        const pulse = worldState.regions.find((item) => item.region_id === detail.region.region_id);

        return {
          detail,
          atlas,
          pulse,
        };
      }),
    [language, regionDetails, worldState.regions],
  );

  const filtered = cards.filter(({ detail, atlas }) => {
    const regionName = localizeRegionName(detail.region.region_id, detail.region.name, language).toLowerCase();
    const matchesQuery = query.trim() ? regionName.includes(query.trim().toLowerCase()) : true;
    const matchesDungeon = dungeonOnly ? Boolean(detail.linked_dungeon) : true;
    const matchesType = selectedType === "all" ? true : detail.region.type === selectedType;

    return matchesQuery && matchesDungeon && matchesType && atlas;
  });
  const visibleEvents = recentEvents.filter((event) => matchesEventFilter(event, eventFilter));

  return (
    <SiteFrame
      active="regions"
      eyebrow={text.eyebrow}
      title={text.title}
      intro={text.intro}
      stats={[
        { label: text.all, value: formatMetric(regions.length, language, "") },
        {
          label: text.withDungeon,
          value: formatMetric(regionDetails.filter((detail) => Boolean(detail.linked_dungeon)).length, language, ""),
        },
      ]}
    >
      <section className="world-panel">
        <div className="world-section-heading">
          <div>
            <p className="world-section-kicker">Filters</p>
            <h2>{text.title}</h2>
          </div>
        </div>

        <div className="world-filter-bar">
          <input
            className="world-search-input"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={text.searchPlaceholder}
          />

          <select
            className="world-select"
            value={selectedType}
            onChange={(event) => setSelectedType(event.target.value)}
          >
            {types.map((type) => (
              <option key={type} value={type}>
                {type === "all" ? text.all : localizeRegionType(type, language)}
              </option>
            ))}
          </select>

          <button
            type="button"
            className={`world-filter-chip ${dungeonOnly ? "active" : ""}`}
            onClick={() => setDungeonOnly((value) => !value)}
          >
            {text.withDungeon}
          </button>
        </div>
      </section>

      <section className="world-region-card-grid">
        {filtered.length > 0 ? (
          filtered.map(({ detail, atlas, pulse }) => (
            <article key={detail.region.region_id} className="world-panel world-region-card">
              <div className="world-section-heading">
                <div>
                  <p className="world-section-kicker">{text.regionType}</p>
                  <h2>{localizeRegionName(detail.region.region_id, detail.region.name, language)}</h2>
                </div>
                <span className="world-chip">{localizeRegionType(detail.region.type, language)}</span>
              </div>

              <p className="world-region-lead">{atlas.shortIntro}</p>

              <div className="world-mini-facts">
                <article>
                  <span>{text.risk}</span>
                  <strong>{atlas.riskTier}</strong>
                </article>
                <article>
                  <span>{text.stage}</span>
                  <strong>{atlas.terrainBand}</strong>
                </article>
                <article>
                  <span>{text.activity}</span>
                  <strong>{formatMetric(pulse?.population ?? 0, language, "")}</strong>
                </article>
              </div>

              <div className="world-stack-list">
                <article className="world-inline-card static">
                  <strong>Dungeon</strong>
                  <span>{detail.linked_dungeon ?? "No linked dungeon"}</span>
                </article>
                <article className="world-inline-card static">
                  <strong>Facilities</strong>
                  <span>{atlas.facilities.slice(0, 2).join(" · ") || "No public facility data yet."}</span>
                </article>
              </div>

              <Link
                className="world-primary-link"
                href={`/regions/${encodeURIComponent(detail.region.region_id)}`}
              >
                {text.openDetail}
              </Link>
            </article>
          ))
        ) : (
          <article className="world-panel">
            <p className="world-empty">{text.noResults}</p>
          </article>
        )}
      </section>

      <section id="world-chronicle" className="world-panel world-chronicle-panel">
        <div className="world-section-heading">
          <div>
            <p className="world-section-kicker">{text.chronicle}</p>
            <h2>{text.chronicleTitle}</h2>
          </div>
        </div>

        <p className="section-note">{text.chronicleIntro}</p>

        <div className="filter-row">
          {filters.map((item) => (
            <button
              key={item.key}
              type="button"
              className={`filter-pill ${eventFilter === item.key ? "active" : ""}`}
              onClick={() => setEventFilter(item.key)}
            >
              {item.label[language]}
            </button>
          ))}
        </div>

        <div className="log-list world-chronicle-list">
          {visibleEvents.length > 0 ? (
            visibleEvents.map((event) => (
              <article key={event.event_id} className="log-entry expanded-log-entry social-feed-entry">
                <div className={`log-marker ${itemType(event.event_type)}`} />
                <div>
                  <Link className="feed-summary-link" href={`/events/${event.event_id}`}>
                    {localizeEventSummary(event.summary, language)}
                  </Link>
                  <p className="log-meta">
                    {event.actor_character_id ? (
                      <Link className="inline-link" href={`/bots/${event.actor_character_id}`}>
                        {event.actor_name ?? (language === "zh-CN" ? "未知角色" : "Unknown actor")}
                      </Link>
                    ) : (
                      <span>{event.actor_name ?? (language === "zh-CN" ? "未知角色" : "Unknown actor")}</span>
                    )}
                    <span>{formatRelativeTime(event.occurred_at, language)}</span>
                  </p>
                  <div className="entry-subline">
                    <span>{formatDateTime(event.occurred_at, language)}</span>
                    {event.region_id ? (
                      <Link className="inline-link" href={`/regions/${event.region_id}`}>
                        {localizeRegionName(event.region_id, event.region_id, language)}
                      </Link>
                    ) : null}
                  </div>
                </div>
              </article>
            ))
          ) : (
            <p className="world-empty">{text.chronicleEmpty}</p>
          )}
        </div>
      </section>
    </SiteFrame>
  );
}

function itemType(eventType: string) {
  if (eventType.startsWith("travel")) return "travel";
  if (eventType.startsWith("quest")) return "quest";
  if (eventType.startsWith("dungeon")) return "dungeon";
  if (eventType.startsWith("arena")) return "arena";
  return "all";
}
