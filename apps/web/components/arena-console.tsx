"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import type { ArenaCurrent, ArenaEntry, ArenaMatchup, ArenaRound, Leaderboards, PublicWorldState, WorldEvent } from "../lib/public-api";
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
  arenaCurrent: ArenaCurrent;
};

export default function ArenaConsole({ worldState, leaderboards, events, arenaCurrent }: ArenaConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const copy = uiText[language].arena;
  const common = uiText[language].common;
  const home = uiText[language].home;
  const arenaStatus = localizeArenaStatus(worldState.current_arena_status.code, language);
  const rounds = arenaCurrent.rounds ?? [];
  const qualifierMatchups = arenaCurrent.qualifier_matchups ?? [];
  const currentMatchups = arenaCurrent.matchups ?? [];
  const contenders = leaderboards.weekly_arena.slice(0, 4);
  const arenaEvents = events.filter((event) => event.event_type.startsWith("arena"));

  const highlightedRound = useMemo(
    () =>
      rounds.find((round) => round.status === "in_progress") ??
      rounds.find((round) => round.status === "scheduled") ??
      rounds[rounds.length - 1] ??
      null,
    [rounds],
  );

  const [selectedRoundName, setSelectedRoundName] = useState<string>(highlightedRound?.name ?? "");

  useEffect(() => {
    if (!rounds.some((round) => round.name === selectedRoundName)) {
      setSelectedRoundName(highlightedRound?.name ?? "");
    }
  }, [highlightedRound, rounds, selectedRoundName]);

  const selectedRound = rounds.find((round) => round.name === selectedRoundName) ?? highlightedRound ?? null;
  const resolvedRounds = rounds.filter((round) => round.status === "resolved").length;
  const progressPercent = rounds.length > 0 ? Math.round((resolvedRounds / rounds.length) * 100) : 0;
  const liveRoundCount = currentMatchups.length > 0 ? currentMatchups.length : qualifierMatchups.length;

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
          { label: copy.entrants, value: formatMetric(arenaCurrent.signup_count, language, "") },
          { label: home.arenaNext, value: arenaStatus.nextMilestone },
        ]}
      />

      <section className="arena-spectacle-grid">
        <section className="pixel-panel detail-panel arena-stage-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.liveBracket}</p>
              <h2>{copy.liveBracket}</h2>
            </div>
            <div className={`arena-stage-pill ${arenaStatusTone(arenaCurrent.status.code)}`}>{arenaStatus.label}</div>
          </div>

          <div className="arena-stage-hero">
            <div className="arena-stage-copy">
              <strong>{selectedRound?.name ?? copy.qualifierLabel}</strong>
              <p>{arenaStatus.details}</p>
              <div className="arena-progress-rail">
                <span className="arena-progress-fill" style={{ width: `${progressPercent}%` }} />
              </div>
              <div className="arena-stage-metrics">
                <article className="hero-metric">
                  <span>{copy.realEntrants}</span>
                  <strong>{formatMetric(arenaCurrent.signup_count, language, "")}</strong>
                </article>
                <article className="hero-metric">
                  <span>{copy.npcFill}</span>
                  <strong>{formatMetric(arenaCurrent.npc_count, language, "")}</strong>
                </article>
                <article className="hero-metric">
                  <span>{copy.activeRound}</span>
                  <strong>{selectedRound?.name ?? arenaStatus.label}</strong>
                </article>
              </div>
              <div className="arena-stage-summary">
                <article className="arena-status-card arena-info-card">
                  <strong>{copy.statusLabel}</strong>
                  <p>{arenaStatus.label}</p>
                  <span>{home.arenaNext}: {arenaStatus.nextMilestone}</span>
                </article>
                <article className="arena-status-card arena-info-card">
                  <strong>{copy.matchupCount}</strong>
                  <p>{formatMetric(liveRoundCount, language, "")}</p>
                  <span>{selectedRound ? localizeRoundStatus(selectedRound.status, language) : arenaStatus.label}</span>
                </article>
              </div>
            </div>
          </div>
        </section>

        <section className="pixel-panel detail-panel arena-timeline-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.roundTimeline}</p>
              <h2>{copy.roundTimeline}</h2>
            </div>
          </div>

          {rounds.length > 0 ? (
            <div className="arena-round-track">
              {rounds.map((round) => (
                <button
                  key={round.name}
                  type="button"
                  className={`arena-round-node ${selectedRound?.name === round.name ? "active" : ""} ${round.status}`}
                  onClick={() => setSelectedRoundName(round.name)}
                >
                  <span className="arena-round-node-step">{formatArenaClock(round.scheduled_at, language)}</span>
                  <strong>{round.name}</strong>
                  <span>{localizeRoundStatus(round.status, language)}</span>
                </button>
              ))}
            </div>
          ) : (
            <p className="empty-state">{copy.noBracketData}</p>
          )}
        </section>
      </section>

      <section className="detail-page-grid arena-layout-grid">
        <section className="detail-stack">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.fullBracket}</p>
                <h2>{selectedRound?.name ?? copy.fullBracket}</h2>
              </div>
              <Link className="section-link" href="/leaderboards?board=weekly_arena">
                {common.openBoard}
              </Link>
            </div>

            {selectedRound ? (
              <>
                <div className="arena-round-summary">
                  <span>{copy.statusLabel}: {localizeRoundStatus(selectedRound.status, language)}</span>
                  <span>{copy.roundStarts}: {formatArenaClock(selectedRound.scheduled_at, language)}</span>
                  <span>{copy.roundResolved}: {selectedRound.resolved_at ? formatArenaClock(selectedRound.resolved_at, language) : (language === "zh-CN" ? "待结算" : "Pending")}</span>
                </div>
                <div className="arena-match-grid expanded">
                  {selectedRound.matchups.map((matchup) => (
                    <ArenaMatchCard
                      key={`${selectedRound.name}-${matchup.match_number}`}
                      matchup={matchup}
                      language={language}
                      layout="expanded"
                    />
                  ))}
                </div>
              </>
            ) : (
              <p className="empty-state">{copy.noBracketData}</p>
            )}
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
          <section className="pixel-panel detail-panel arena-champion-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.championDeck}</p>
                <h2>{copy.championDeck}</h2>
              </div>
            </div>

            {arenaCurrent.champion ? (
              <article className="arena-champion-card">
                <div className="arena-crown">01</div>
                <strong>{arenaCurrent.champion.character_name}</strong>
                <p className="bot-classline">
                  {localizeClass(arenaCurrent.champion.class, language)} / {localizeWeapon(arenaCurrent.champion.weapon_style, language)}
                </p>
                <div className="bot-stat-row">
                  <span>{common.scoreLabel}</span>
                  <strong>{formatMetric(arenaCurrent.champion.equipment_score, language, "")}</strong>
                </div>
              </article>
            ) : (
              <p className="empty-state">{copy.noChampionYet}</p>
            )}
          </section>

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.fieldBreakdown}</p>
                <h2>{copy.fieldBreakdown}</h2>
              </div>
            </div>

            <div className="arena-field-breakdown">
              <article className="arena-status-card arena-info-card">
                <strong>{copy.realEntrants}</strong>
                <p>{formatMetric(arenaCurrent.signup_count, language, "")}</p>
              </article>
              <article className="arena-status-card arena-info-card">
                <strong>{copy.qualifierLabel}</strong>
                <p>{formatMetric(arenaCurrent.qualified_count, language, "")}</p>
              </article>
              <article className="arena-status-card arena-info-card">
                <strong>{copy.npcFill}</strong>
                <p>{formatMetric(arenaCurrent.npc_count, language, "")}</p>
              </article>
            </div>
            <p className="section-note">{copy.npcNote}</p>
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

function ArenaCombatantCard({ entry, language }: { entry?: ArenaEntry; language: "zh-CN" | "en-US" }) {
  if (!entry) {
    return <div className="arena-combatant ghost">{language === "zh-CN" ? "轮空" : "Bye"}</div>;
  }

  return (
    <div className={`arena-combatant ${entry.is_npc ? "npc" : ""}`}>
      <strong>{entry.character_name}</strong>
      <span>
        {localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}
      </span>
      <span>{language === "zh-CN" ? "装备分" : "Gear"} {entry.equipment_score}</span>
    </div>
  );
}

function ArenaMatchCard({
  matchup,
  language,
  layout,
}: {
  matchup: ArenaMatchup;
  language: "zh-CN" | "en-US";
  layout: "compact" | "expanded";
}) {
  return (
    <article className={`arena-match-card ${matchup.status} ${layout}`}>
      <div className="arena-match-head">
        <strong>#{matchup.match_number}</strong>
        <span>{localizeMatchStatus(matchup.status, language)}</span>
      </div>
      <div className={`arena-match-bout ${layout}`}>
        <ArenaCombatantCard entry={matchup.left_entry} language={language} />
        <div className={`arena-mini-versus ${layout}`}>VS</div>
        <ArenaCombatantCard entry={matchup.right_entry} language={language} />
      </div>
      <div className="arena-match-foot">
        <span>{formatArenaClock(matchup.scheduled_at, language)}</span>
        {matchup.winner_entry ? <span>{language === "zh-CN" ? "胜者" : "Winner"}: {matchup.winner_entry.character_name}</span> : null}
      </div>
    </article>
  );
}

function localizeMatchStatus(status: string, language: "zh-CN" | "en-US") {
  const labels: Record<string, { "zh-CN": string; "en-US": string }> = {
    scheduled: { "zh-CN": "待开打", "en-US": "Scheduled" },
    in_progress: { "zh-CN": "进行中", "en-US": "Live" },
    resolved: { "zh-CN": "已结算", "en-US": "Resolved" },
    walkover: { "zh-CN": "轮空晋级", "en-US": "Walkover" },
  };
  return labels[status]?.[language] ?? status;
}

function localizeRoundStatus(status: string, language: "zh-CN" | "en-US") {
  const labels: Record<string, { "zh-CN": string; "en-US": string }> = {
    scheduled: { "zh-CN": "待开始", "en-US": "Scheduled" },
    in_progress: { "zh-CN": "进行中", "en-US": "In Progress" },
    resolved: { "zh-CN": "已完成", "en-US": "Resolved" },
  };
  return labels[status]?.[language] ?? status;
}

function arenaStatusTone(code: string) {
  if (code === "in_progress") return "live";
  if (code === "results_live") return "complete";
  if (code === "signup_locked") return "seeding";
  return "open";
}

function formatArenaClock(value: string, language: "zh-CN" | "en-US") {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat(language, {
    hour: "2-digit",
    minute: "2-digit",
    month: "2-digit",
    day: "2-digit",
  }).format(date);
}
