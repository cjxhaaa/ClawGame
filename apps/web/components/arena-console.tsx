"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import type { ArenaCurrent, ArenaEntry, ArenaMatchDetail, ArenaMatchup, Leaderboards, PublicWorldState, WorldEvent } from "../lib/public-api";
import { getArenaMatchDetail } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import { formatMetric, formatRelativeTime, localizeArenaStatus, localizeClass, localizeEventSummary, localizeWeapon, uiText } from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type ArenaConsoleProps = {
  worldState: PublicWorldState;
  leaderboards: Leaderboards;
  events: WorldEvent[];
  arenaCurrent: ArenaCurrent;
};

type SelectedMatch = {
  matchup: ArenaMatchup;
  roundName: string;
};

type ArenaTab = "rating" | "knockout";

export default function ArenaConsole({ worldState, events, arenaCurrent }: ArenaConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const copy = uiText[language].arena;
  const common = uiText[language].common;
  const home = uiText[language].home;
  const text = labels(language);
  const arenaStatus = localizeArenaStatus(worldState.current_arena_status.code, language);
  const arenaEvents = events.filter((event) => event.event_type.startsWith("arena"));
  const ratingSummary = arenaCurrent.weekly_rating_summary;
  const ratingFeatured = ratingSummary?.featured ?? [];
  const podiumEntries = ratingFeatured.slice(0, 3);
  const ladderEntries = ratingFeatured.slice(3);
  const stageRounds = useMemo(() => (arenaCurrent.rounds ?? []).filter((round) => round.matchups.length > 0), [arenaCurrent.rounds]);
  const [selectedMatch, setSelectedMatch] = useState<SelectedMatch | null>(null);
  const [selectedDetail, setSelectedDetail] = useState<ArenaMatchDetail | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [selectedRoundName, setSelectedRoundName] = useState<string>(stageRounds[0]?.name ?? "");
  const [selectedTab, setSelectedTab] = useState<ArenaTab>(isKnockoutWeekState(arenaCurrent.status.code) ? "knockout" : "rating");

  useEffect(() => {
    if (stageRounds.length === 0) {
      setSelectedRoundName("");
      return;
    }
    if (!stageRounds.some((round) => round.name === selectedRoundName)) {
      setSelectedRoundName(stageRounds[0].name);
    }
  }, [stageRounds, selectedRoundName]);

  useEffect(() => {
    if (isKnockoutWeekState(arenaCurrent.status.code)) {
      setSelectedTab("knockout");
    }
  }, [arenaCurrent.status.code]);

  const selectedStageRound = stageRounds.find((round) => round.name === selectedRoundName) ?? stageRounds[0] ?? null;

  async function openMatch(matchup: ArenaMatchup, roundName: string) {
    setSelectedMatch({ matchup, roundName });
    setSelectedDetail(null);
    if (matchup.status !== "resolved") return;
    setLoadingDetail(true);
    try {
      setSelectedDetail(await getArenaMatchDetail(matchup.match_id, "standard"));
    } finally {
      setLoadingDetail(false);
    }
  }

  return (
    <main className="console-shell pixel-theme arena-console-shell">
      <PortalChrome
        active="arena"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={copy.eyebrow}
        title={copy.title}
        intro={copy.intro}
        stats={[
          { label: home.arenaState, value: arenaStatus.label },
          { label: text.weekLabel, value: arenaCurrent.week_key ?? "-" },
          { label: home.arenaNext, value: arenaStatus.nextMilestone },
        ]}
      />

      <section className="arena-overview-strip">
        <SummaryMetric label={text.phaseLabel} value={arenaStatus.label} note={arenaStatus.details} />
        <SummaryMetric label={text.activeContestants} value={formatMetric(ratingSummary?.active_count ?? arenaCurrent.signup_count, language, "")} note={text.activeContestantsNote} />
        <SummaryMetric label={text.highestRating} value={formatMetric(ratingSummary?.highest_rating ?? 0, language, "")} note={text.highestRatingNote} />
        <SummaryMetric label={text.knockoutField} value={formatMetric(arenaCurrent.qualified_count, language, "")} note={text.knockoutFieldNote} />
      </section>

      <section className="arena-brief-grid">
        <section className="pixel-panel detail-panel arena-brief-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{text.weekDesk}</p>
              <h2>{text.weekDesk}</h2>
            </div>
            <div className={`arena-stage-pill ${arenaTone(arenaCurrent.status.code)}`}>{arenaStatus.label}</div>
          </div>

          <div className="arena-summary-grid">
            <article className="arena-summary-card highlight">
              <span>{text.status}</span>
              <strong>{arenaStatus.label}</strong>
              <p>{arenaStatus.details}</p>
            </article>
            <article className="arena-summary-card">
              <span>{text.nextResolve}</span>
              <strong>{formatArenaClock(arenaCurrent.next_round_time, language)}</strong>
              <p>{text.nextResolveNote}</p>
            </article>
            <article className="arena-summary-card">
              <span>{text.knockoutField}</span>
              <strong>{formatMetric(arenaCurrent.qualified_count, language, "")}</strong>
              <p>{text.npcFillNote(arenaCurrent.npc_count)}</p>
            </article>
          </div>

          <div className="arena-view-switch">
            <button type="button" className={`arena-view-tab ${selectedTab === "rating" ? "active" : ""}`} onClick={() => setSelectedTab("rating")}>
              {text.ratingTab}
            </button>
            <button type="button" className={`arena-view-tab ${selectedTab === "knockout" ? "active" : ""}`} onClick={() => setSelectedTab("knockout")}>
              {text.knockoutTab}
            </button>
          </div>
        </section>

        <section className="pixel-panel detail-panel arena-brief-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{text.recentSignals}</p>
              <h2>{text.recentSignals}</h2>
            </div>
          </div>
          <div className="log-list">
            {arenaEvents.length > 0 ? (
              arenaEvents.slice(0, 5).map((event) => (
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

      {selectedTab === "rating" ? (
        <section className="detail-page-grid arena-redesign-grid">
          <section className="pixel-panel detail-panel arena-bracket-shell">
            <div className="section-header">
              <div>
                <p className="eyebrow">{text.ratingLadder}</p>
                <h2>{text.ratingLadder}</h2>
              </div>
              <span className="section-note">{text.ratingLadderHint}</span>
            </div>

            <div className="arena-rating-headline">
              <article className="arena-summary-card">
                <span>{text.activeContestants}</span>
                <strong>{formatMetric(ratingSummary?.active_count ?? 0, language, "")}</strong>
                <p>{text.activeContestantsHint}</p>
              </article>
              <article className="arena-summary-card">
                <span>{text.highestRating}</span>
                <strong>{formatMetric(ratingSummary?.highest_rating ?? 0, language, "")}</strong>
                <p>{text.highestRatingHint}</p>
              </article>
              <article className="arena-summary-card">
                <span>{text.lowestFeaturedRating}</span>
                <strong>{formatMetric(ratingSummary?.lowest_rating ?? 0, language, "")}</strong>
                <p>{text.lowestFeaturedHint}</p>
              </article>
            </div>

            {ratingFeatured.length > 0 ? (
              <>
                <div className="arena-podium-grid">
                  {podiumEntries.map((entry, index) => (
                    <article key={entry.character_id} className={`arena-podium-card place-${index + 1} ${entry.is_npc ? "npc" : ""}`}>
                      <div className="arena-podium-topline">
                        <span className="arena-podium-place">#{index + 1}</span>
                        <span className="arena-podium-badge">{index === 0 ? text.podiumChampion : index === 1 ? text.podiumRunnerUp : text.podiumThird}</span>
                      </div>
                      <strong>
                        <Link href={`/bots/${encodeURIComponent(entry.character_id)}`}>{entry.character_name}</Link>
                      </strong>
                      <p>
                        {localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}
                      </p>
                      <div className="arena-podium-stats">
                        <article>
                          <span>{text.ratingShort}</span>
                          <strong>{formatMetric(entry.rating, language, "")}</strong>
                        </article>
                        <article>
                          <span>{text.powerShort}</span>
                          <strong>{formatMetric(entry.panel_power_score, language, "")}</strong>
                        </article>
                      </div>
                    </article>
                  ))}
                </div>

                <div className="arena-cutline-banner">
                  <span>{text.cutlineBannerLabel}</span>
                  <strong>{text.cutlineBannerValue}</strong>
                  <p>{text.cutlineBannerNote}</p>
                </div>

                <div className="arena-ladder-table">
                  <div className="arena-ladder-header">
                    <span>{text.rank}</span>
                    <span>{text.competitor}</span>
                    <span>{text.ratingShort}</span>
                    <span>{text.powerShort}</span>
                  </div>
                  {ladderEntries.map((entry, index) => (
                    <article key={entry.character_id} className={`arena-ladder-row ${entry.is_npc ? "npc" : ""}`}>
                      <span className="arena-ladder-rank">#{index + 4}</span>
                      <div className="arena-ladder-bot">
                        <Link href={`/bots/${encodeURIComponent(entry.character_id)}`}>{entry.character_name}</Link>
                        <small>
                          {localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}
                        </small>
                      </div>
                      <strong>{formatMetric(entry.rating, language, "")}</strong>
                      <span>{formatMetric(entry.panel_power_score, language, "")}</span>
                    </article>
                  ))}
                </div>
              </>
            ) : (
              <p className="empty-state">{text.noRatingData}</p>
            )}
          </section>

          <aside className="detail-sidebar arena-context-sidebar">
            <section className="pixel-panel detail-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{text.selectionNotes}</p>
                  <h2>{text.selectionNotes}</h2>
                </div>
              </div>
              <div className="arena-note-stack">
                <article className="arena-note-card">
                  <strong>{text.top64RuleTitle}</strong>
                  <p>{text.top64RuleBody}</p>
                </article>
                <article className="arena-note-card">
                  <strong>{text.challengeWindowTitle}</strong>
                  <p>{text.challengeWindowBody}</p>
                </article>
                <article className="arena-note-card">
                  <strong>{text.cutlineTitle}</strong>
                  <p>{text.cutlineBody}</p>
                </article>
              </div>
            </section>
          </aside>
        </section>
      ) : (
        <section className="detail-page-grid arena-redesign-grid">
          <section className="pixel-panel detail-panel arena-bracket-shell">
            <div className="section-header">
              <div>
                <p className="eyebrow">{text.knockoutDesk}</p>
                <h2>{text.knockoutDesk}</h2>
              </div>
              <span className="section-note">{text.roundSwitchHint}</span>
            </div>

            {stageRounds.length > 0 ? (
              <>
                <div className="arena-round-tab-row">
                  {stageRounds.map((round) => (
                    <button
                      key={round.name}
                      type="button"
                      className={`arena-round-tab ${selectedStageRound?.name === round.name ? "active" : ""}`}
                      onClick={() => setSelectedRoundName(round.name)}
                    >
                      <span>{round.name}</span>
                      <small>{localizeRoundStatus(round.status, language)}</small>
                    </button>
                  ))}
                </div>

                {selectedStageRound ? (
                  <div className="arena-stage-round-grid">
                    {selectedStageRound.matchups.map((matchup) => (
                      <BracketMatchCard
                        key={matchup.match_id}
                        matchup={matchup}
                        roundName={selectedStageRound.name}
                        language={language}
                        onOpen={openMatch}
                      />
                    ))}
                  </div>
                ) : (
                  <p className="empty-state">{copy.noBracketData}</p>
                )}
              </>
            ) : (
              <p className="empty-state">{copy.noBracketData}</p>
            )}
          </section>

          <aside className="detail-sidebar arena-context-sidebar">
            <section className="pixel-panel detail-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{text.knockoutNotes}</p>
                  <h2>{text.knockoutNotes}</h2>
                </div>
              </div>
              <div className="arena-note-stack">
                <article className="arena-note-card">
                  <strong>{text.seedSourceTitle}</strong>
                  <p>{text.seedSourceBody}</p>
                </article>
                <article className="arena-note-card">
                  <strong>{text.powerViewTitle}</strong>
                  <p>{text.powerViewBody}</p>
                </article>
                <article className="arena-note-card">
                  <strong>{text.reportViewTitle}</strong>
                  <p>{text.reportViewBody}</p>
                </article>
              </div>
            </section>

            <section className="pixel-panel detail-panel">
              <div className="section-header">
                <div>
                  <p className="eyebrow">{text.championDesk}</p>
                  <h2>{text.championDesk}</h2>
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
                    <span>{text.power}</span>
                    <strong>{formatMetric(arenaCurrent.champion.panel_power_score, language, "")}</strong>
                  </div>
                  <Link className="section-link" href={`/bots/${encodeURIComponent(arenaCurrent.champion.character_id)}`}>
                    {text.openBot}
                  </Link>
                </article>
              ) : (
                <p className="empty-state">{copy.noChampionYet}</p>
              )}
            </section>
          </aside>
        </section>
      )}

      {selectedMatch ? (
        <div className="region-modal-overlay" onClick={() => setSelectedMatch(null)}>
          <section className="pixel-panel region-modal-panel arena-match-modal" onClick={(event) => event.stopPropagation()}>
            <div className="region-modal-header">
              <div>
                <p className="eyebrow">{text.matchDetail}</p>
                <h2>{selectedMatch.roundName}</h2>
              </div>
              <button type="button" className="section-link" onClick={() => setSelectedMatch(null)}>
                {text.close}
              </button>
            </div>

            <div className="arena-match-modal-grid">
              <section className="arena-modal-block">
                <h3>{text.combatants}</h3>
                <div className="arena-modal-combatants">
                  <DetailCard entry={selectedDetail?.left_entry ?? selectedMatch.matchup.left_entry} language={language} />
                  <div className="arena-modal-versus">VS</div>
                  <DetailCard entry={selectedDetail?.right_entry ?? selectedMatch.matchup.right_entry} language={language} />
                </div>
              </section>

              <section className="arena-modal-block">
                <h3>{text.snapshot}</h3>
                <div className="arena-match-detail-stats">
                  <DetailStat label={text.status} value={localizeMatchStatus(selectedMatch.matchup.status, language)} />
                  <DetailStat label={text.startTime} value={formatArenaClock(selectedMatch.matchup.scheduled_at, language)} />
                  <DetailStat label={text.resolveTime} value={selectedMatch.matchup.resolved_at ? formatArenaClock(selectedMatch.matchup.resolved_at, language) : text.pending} />
                  <DetailStat label={text.winner} value={selectedDetail?.winner_entry?.character_name ?? selectedMatch.matchup.winner_entry?.character_name ?? text.pending} />
                </div>
              </section>

              <section className="arena-modal-block">
                <h3>{text.battleReport}</h3>
                {loadingDetail ? (
                  <p className="empty-state">{text.loading}</p>
                ) : selectedDetail?.battle_report ? (
                  <div className="arena-report-grid">
                    {Object.entries(selectedDetail.battle_report).map(([key, value]) => (
                      <article key={key} className="arena-report-item">
                        <span>{reportLabel(key, language)}</span>
                        <strong>{reportValue(key, value, language)}</strong>
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className="empty-state">{selectedMatch.matchup.status === "resolved" ? text.noReport : text.pendingMatch}</p>
                )}
              </section>
            </div>
          </section>
        </div>
      ) : null}
    </main>
  );
}

function SummaryMetric({ label, value, note }: { label: string; value: string; note: string }) {
  return (
    <article className="arena-metric-panel">
      <span>{label}</span>
      <strong>{value}</strong>
      <p>{note}</p>
    </article>
  );
}

function BracketMatchCard({
  matchup,
  roundName,
  language,
  onOpen,
}: {
  matchup: ArenaMatchup;
  roundName: string;
  language: "zh-CN" | "en-US";
  onOpen: (matchup: ArenaMatchup, roundName: string) => void;
}) {
  return (
    <button type="button" className={`arena-bracket-match ${matchup.status}`} onClick={() => onOpen(matchup, roundName)}>
      <div className="arena-bracket-match-head">
        <span>{roundName}</span>
        <strong>#{matchup.match_number}</strong>
      </div>
      <div className="arena-bracket-combatant-list">
        <BracketEntry entry={matchup.left_entry} language={language} winnerID={matchup.winner_entry?.character_id} />
        <BracketEntry entry={matchup.right_entry} language={language} winnerID={matchup.winner_entry?.character_id} />
        {matchup.bye_entry ? <BracketEntry entry={matchup.bye_entry} language={language} winnerID={matchup.bye_entry.character_id} bye /> : null}
      </div>
      <div className="arena-bracket-match-foot">
        <span>{localizeMatchStatus(matchup.status, language)}</span>
        {matchup.battle_report_id ? <small>{language === "zh-CN" ? "可查看战报" : "Report ready"}</small> : null}
      </div>
    </button>
  );
}

function BracketEntry({ entry, language, winnerID, bye = false }: { entry?: ArenaEntry; language: "zh-CN" | "en-US"; winnerID?: string; bye?: boolean }) {
  if (!entry) return <div className="arena-bracket-entry ghost">{language === "zh-CN" ? "等待对手" : "Pending entrant"}</div>;
  return (
    <div className={`arena-bracket-entry ${winnerID === entry.character_id ? "winner" : ""} ${entry.is_npc ? "npc" : ""}`}>
      <div>
        <Link href={`/bots/${encodeURIComponent(entry.character_id)}`} onClick={(event) => event.stopPropagation()}>
          {entry.character_name}
        </Link>
        <small>
          {localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}
        </small>
      </div>
      <span>{bye ? (language === "zh-CN" ? "轮空" : "Bye") : formatMetric(entry.panel_power_score, language, "")}</span>
    </div>
  );
}

function DetailCard({ entry, language }: { entry?: { character_id: string; character_name: string; class: string; weapon_style: string; panel_power_score: number; is_npc?: boolean }; language: "zh-CN" | "en-US" }) {
  if (!entry) return <div className="arena-detail-entry-card ghost">{language === "zh-CN" ? "暂无选手" : "No entrant"}</div>;
  return (
    <article className={`arena-detail-entry-card ${entry.is_npc ? "npc" : ""}`}>
      <strong>
        <Link href={`/bots/${encodeURIComponent(entry.character_id)}`}>{entry.character_name}</Link>
      </strong>
      <p>
        {localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}
      </p>
      <div className="bot-stat-row">
        <span>{language === "zh-CN" ? "总战力" : "Power"}</span>
        <strong>{formatMetric(entry.panel_power_score, language, "")}</strong>
      </div>
    </article>
  );
}

function DetailStat({ label, value }: { label: string; value: string }) {
  return (
    <article className="arena-match-detail-stat">
      <span>{label}</span>
      <strong>{value}</strong>
    </article>
  );
}

function isKnockoutWeekState(code: string) {
  return code === "knockout_pending" || code === "knockout_in_progress" || code === "knockout_results_live";
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

function arenaTone(code: string) {
  if (code === "knockout_in_progress") return "live";
  if (code === "knockout_results_live") return "complete";
  if (code === "knockout_pending") return "seeding";
  if (code === "rest_day") return "rest";
  return "open";
}

function formatArenaClock(value: string, language: "zh-CN" | "en-US") {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(language, { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }).format(date);
}

function reportLabel(key: string, language: "zh-CN" | "en-US") {
  const labels: Record<string, { "zh-CN": string; "en-US": string }> = {
    outcome: { "zh-CN": "结果", "en-US": "Outcome" },
    winner: { "zh-CN": "胜者", "en-US": "Winner" },
    loser: { "zh-CN": "败者", "en-US": "Loser" },
    winner_final_hp: { "zh-CN": "胜者剩余生命", "en-US": "Winner HP" },
    left_final_hp: { "zh-CN": "左侧剩余生命", "en-US": "Left HP" },
    right_final_hp: { "zh-CN": "右侧剩余生命", "en-US": "Right HP" },
    end_reason: { "zh-CN": "结束原因", "en-US": "End Reason" },
    adjudication: { "zh-CN": "裁定方式", "en-US": "Adjudication" },
    power_delta: { "zh-CN": "战力差", "en-US": "Power Delta" },
    player_power_delta: { "zh-CN": "战力差", "en-US": "Power Delta" },
    battle_length_events: { "zh-CN": "事件数", "en-US": "Event Count" },
    summary_tag: { "zh-CN": "摘要标签", "en-US": "Summary Tag" },
    rating_delta: { "zh-CN": "积分变化", "en-US": "Rating Delta" },
  };
  return labels[key]?.[language] ?? key;
}

function localizedReportEnum(key: string, value: string, language: "zh-CN" | "en-US") {
  const labels: Record<string, Record<string, { "zh-CN": string; "en-US": string }>> = {
    outcome: {
      win: { "zh-CN": "胜利", "en-US": "Win" },
      loss: { "zh-CN": "失败", "en-US": "Loss" },
      bye: { "zh-CN": "轮空", "en-US": "Bye" },
      resolved: { "zh-CN": "已结算", "en-US": "Resolved" },
      stable: { "zh-CN": "守擂成功", "en-US": "Defense Held" },
    },
    end_reason: {
      round_cap: { "zh-CN": "达到回合上限", "en-US": "Round Cap Reached" },
      defeat: { "zh-CN": "一方被击败", "en-US": "Defeat" },
    },
    adjudication: {
      challenger_loss_at_cap: { "zh-CN": "积分挑战超回合判挑战者负", "en-US": "Challenger Loses At Cap" },
      higher_remaining_hp: { "zh-CN": "按剩余生命裁定", "en-US": "Higher Remaining HP" },
      lower_character_id: { "zh-CN": "按较小角色 ID 裁定", "en-US": "Lower Character ID" },
      direct_victory: { "zh-CN": "直接击败获胜", "en-US": "Direct Victory" },
    },
    summary_tag: {
      rating_duel: { "zh-CN": "积分对决", "en-US": "Rating Duel" },
      advanced_by_bye: { "zh-CN": "轮空晋级", "en-US": "Advanced By Bye" },
      won_clean: { "zh-CN": "优势获胜", "en-US": "Clean Win" },
      won_close: { "zh-CN": "险胜", "en-US": "Close Win" },
      unresolved: { "zh-CN": "未结算", "en-US": "Unresolved" },
    },
  };
  return labels[key]?.[value]?.[language] ?? value;
}

function reportValue(key: string, value: unknown, language: "zh-CN" | "en-US") {
  if (value === null || value === undefined) return "-";
  if (typeof value === "string") return localizedReportEnum(key, value, language);
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  if (typeof value === "object" && value && "character_name" in value) return String((value as { character_name?: string }).character_name ?? "-");
  return JSON.stringify(value);
}

function labels(language: "zh-CN" | "en-US") {
  if (language === "zh-CN") {
    return {
      weekLabel: "周赛周期",
      phaseLabel: "当前阶段",
      activeContestants: "活跃积分选手",
      activeContestantsNote: "本周已进入积分赛状态的 Bot 数量",
      highestRating: "最高积分",
      highestRatingNote: "当前榜首积分",
      knockoutField: "周六 64 强",
      knockoutFieldNote: "周六淘汰赛的标准主赛人数",
      weekDesk: "本周竞技场概览",
      nextResolve: "下一阶段时间",
      nextResolveNote: "用于观察本周赛程推进",
      npcFillNote: (npcCount: number) => `当前补位 NPC ${npcCount}`,
      recentSignals: "近期竞技场信号",
      ratingTab: "积分赛排行榜",
      knockoutTab: "淘汰赛",
      ratingLadder: "本周积分赛排行榜",
      ratingLadderHint: "这里展示当前积分赛前列选手，周六将以积分排名确定 64 强种子。",
      activeContestantsHint: "周一到周五可进行积分挑战",
      highestRatingHint: "榜首正在冲击更高种子位",
      lowestFeaturedRating: "榜单观察低位",
      lowestFeaturedHint: "当前榜单展示范围内的最低积分",
      podiumChampion: "榜首",
      podiumRunnerUp: "次席",
      podiumThird: "第三",
      cutlineBannerLabel: "周赛晋级规则",
      cutlineBannerValue: "前 64 名进入周六淘汰赛",
      cutlineBannerNote: "当前榜单用于观察谁正在争夺种子位与 64 强安全区。",
      rank: "排名",
      competitor: "选手",
      ratingShort: "积分",
      powerShort: "战力",
      noRatingData: "当前还没有可展示的积分榜数据。",
      selectionNotes: "积分赛说明",
      top64RuleTitle: "64 强资格",
      top64RuleBody: "周五积分结算后，前 64 名将成为周六淘汰赛种子选手，不足名额由 NPC 补位。",
      challengeWindowTitle: "挑战窗口",
      challengeWindowBody: "积分挑战在周一到周五开放，Bot 会围绕接近积分段的对手进行排名争夺。",
      cutlineTitle: "榜单阅读",
      cutlineBody: "榜首积分适合观察夺冠热门，前列榜单则更适合观察谁正在逼近周六主赛区。",
      knockoutDesk: "周六淘汰赛",
      roundSwitchHint: "使用小标签切换 64 强到决赛各轮对局，点击对局可查看战斗详情。",
      knockoutNotes: "淘汰赛说明",
      seedSourceTitle: "种子来源",
      seedSourceBody: "淘汰赛名单来自本周积分赛排名，因此积分和总战力需要同时看。",
      powerViewTitle: "信息重点",
      powerViewBody: "淘汰赛卡片更适合观察对阵关系与总战力差距，而不是继续当作积分榜阅读。",
      reportViewTitle: "战斗报告",
      reportViewBody: "已结算对局可以打开战斗报告，适合对比强者稳定性和冷门翻盘情况。",
      championDesk: "冠军视角",
      power: "总战力",
      openBot: "打开角色详情",
      matchDetail: "对局详情",
      close: "关闭",
      combatants: "参赛角色",
      snapshot: "对局快照",
      status: "赛事状态",
      startTime: "开始时间",
      resolveTime: "结算时间",
      winner: "获胜者",
      pending: "待结算",
      battleReport: "战斗报告",
      loading: "正在读取战斗报告…",
      noReport: "当前没有可展示的战斗报告。",
      pendingMatch: "该对局尚未完成，稍后可以查看战报。",
    };
  }
  return {
    weekLabel: "Weekly Cycle",
    phaseLabel: "Current Phase",
    activeContestants: "Active Rated Bots",
    activeContestantsNote: "Bots currently participating in this week's rating race",
    highestRating: "Highest Rating",
    highestRatingNote: "Current first-place rating",
    knockoutField: "Saturday Top 64",
    knockoutFieldNote: "Standard field size for the weekly knockout",
    weekDesk: "Weekly Arena Snapshot",
    nextResolve: "Next Milestone",
    nextResolveNote: "Used to read this week's arena pacing",
    npcFillNote: (npcCount: number) => `NPC fill count ${npcCount}`,
    recentSignals: "Recent Arena Signals",
    ratingTab: "Rating Ladder",
    knockoutTab: "Knockout",
    ratingLadder: "Weekly Rating Ladder",
    ratingLadderHint: "This board tracks the visible leaders of the weekly rating season. Saturday seeds are locked from rating order.",
    activeContestantsHint: "Rating challenges run Monday through Friday",
    highestRatingHint: "The lead seed is still climbing",
    lowestFeaturedRating: "Lowest Featured Rating",
    lowestFeaturedHint: "Lowest score currently visible in the featured ladder slice",
    podiumChampion: "Leader",
    podiumRunnerUp: "Second",
    podiumThird: "Third",
    cutlineBannerLabel: "Weekly Qualification Rule",
    cutlineBannerValue: "Top 64 advance to Saturday knockout",
    cutlineBannerNote: "Use the live ladder to track seed favorites and the edge of the qualification zone.",
    rank: "Rank",
    competitor: "Competitor",
    ratingShort: "Rating",
    powerShort: "Power",
    noRatingData: "No rating ladder data is available yet.",
    selectionNotes: "Rating Notes",
    top64RuleTitle: "Top 64 Qualification",
    top64RuleBody: "After Friday closes, the top 64 ratings become Saturday knockout seeds. NPCs fill any remaining slots.",
    challengeWindowTitle: "Challenge Window",
    challengeWindowBody: "Rating challenges are open Monday through Friday, with bots contesting nearby opponents on the ladder.",
    cutlineTitle: "How to Read the Ladder",
    cutlineBody: "The top rating shows title favorites, while the visible ladder slice is best used to track who is pushing toward Saturday's field.",
    knockoutDesk: "Saturday Knockout",
    roundSwitchHint: "Use the small tabs to switch between Top 64 through Final matchups. Click any matchup for battle detail.",
    knockoutNotes: "Knockout Notes",
    seedSourceTitle: "Seed Source",
    seedSourceBody: "The weekly knockout field comes from the rating ladder, so rating and total power should be read together.",
    powerViewTitle: "Viewing Priority",
    powerViewBody: "Knockout cards are best for reading pairings and power gaps, not as a replacement for the ladder.",
    reportViewTitle: "Battle Reports",
    reportViewBody: "Resolved matches expose battle reports so you can compare stability, upsets, and closing strength.",
    championDesk: "Champion Desk",
    power: "Power",
    openBot: "Open bot profile",
    matchDetail: "Match Detail",
    close: "Close",
    combatants: "Combatants",
    snapshot: "Match Snapshot",
    status: "Status",
    startTime: "Started",
    resolveTime: "Resolved",
    winner: "Winner",
    pending: "Pending",
    battleReport: "Battle Report",
    loading: "Loading battle report…",
    noReport: "No battle report available.",
    pendingMatch: "This matchup has not resolved yet.",
  };
}
