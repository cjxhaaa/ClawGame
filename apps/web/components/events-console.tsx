"use client";

import Link from "next/link";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

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
  initialEvents: WorldEvent[];
  initialNextCursor: string | null;
  initialFilter?: string;
  initialRegion?: string;
  focusEventId?: string;
};

type PublicEventsEnvelope = {
  data: {
    items: WorldEvent[];
    next_cursor?: string | null;
  };
};

export default function EventsConsole({
  worldState,
  initialEvents,
  initialNextCursor,
  initialFilter,
  initialRegion,
  focusEventId,
}: EventsConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const common = uiText[language].common;
  const copy = uiText[language].events;
  const [filter, setFilter] = useState(toEventFilter(initialFilter));
  const [events, setEvents] = useState(initialEvents);
  const [nextCursor, setNextCursor] = useState<string | null>(initialNextCursor);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const selectedRegion = initialRegion
    ? worldState.regions.find((region) => region.region_id === initialRegion)
    : undefined;
  const regionScope = selectedRegion?.region_id;
  const filteredEvents = useMemo(
    () =>
      events.filter((event) => {
        if (regionScope && event.region_id !== regionScope) {
          return false;
        }

        return matchesEventFilter(event, filter);
      }),
    [events, filter, regionScope],
  );
  const regionHeat = [...worldState.regions].sort((left, right) => right.recent_event_count - left.recent_event_count);
  const arenaStatus = localizeArenaStatus(worldState.current_arena_status.code, language);
  const feedCopy =
    language === "zh-CN"
      ? {
          feedNote: "最新内容永远排在最前面，继续往下滚动时再逐步刷出更早的世界记录。",
          scopeLabel: "当前范围",
          clearScope: "查看全世界日志",
          backToRegion: "返回区域",
          loadingOlder: "正在加载更早的事件…",
          endReached: "已经到底了，当前没有更早的公开事件。",
        }
      : {
          feedNote: "Newest events stay on top. Keep scrolling down to progressively pull in older world records.",
          scopeLabel: "Current scope",
          clearScope: "View global feed",
          backToRegion: "Back to region",
          loadingOlder: "Loading older events…",
          endReached: "You've reached the end of the current public timeline.",
        };

  const loadMore = useCallback(async () => {
    if (!nextCursor || isLoadingMore) {
      return;
    }

    setIsLoadingMore(true);
    try {
      const response = await fetch(`/api/public-events?limit=20&cursor=${encodeURIComponent(nextCursor)}`, {
        cache: "no-store",
      });

      if (!response.ok) {
        throw new Error(`request failed with ${response.status}`);
      }

      const payload = (await response.json()) as PublicEventsEnvelope;
      setEvents((current) => {
        const seen = new Set(current.map((item) => item.event_id));
        const appended = payload.data.items.filter((item) => !seen.has(item.event_id));
        return [...current, ...appended];
      });
      setNextCursor(payload.data.next_cursor ?? null);
    } catch {
      setNextCursor(null);
    } finally {
      setIsLoadingMore(false);
    }
  }, [isLoadingMore, nextCursor]);

  useEffect(() => {
    const node = loadMoreRef.current;
    if (!node || !nextCursor) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          void loadMore();
        }
      },
      { rootMargin: "240px 0px" },
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [loadMore, nextCursor, filteredEvents.length]);

  useEffect(() => {
    if (filteredEvents.length === 0 && nextCursor && !isLoadingMore) {
      void loadMore();
    }
  }, [filteredEvents.length, isLoadingMore, loadMore, nextCursor]);

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
              <p className="section-note">{feedCopy.feedNote}</p>
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

            {selectedRegion ? (
              <div className="feed-scope-bar">
                <div>
                  <span>{feedCopy.scopeLabel}</span>
                  <strong>{localizeRegionName(selectedRegion.region_id, selectedRegion.name, language)}</strong>
                </div>
                <div className="feed-scope-actions">
                  <Link className="inline-link" href={`/regions/${selectedRegion.region_id}`}>
                    {feedCopy.backToRegion}
                  </Link>
                  <Link className="inline-link" href={filter === "all" ? "/events" : `/events?filter=${filter}`}>
                    {feedCopy.clearScope}
                  </Link>
                </div>
              </div>
            ) : (
              <p className="feed-behavior-note">{copy.filterNote}</p>
            )}

            <div className="log-list">
              {filteredEvents.length > 0 ? (
                filteredEvents.map((event) => {
                  const detailSearch = new URLSearchParams();
                  if (filter !== "all") detailSearch.set("filter", filter);
                  if (regionScope) detailSearch.set("region", regionScope);
                  detailSearch.set("focus", event.event_id);
                  const detailHref = `/events/${event.event_id}${detailSearch.toString() ? `?${detailSearch.toString()}` : ""}`;

                  return (
                    <article
                      key={event.event_id}
                      id={event.event_id}
                      className={`log-entry expanded-log-entry social-feed-entry ${
                        focusEventId === event.event_id ? "focused" : ""
                      }`}
                    >
                      <div className={`log-marker ${itemType(event.event_type)}`} />
                      <div>
                        <Link className="feed-summary-link" href={detailHref}>
                          {localizeEventSummary(event.summary, language)}
                        </Link>
                        <p className="log-meta">
                          {event.actor_character_id ? (
                            <Link className="inline-link" href={`/bots/${event.actor_character_id}`}>
                              {event.actor_name ?? common.unknownActor}
                            </Link>
                          ) : (
                            <span>{event.actor_name ?? common.unknownActor}</span>
                          )}
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
                  );
                })
              ) : (
                <p className="empty-state">{copy.emptyFiltered}</p>
              )}
            </div>

            {filteredEvents.length > 0 ? (
              <>
                {nextCursor ? <div ref={loadMoreRef} className="feed-load-sentinel" /> : null}
                <p className="feed-pagination-note">
                  {isLoadingMore ? feedCopy.loadingOlder : nextCursor ? "" : feedCopy.endReached}
                </p>
              </>
            ) : null}
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
