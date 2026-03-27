"use client";

import Link from "next/link";
import { useState } from "react";

import type { PublicWorldState, WorldEvent } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  filters,
  formatDateTime,
  formatMetric,
  formatRelativeTime,
  localizeArenaStatus,
  localizeEventSummary,
  localizeRegionName,
  matchesEventFilter,
  toEventFilter,
  uiText,
} from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type EventsConsoleProps = {
  worldState: PublicWorldState;
  events: WorldEvent[];
  initialFilter?: string;
};

export default function EventsConsole({ worldState, events, initialFilter }: EventsConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const common = uiText[language].common;
  const copy = uiText[language].events;
  const [filter, setFilter] = useState(toEventFilter(initialFilter));
  const filteredEvents = events.filter((event) => matchesEventFilter(event, filter));
  const regionHeat = [...worldState.regions].sort((left, right) => right.recent_event_count - left.recent_event_count);
  const arenaStatus = localizeArenaStatus(worldState.current_arena_status.code, language);

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="events"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={copy.eyebrow}
        title={copy.title}
        intro={copy.intro}
        stats={[
          {
            label: common.population,
            value: formatMetric(worldState.active_bot_count, language, ""),
          },
          {
            label: uiText[language].home.actionLog,
            value: formatMetric(events.length, language, ""),
          },
          {
            label: uiText[language].home.arenaState,
            value: arenaStatus.label,
          },
        ]}
      />

      <section className="detail-page-grid">
        <section className="detail-stack">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.timelineTitle}</p>
                <h2>{copy.timelineTitle}</h2>
              </div>
              <p className="section-note">{copy.filterNote}</p>
            </div>

            <div className="filter-row">
              {filters.map((item) => (
                <button
                  key={item.key}
                  type="button"
                  className={`filter-pill ${filter === item.key ? "active" : ""}`}
                  onClick={() => setFilter(item.key)}
                >
                  {item.label[language]}
                </button>
              ))}
            </div>

            <div className="log-list">
              {filteredEvents.length > 0 ? (
                filteredEvents.map((event) => (
                  <article key={event.event_id} id={event.event_id} className="log-entry expanded-log-entry">
                    <div className={`log-marker ${itemType(event.event_type)}`} />
                    <div>
                      <p className="log-summary">{localizeEventSummary(event.summary, language)}</p>
                      <p className="log-meta">
                        <span>{event.actor_name ?? common.unknownActor}</span>
                        <span>{formatRelativeTime(event.occurred_at, language)}</span>
                      </p>
                      <div className="entry-subline">
                        <span>
                          {common.occurredAt}: {formatDateTime(event.occurred_at, language)}
                        </span>
                        {event.region_id ? (
                          <Link className="inline-link" href={`/regions/${event.region_id}`}>
                            {common.linkedRegion}:{" "}
                            {localizeRegionName(event.region_id, event.region_id, language)}
                          </Link>
                        ) : null}
                      </div>
                    </div>
                  </article>
                ))
              ) : (
                <p className="empty-state">{copy.emptyFiltered}</p>
              )}
            </div>
          </section>
        </section>

        <aside className="detail-sidebar">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.regionContext}</p>
                <h2>{copy.regionContext}</h2>
              </div>
            </div>

            <div className="atlas-list">
              {regionHeat.map((region) => (
                <Link key={region.region_id} className="atlas-link" href={`/regions/${region.region_id}`}>
                  <strong>{localizeRegionName(region.region_id, region.name, language)}</strong>
                  <span>
                    {formatMetric(region.recent_event_count, language, "")} /{" "}
                    {formatMetric(region.population, language, "")}
                  </span>
                </Link>
              ))}
            </div>
          </section>
        </aside>
      </section>
    </main>
  );
}

function itemType(value: string) {
  if (value.startsWith("travel")) {
    return "travel";
  }
  if (value.startsWith("quest")) {
    return "quest";
  }
  if (value.startsWith("dungeon")) {
    return "dungeon";
  }
  if (value.startsWith("arena")) {
    return "arena";
  }

  return "all";
}
