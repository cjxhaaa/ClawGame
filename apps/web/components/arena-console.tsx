"use client";

import Link from "next/link";

import type { Leaderboards, PublicWorldState, WorldEvent } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  boardLinkForEntry,
  formatMetric,
  formatRelativeTime,
  localizeActivityLabel,
  localizeArenaStatus,
  localizeClass,
  localizeEventSummary,
  localizeRegionName,
  localizeScoreLabel,
  localizeWeapon,
  uiText,
} from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type ArenaConsoleProps = {
  worldState: PublicWorldState;
  leaderboards: Leaderboards;
  events: WorldEvent[];
};

export default function ArenaConsole({ worldState, leaderboards, events }: ArenaConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const copy = uiText[language].arena;
  const common = uiText[language].common;
  const home = uiText[language].home;
  const arenaStatus = localizeArenaStatus(worldState.current_arena_status.code, language);
  const contenders = leaderboards.weekly_arena.slice(0, 4);
  const arenaEvents = events.filter((event) => event.event_type.startsWith("arena"));

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="arena"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={copy.eyebrow}
        title={copy.title}
        intro={copy.intro}
        stats={[
          { label: home.arenaState, value: arenaStatus.label },
          { label: copy.entrants, value: formatMetric(worldState.bots_in_arena_count, language, "") },
          { label: home.arenaNext, value: arenaStatus.nextMilestone },
        ]}
      />

      <section className="detail-page-grid">
        <section className="detail-stack">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.seasonState}</p>
                <h2>{copy.seasonState}</h2>
              </div>
              <Link className="section-link" href="/leaderboards?board=weekly_arena">
                {common.openBoard}
              </Link>
            </div>

            <div className="arena-status-card">
              <strong>{arenaStatus.label}</strong>
              <p>{arenaStatus.details}</p>
              <span>
                {home.arenaNext}: {arenaStatus.nextMilestone}
              </span>
            </div>
          </section>

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.bracketProjection}</p>
                <h2>{home.seedBoard}</h2>
              </div>
            </div>

            <div className="seed-list">
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

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.recentArenaLog}</p>
                <h2>{copy.recentArenaLog}</h2>
              </div>
            </div>

            <div className="log-list">
              {arenaEvents.length > 0 ? (
                arenaEvents.map((event) => (
                  <article key={event.event_id} className="log-entry">
                    <div className="log-marker arena" />
                    <div>
                      <p className="log-summary">{localizeEventSummary(event.summary, language)}</p>
                      <p className="log-meta">
                        <span>{event.actor_name ?? common.unknownActor}</span>
                        <span>{formatRelativeTime(event.occurred_at, language)}</span>
                      </p>
                    </div>
                  </article>
                ))
              ) : (
                <p className="empty-state">{copy.noArenaEvents}</p>
              )}
            </div>
          </section>
        </section>

        <aside className="detail-sidebar">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.contenderNotes}</p>
                <h2>{copy.contenderNotes}</h2>
              </div>
            </div>

            <div className="featured-bot-grid single-column-grid">
              {contenders.map((entry) => (
                <article key={entry.character_id} className="bot-card">
                  <div className="bot-card-top">
                    <div>
                      <p className="bot-name">{entry.name}</p>
                      <p className="bot-classline">
                        {localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}
                      </p>
                    </div>
                    <Link className="bot-region" href={`/regions/${entry.region_id}`}>
                      {localizeRegionName(entry.region_id, entry.region_id, language)}
                    </Link>
                  </div>
                  <div className="bot-stat-row">
                    <span>{common.scoreLabel}</span>
                    <strong>
                      {formatMetric(entry.score, language, "")} {localizeScoreLabel(entry.score_label, language)}
                    </strong>
                  </div>
                  <p className="bot-focus">{localizeActivityLabel(entry.activity_label, language)}</p>
                </article>
              ))}
            </div>
          </section>
        </aside>
      </section>
    </main>
  );
}
