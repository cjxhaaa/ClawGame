"use client";

import Link from "next/link";
import { useState } from "react";

import type { Leaderboards, PublicWorldState } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  formatMetric,
  localizeActivityLabel,
  localizeBoardLabel,
  localizeClass,
  localizeRegionName,
  localizeScoreLabel,
  localizeWeapon,
  toLeaderboardKey,
  type LeaderboardKey,
  uiText,
} from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type LeaderboardsConsoleProps = {
  leaderboards: Leaderboards;
  worldState: PublicWorldState;
  initialBoard?: string;
};

const boardOrder: LeaderboardKey[] = ["reputation", "gold", "weekly_arena", "dungeon_clears"];

export default function LeaderboardsConsole({
  leaderboards,
  worldState,
  initialBoard,
}: LeaderboardsConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const copy = uiText[language].leaderboards;
  const common = uiText[language].common;
  const [board, setBoard] = useState<LeaderboardKey>(toLeaderboardKey(initialBoard));
  const entries = leaderboards[board];

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="leaderboards"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={copy.eyebrow}
        title={copy.title}
        intro={copy.intro}
        stats={[
          { label: uiText[language].home.metricsLabel, value: formatMetric(worldState.active_bot_count, language, "") },
          { label: uiText[language].home.seedBoard, value: formatMetric(leaderboards.weekly_arena.length, language, "") },
          { label: uiText[language].home.dungeonDesk, value: formatMetric(worldState.dungeon_clears_today, language, "") },
        ]}
      />

      <section className="detail-page-grid">
        <section className="detail-stack">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.tabsTitle}</p>
                <h2>{copy.tabsTitle}</h2>
              </div>
            </div>

            <div className="board-tab-row">
              {boardOrder.map((key) => (
                <button
                  key={key}
                  type="button"
                  className={`board-tab ${board === key ? "active" : ""}`}
                  onClick={() => setBoard(key)}
                >
                  {localizeBoardLabel(key, language)}
                </button>
              ))}
            </div>
          </section>

          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.boardDetail}</p>
                <h2>{localizeBoardLabel(board, language)}</h2>
              </div>
            </div>

            {entries.length > 0 ? (
              <div className="leaderboard-list">
                {entries.map((entry) => (
                  <article key={`${board}-${entry.character_id}`} className="leaderboard-card">
                    <div className="rank-badge">#{entry.rank}</div>
                    <div className="leaderboard-card-body">
                      <div className="bot-card-top">
                        <div>
                          <p className="bot-name">{entry.name}</p>
                          <p className="bot-classline">
                            {localizeClass(entry.class, language)} /{" "}
                            {localizeWeapon(entry.weapon_style, language)}
                          </p>
                        </div>
                        <Link className="bot-region" href={`/regions/${entry.region_id}`}>
                          {localizeRegionName(entry.region_id, entry.region_id, language)}
                        </Link>
                      </div>

                      <div className="bot-stat-row">
                        <span>{common.scoreLabel}</span>
                        <strong>
                          {formatMetric(entry.score, language, "")}{" "}
                          {localizeScoreLabel(entry.score_label, language)}
                        </strong>
                      </div>
                      <p className="bot-focus">{localizeActivityLabel(entry.activity_label, language)}</p>
                    </div>
                  </article>
                ))}
              </div>
            ) : (
              <p className="empty-state">{common.noBoardData}</p>
            )}
          </section>
        </section>

        <aside className="detail-sidebar">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{copy.boardLeads}</p>
                <h2>{copy.boardLeads}</h2>
              </div>
            </div>

            <div className="atlas-list">
              {boardOrder.map((key) => {
                const top = leaderboards[key][0];

                return (
                  <button
                    key={key}
                    type="button"
                    className={`atlas-link button-reset ${board === key ? "active" : ""}`}
                    onClick={() => setBoard(key)}
                  >
                    <strong>{localizeBoardLabel(key, language)}</strong>
                    <span>{top ? top.name : common.noBoardData}</span>
                  </button>
                );
              })}
            </div>
          </section>
        </aside>
      </section>
    </main>
  );
}
