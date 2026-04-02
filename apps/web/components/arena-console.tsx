"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import type { ArenaCurrent, ArenaEntry, ArenaMatchDetail, ArenaMatchup, ArenaRound, Leaderboards, PublicWorldState, WorldEvent } from "../lib/public-api";
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

export default function ArenaConsole({ worldState, events, arenaCurrent }: ArenaConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const copy = uiText[language].arena;
  const common = uiText[language].common;
  const home = uiText[language].home;
  const text = labels(language);
  const arenaStatus = localizeArenaStatus(worldState.current_arena_status.code, language);
  const arenaEvents = events.filter((event) => event.event_type.startsWith("arena"));
  const featuredEntries = arenaCurrent.featured_entries ?? [];
  const qualifierRounds = arenaCurrent.qualifier_rounds ?? [];
  const mainRounds = arenaCurrent.rounds ?? [];
  const stageRounds = useMemo(() => mainRounds.filter((round) => round.matchups.length > 0), [mainRounds]);
  const [selectedMatch, setSelectedMatch] = useState<SelectedMatch | null>(null);
  const [selectedDetail, setSelectedDetail] = useState<ArenaMatchDetail | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [selectedRoundName, setSelectedRoundName] = useState<string>(stageRounds[0]?.name ?? "");

  useEffect(() => {
    if (stageRounds.length === 0) {
      setSelectedRoundName("");
      return;
    }
    if (!stageRounds.some((round) => round.name === selectedRoundName)) {
      setSelectedRoundName(stageRounds[0].name);
    }
  }, [stageRounds, selectedRoundName]);

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
          { label: text.totalEntrants, value: formatMetric(arenaCurrent.signup_count, language, "") },
          { label: home.arenaNext, value: arenaStatus.nextMilestone },
        ]}
      />

      <section className="arena-overview-strip">
        <SummaryMetric label={text.totalEntrants} value={formatMetric(arenaCurrent.signup_count, language, "")} note={text.todayNote} />
        <SummaryMetric label={text.highestPower} value={formatMetric(arenaCurrent.highest_panel_power, language, "")} note={text.highestNote} />
        <SummaryMetric label={text.lowestPower} value={formatMetric(arenaCurrent.lowest_panel_power, language, "")} note={text.lowestNote} />
        <SummaryMetric label={text.medianPower} value={formatMetric(arenaCurrent.median_panel_power, language, "")} note={text.medianNote} />
        <SummaryMetric label={text.top64} value={formatMetric(arenaCurrent.qualified_count, language, "")} note={text.top64Note} />
        <SummaryMetric label={text.currentStage} value={activeStage(arenaCurrent, language)} note={text.currentStageNote} />
      </section>

      <section className="arena-brief-grid">
        <section className="pixel-panel detail-panel arena-brief-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{text.todayDesk}</p>
              <h2>{text.todayDesk}</h2>
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
              <p>{text.dragHint}</p>
            </article>
            <article className="arena-summary-card">
              <span>{text.npcFill}</span>
              <strong>{formatMetric(arenaCurrent.npc_count, language, "")}</strong>
              <p>{text.qualifierInfo(arenaCurrent.qualifier_rounds?.length ?? 0)}</p>
            </article>
          </div>

          <div className="arena-entrant-ribbon">
            {featuredEntries.map((entry, index) => (
              <Link key={entry.character_id} href={`/bots/${encodeURIComponent(entry.character_id)}`} className="arena-entrant-pill">
                <span>#{index + 1}</span>
                <strong>{entry.character_name}</strong>
                <small>{formatMetric(entry.panel_power_score, language, "")}</small>
              </Link>
            ))}
          </div>
        </section>

        <section className="pixel-panel detail-panel arena-brief-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{text.qualifierTrace}</p>
              <h2>{text.qualifierTrace}</h2>
            </div>
          </div>
          {qualifierRounds.length > 0 ? (
            <div className="arena-round-track compact">
              {qualifierRounds.map((round) => (
                <article key={`${round.stage}-${round.round_number}`} className={`arena-round-node ${round.status}`}>
                  <span className="arena-round-node-step">{formatArenaClock(round.scheduled_at, language)}</span>
                  <strong>{round.name}</strong>
                  <span>{localizeRoundStatus(round.status, language)}</span>
                  <small>{formatMetric(round.entrant_count, language, "")}</small>
                </article>
              ))}
            </div>
          ) : (
            <p className="empty-state">{text.noQualifier}</p>
          )}
        </section>
      </section>

      <section className="detail-page-grid arena-redesign-grid">
        <section className="pixel-panel detail-panel arena-bracket-shell">
          <div className="section-header">
            <div>
              <p className="eyebrow">{text.stageRounds}</p>
              <h2>{text.stageRounds}</h2>
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

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{text.signals}</p>
                <h2>{text.signals}</h2>
              </div>
            </div>
            <div className="log-list">
              {arenaEvents.length > 0 ? (
                arenaEvents.slice(0, 10).map((event) => (
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
        </aside>
      </section>

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
                        <strong>{reportValue(value)}</strong>
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
  return <article className="arena-metric-panel"><span>{label}</span><strong>{value}</strong><p>{note}</p></article>;
}

function BracketMatchCard({ matchup, roundName, language, onOpen, finalNode = false }: { matchup: ArenaMatchup; roundName: string; language: "zh-CN" | "en-US"; onOpen: (matchup: ArenaMatchup, roundName: string) => void; finalNode?: boolean }) {
  return (
    <button type="button" className={`arena-bracket-match ${matchup.status} ${finalNode ? "final-node" : ""}`} onClick={() => onOpen(matchup, roundName)}>
      <div className="arena-bracket-match-head"><span>{roundName}</span><strong>#{matchup.match_number}</strong></div>
      <div className="arena-bracket-combatant-list">
        <BracketEntry entry={matchup.left_entry} language={language} winnerID={matchup.winner_entry?.character_id} />
        <BracketEntry entry={matchup.right_entry} language={language} winnerID={matchup.winner_entry?.character_id} />
        {matchup.bye_entry ? <BracketEntry entry={matchup.bye_entry} language={language} winnerID={matchup.bye_entry.character_id} bye /> : null}
      </div>
      <div className="arena-bracket-match-foot"><span>{localizeMatchStatus(matchup.status, language)}</span>{matchup.battle_report_id ? <small>{language === "zh-CN" ? "可查看战报" : "Report ready"}</small> : null}</div>
    </button>
  );
}

function BracketEntry({ entry, language, winnerID, bye = false }: { entry?: ArenaEntry; language: "zh-CN" | "en-US"; winnerID?: string; bye?: boolean }) {
  if (!entry) return <div className="arena-bracket-entry ghost">{language === "zh-CN" ? "等待对手" : "Pending entrant"}</div>;
  return (
    <div className={`arena-bracket-entry ${winnerID === entry.character_id ? "winner" : ""} ${entry.is_npc ? "npc" : ""}`}>
      <div>
        <Link href={`/bots/${encodeURIComponent(entry.character_id)}`} onClick={(event) => event.stopPropagation()}>{entry.character_name}</Link>
        <small>{localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}</small>
      </div>
      <span>{bye ? (language === "zh-CN" ? "轮空" : "Bye") : formatMetric(entry.panel_power_score, language, "")}</span>
    </div>
  );
}

function DetailCard({ entry, language }: { entry?: { character_id: string; character_name: string; class: string; weapon_style: string; panel_power_score: number; is_npc?: boolean }; language: "zh-CN" | "en-US" }) {
  if (!entry) return <div className="arena-detail-entry-card ghost">{language === "zh-CN" ? "暂无选手" : "No entrant"}</div>;
  return (
    <article className={`arena-detail-entry-card ${entry.is_npc ? "npc" : ""}`}>
      <strong><Link href={`/bots/${encodeURIComponent(entry.character_id)}`}>{entry.character_name}</Link></strong>
      <p>{localizeClass(entry.class, language)} / {localizeWeapon(entry.weapon_style, language)}</p>
      <div className="bot-stat-row"><span>{language === "zh-CN" ? "总战力" : "Power"}</span><strong>{formatMetric(entry.panel_power_score, language, "")}</strong></div>
    </article>
  );
}

function DetailStat({ label, value }: { label: string; value: string }) {
  return <article className="arena-match-detail-stat"><span>{label}</span><strong>{value}</strong></article>;
}

function activeStage(arenaCurrent: ArenaCurrent, language: "zh-CN" | "en-US") {
  const qualifier = arenaCurrent.qualifier_rounds?.find((round) => round.status === "in_progress");
  if (qualifier) return qualifier.name;
  const main = arenaCurrent.rounds.find((round) => round.status === "in_progress") ?? arenaCurrent.rounds.find((round) => round.status === "scheduled");
  if (main) return main.name;
  return language === "zh-CN" ? "等待下一轮" : "Waiting";
}

function localizeMatchStatus(status: string, language: "zh-CN" | "en-US") {
  const labels: Record<string, { "zh-CN": string; "en-US": string }> = { scheduled: { "zh-CN": "待开打", "en-US": "Scheduled" }, in_progress: { "zh-CN": "进行中", "en-US": "Live" }, resolved: { "zh-CN": "已结算", "en-US": "Resolved" }, walkover: { "zh-CN": "轮空晋级", "en-US": "Walkover" } };
  return labels[status]?.[language] ?? status;
}

function localizeRoundStatus(status: string, language: "zh-CN" | "en-US") {
  const labels: Record<string, { "zh-CN": string; "en-US": string }> = { scheduled: { "zh-CN": "待开始", "en-US": "Scheduled" }, in_progress: { "zh-CN": "进行中", "en-US": "In Progress" }, resolved: { "zh-CN": "已完成", "en-US": "Resolved" } };
  return labels[status]?.[language] ?? status;
}

function arenaTone(code: string) {
  if (code === "in_progress") return "live";
  if (code === "results_live") return "complete";
  if (code === "signup_locked") return "seeding";
  return "open";
}

function formatArenaClock(value: string, language: "zh-CN" | "en-US") {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(language, { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }).format(date);
}

function reportLabel(key: string, language: "zh-CN" | "en-US") {
  const labels: Record<string, { "zh-CN": string; "en-US": string }> = { outcome: { "zh-CN": "结果", "en-US": "Outcome" }, winner: { "zh-CN": "胜者", "en-US": "Winner" }, loser: { "zh-CN": "败者", "en-US": "Loser" }, winner_final_hp: { "zh-CN": "胜者剩余生命", "en-US": "Winner HP" }, power_delta: { "zh-CN": "战力差", "en-US": "Power Delta" }, player_power_delta: { "zh-CN": "战力差", "en-US": "Power Delta" }, battle_length_events: { "zh-CN": "事件数", "en-US": "Event Count" }, summary_tag: { "zh-CN": "摘要标签", "en-US": "Summary Tag" } };
  return labels[key]?.[language] ?? key;
}

function reportValue(value: unknown) {
  if (value === null || value === undefined) return "-";
  if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") return String(value);
  if (typeof value === "object" && value && "character_name" in value) return String((value as { character_name?: string }).character_name ?? "-");
  return JSON.stringify(value);
}

function labels(language: "zh-CN" | "en-US") {
  if (language === "zh-CN") {
    return {
      totalEntrants: "今日报名",
      todayNote: "当天成功进入竞技场池的 Bot 总数",
      highestPower: "最高战力",
      highestNote: "今日参赛者中的最高总战力",
      lowestPower: "最低战力",
      lowestNote: "今日参赛者中的最低总战力",
      medianPower: "中位战力",
      medianNote: "用于观察今天整体强度",
      top64: "进入 64 强",
      top64Note: "资格赛结束后进入主赛的人数",
      currentStage: "当前阶段",
      currentStageNote: "正在结算或即将开始的轮次",
      stageRounds: "主赛轮次切换",
      roundSwitchHint: "所有主赛轮次都可以通过小标签切换查看单轮对局。",
      todayDesk: "今日竞技场概览",
      nextResolve: "下一次结算",
      dragHint: "点击任意对局可以查看详细战斗信息。",
      npcFill: "NPC 补位",
      qualifierInfo: (count: number) => `资格赛轮数 ${count}`,
      qualifierTrace: "资格赛推进",
      noQualifier: "当前还没有资格赛数据。",
      mainBracket: "64 强主赛图谱",
      finalLabel: "总决赛",
      champion: "今日冠军",
      championDesk: "冠军视角",
      power: "总战力",
      openBot: "打开角色详情",
      signals: "近期竞技场信号",
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
    totalEntrants: "Today's Entrants",
    todayNote: "Bots that successfully entered today's arena pool",
    highestPower: "Highest Power",
    highestNote: "Top panel power among today's entrants",
    lowestPower: "Lowest Power",
    lowestNote: "Lowest panel power among today's entrants",
    medianPower: "Median Power",
    medianNote: "A quick read on today's field strength",
    top64: "Qualified to Top 64",
    top64Note: "Entrants that survived qualifiers into the main bracket",
    currentStage: "Current Stage",
    currentStageNote: "The round resolving now or starting next",
    stageRounds: "Main Bracket Rounds",
    roundSwitchHint: "Use the small tabs to switch every main bracket round as a single-round matchup view.",
    todayDesk: "Today's Arena Snapshot",
    nextResolve: "Next Resolution",
    dragHint: "Click any matchup to inspect its battle detail.",
    npcFill: "NPC Fill",
    qualifierInfo: (count: number) => `Qualifier rounds ${count}`,
    qualifierTrace: "Qualifier Progress",
    noQualifier: "No qualifier data yet.",
    mainBracket: "Top 64 Bracket Map",
    finalLabel: "Grand Final",
    champion: "Champion",
    championDesk: "Champion Desk",
    power: "Power",
    openBot: "Open bot profile",
    signals: "Recent Arena Signals",
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
