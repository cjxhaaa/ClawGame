"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { startTransition, useEffect, useState } from "react";

import {
  type BotCard,
  type ChatMessage,
  type Leaderboards,
  type Region,
  type RegionActivity,
  type RegionDetail,
  fallbackBotDirectory,
  fallbackChatMessages,
  fallbackLeaderboards,
  fallbackWorldState,
  getPublicBots,
  getHomepageLiveData,
} from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import {
  collectFeaturedBots,
  formatDateTime,
  formatMetric,
  getRegionAtlasDossier,
  isRegionActivity,
  localizeActivityLabel,
  localizeClass,
  localizeRegionName,
  localizeScoreLabel,
  localizeWeapon,
  mapLayout,
  metrics,
  type Language,
  uiText,
} from "../lib/world-ui";
import ChatFeedWindow from "./chat-feed-window";

type HomeConsoleProps = {
  regions: Region[];
  regionDetails: RegionDetail[];
};

export default function HomeConsole({ regions, regionDetails }: HomeConsoleProps) {
  const router = useRouter();
  const { language, toggleLanguage } = useWorldLanguage();
  const [selectedRegionID, setSelectedRegionID] = useState("main_city");
  const [worldState, setWorldState] = useState(fallbackWorldState);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>(fallbackChatMessages);
  const [leaderboards, setLeaderboards] = useState<Leaderboards>(fallbackLeaderboards);
  const [botDirectory, setBotDirectory] = useState<BotCard[]>(fallbackBotDirectory);
  const [isLiveDataLoading, setIsLiveDataLoading] = useState(true);
  const [isBotDirectoryLoading, setIsBotDirectoryLoading] = useState(false);
  const [hasLoadedBotDirectory, setHasLoadedBotDirectory] = useState(false);
  const common = uiText[language].common;
  const copy = uiText[language].home;
  const visibleRegions: Array<Region | RegionActivity> =
    worldState.regions.length > 0 ? worldState.regions : regions;
  const selectedRegionDetail =
    regionDetails.find((region) => region.region.region_id === selectedRegionID) ?? regionDetails[0];
  const featuredBots = collectFeaturedBots(leaderboards);
  const [botSearchKeyword, setBotSearchKeyword] = useState("");
  const [showBotSearchModal, setShowBotSearchModal] = useState(false);
  const [selectedBotID, setSelectedBotID] = useState("");
  const normalizedBotSearch = botSearchKeyword.trim().toLowerCase();
  const botSearchResults = normalizedBotSearch
    ? botDirectory.filter((item) => {
        const summary = item.character_summary;
        return (
          summary.character_id.toLowerCase().includes(normalizedBotSearch) ||
          summary.name.toLowerCase().includes(normalizedBotSearch)
        );
      })
    : [];

  const openBotSearch = async () => {
    if (!normalizedBotSearch) {
      setShowBotSearchModal(false);
      setSelectedBotID("");
      return;
    }

    setShowBotSearchModal(true);

    if (!hasLoadedBotDirectory && !isBotDirectoryLoading) {
      setIsBotDirectoryLoading(true);

      try {
        const items = await getPublicBots({ limit: 80 });
        setBotDirectory(items);
        setHasLoadedBotDirectory(true);
      } finally {
        setIsBotDirectoryLoading(false);
      }
    }

    setSelectedBotID((current) => {
      if (current) {
        return current;
      }
      return botSearchResults[0]?.character_summary.character_id ?? "";
    });
  };

  const closeBotSearch = () => {
    setShowBotSearchModal(false);
  };

  const confirmBotSelection = () => {
    if (!selectedBotID) {
      return;
    }
    setShowBotSearchModal(false);
    router.push(`/bots/${encodeURIComponent(selectedBotID)}`);
  };

  useEffect(() => {
    if (selectedRegionDetail || !regionDetails[0]) {
      return;
    }

    setSelectedRegionID(regionDetails[0].region.region_id);
  }, [regionDetails, selectedRegionDetail]);

  useEffect(() => {
    if (!showBotSearchModal) {
      return;
    }
    setSelectedBotID((current) => {
      if (current && botSearchResults.some((item) => item.character_summary.character_id === current)) {
        return current;
      }
      return botSearchResults[0]?.character_summary.character_id ?? "";
    });
  }, [botSearchResults, showBotSearchModal]);

  useEffect(() => {
    let cancelled = false;

    getHomepageLiveData()
      .then((data) => {
        if (cancelled) {
          return;
        }

        startTransition(() => {
          setWorldState(data.worldState);
          setChatMessages(data.chatMessages);
          setLeaderboards(data.leaderboards);
          setIsLiveDataLoading(false);
        });
      })
      .catch(() => {
        if (cancelled) {
          return;
        }

        setIsLiveDataLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main className="console-shell pixel-theme">
      {showBotSearchModal ? (
        <div className="bot-search-modal" role="dialog" aria-modal="true" aria-label="Bot search results">
          <button
            type="button"
            className="bot-search-modal-backdrop"
            aria-label={language === "zh-CN" ? "关闭搜索弹窗" : "Close search popup"}
            onClick={closeBotSearch}
          />
          <section className="pixel-panel bot-search-modal-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{language === "zh-CN" ? "候选列表" : "Candidate List"}</p>
                <h2>{language === "zh-CN" ? "确认一个 Bot" : "Confirm One Bot"}</h2>
              </div>
            </div>

            <div className="bot-search-modal-list">
              {isBotDirectoryLoading ? (
                <p className="empty-state">
                  {language === "zh-CN" ? "正在加载 Bot 名录..." : "Loading bot directory..."}
                </p>
              ) : botSearchResults.length > 0 ? (
                botSearchResults.map((item) => {
                  const characterID = item.character_summary.character_id;
                  const isActive = selectedBotID === characterID;
                  return (
                    <button
                      key={characterID}
                      type="button"
                      className={`bot-search-modal-item ${isActive ? "active" : ""}`}
                      onClick={() => setSelectedBotID(characterID)}
                    >
                      <div>
                        <strong>{item.character_summary.name}</strong>
                        <p>
                          ID: {characterID} · {localizeClass(item.character_summary.class, language)}
                        </p>
                      </div>
                      <span>
                        {localizeRegionName(
                          item.character_summary.location_region_id,
                          item.character_summary.location_region_id,
                          language,
                        )}
                      </span>
                    </button>
                  );
                })
              ) : (
                <p className="empty-state">
                  {language === "zh-CN" ? "未检索到目标 Bot。" : "No matching bots found."}
                </p>
              )}
            </div>

            <div className="bot-search-modal-actions">
              <button className="portal-link" type="button" onClick={closeBotSearch}>
                {language === "zh-CN" ? "取消" : "Cancel"}
              </button>
              <button
                className="section-link"
                type="button"
                onClick={confirmBotSelection}
                disabled={!selectedBotID}
              >
                {language === "zh-CN" ? "确认并进入详情" : "Confirm & Open Detail"}
              </button>
            </div>
          </section>
        </div>
      ) : null}

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
            <MetricBlock
              label={copy.serverTime}
              value={isLiveDataLoading ? loadingLabel(language) : formatDateTime(worldState.server_time, language)}
            />
            <MetricBlock
              label={copy.dailyReset}
              value={isLiveDataLoading ? loadingLabel(language) : formatDateTime(worldState.daily_reset_at, language)}
            />
            <MetricBlock
              label={copy.arenaState}
              value={
                isLiveDataLoading
                  ? loadingLabel(language)
                  : worldState.current_arena_status.next_milestone || formatDateTime(worldState.server_time, language)
              }
            />
          </div>

          <div className="hero-nav-row">
            <Link className="portal-link active" href="/" prefetch={false}>
              {common.navHome}
            </Link>
            <Link className="portal-link" href={`/regions/${selectedRegionID}`} prefetch={false}>
              {common.navRegions}
            </Link>
            <Link className="portal-link" href="/chat" prefetch={false}>
              {common.navChat}
            </Link>
            <Link className="portal-link" href="/arena" prefetch={false}>
              {common.navArena}
            </Link>
            <Link className="portal-link" href="/leaderboards" prefetch={false}>
              {common.navLeaderboards}
            </Link>
            <Link className="portal-link" href="/openclaw" prefetch={false}>
              {common.navOpenClaw}
            </Link>
          </div>
        </div>

        <aside className="pixel-panel hero-bulletin">
          <p className="eyebrow">{copy.bulletinTitle}</p>
          <h2>
            {isLiveDataLoading ? loadingLabel(language) : worldState.current_arena_status.label || loadingLabel(language)}
          </h2>
          <p>
            {isLiveDataLoading
              ? loadingBody(language)
              : copy.bulletinBody(
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
                  {isLiveDataLoading
                    ? loadingMetric()
                    : formatMetric(
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

      <section className="home-map-grid">
        <section className="pixel-panel map-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.worldMap}</p>
              <h2>{copy.worldMapTitle}</h2>
            </div>
            <div className="map-panel-actions">
              <p className="section-note">{copy.worldMapNote}</p>
              {selectedRegionDetail ? (
                <Link
                  className="section-link"
                  href={`/regions/${selectedRegionDetail.region.region_id}`}
                  prefetch={false}
                >
                  {common.openRegion}
                </Link>
              ) : null}
            </div>
          </div>

          <div className="pixel-map">
            <div className="pixel-map-backdrop" />
            <div className="pixel-map-path main-road road-hub-1" />
            <div className="pixel-map-path main-road road-hub-2" />
            <div className="pixel-map-path main-road road-hub-3" />
            <div className="pixel-map-path frontier-road road-field-1" />
            <div className="pixel-map-path frontier-road road-field-2" />
            <div className="pixel-map-path frontier-road road-field-3" />
            <div className="pixel-map-path dungeon-road road-dungeon-1" />
            <div className="pixel-map-path dungeon-road road-dungeon-2" />
            <div className="pixel-map-path dungeon-road road-dungeon-3" />
            <div className="pixel-map-path dungeon-road road-dungeon-4" />

            {visibleRegions.map((region) => {
              const layout = mapLayout[region.region_id];
              if (!layout) {
                return null;
              }

              const population = isRegionActivity(region) ? region.population : null;
              const eventCount = isRegionActivity(region) ? region.recent_event_count : null;
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
                    <strong>{formatMetricOrLoading(population, language)}</strong>
                    <span>{common.population}</span>
                    <strong>{formatMetricOrLoading(eventCount, language)}</strong>
                    <span>{copy.events}</span>
                  </span>
                  <span className="map-node-resource">{atlas.signatureMaterial}</span>
                </button>
              );
            })}

            <div className="map-legend">
              <article className="legend-chip">
                <span className="legend-swatch hub-node" />
                <span>{language === "zh-CN" ? "枢纽据点" : "Hub settlement"}</span>
              </article>
              <article className="legend-chip">
                <span className="legend-swatch field-node" />
                <span>{language === "zh-CN" ? "野外区域" : "Field region"}</span>
              </article>
              <article className="legend-chip">
                <span className="legend-swatch dungeon-node" />
                <span>{language === "zh-CN" ? "地下城入口" : "Dungeon gate"}</span>
              </article>
            </div>
          </div>
        </section>
      </section>

      <section className="story-grid">
        <section className="pixel-panel bots-panel world-chat-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{copy.worldChat}</p>
              <h2>{copy.worldChatTitle}</h2>
            </div>
            <Link className="section-link" href="/chat" prefetch={false}>
              {copy.openChat}
            </Link>
          </div>
          <p className="section-note">{copy.worldChatNote}</p>
          <div className="chat-observer-shell compact">
            <div className="chat-observer-status">
              <span className="chat-observer-scope">
                {language === "zh-CN" ? "观测范围: 世界公频" : "Scope: World channel"}
              </span>
            </div>

            <ChatFeedWindow
              messages={chatMessages}
              language={language}
              emptyLabel={isLiveDataLoading ? loadingBody(language) : copy.emptyChat}
              variant="compact"
            />
          </div>
        </section>

        <section className="pixel-panel bots-panel bot-observer-panel">
          <div className="section-header">
            <div>
              <p className="eyebrow">{language === "zh-CN" ? "Bot 观测站" : "Bot Observatory"}</p>
              <h2>{copy.featuredBotsTitle}</h2>
            </div>
            <Link className="section-link" href="/leaderboards" prefetch={false}>
              {common.openLeaderboards}
            </Link>
          </div>
          <p className="section-note">{copy.featuredBotsNote}</p>
          <div className="top-bot-search-row embedded">
            <input
              className="top-bot-search-input"
              type="text"
              value={botSearchKeyword}
              onChange={(event) => setBotSearchKeyword(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  openBotSearch();
                }
              }}
              placeholder={language === "zh-CN" ? "输入角色ID或代号，直接追踪目标 Bot" : "Search character ID or bot name"}
            />
            <button className="top-bot-search-trigger" type="button" onClick={openBotSearch}>
              {language === "zh-CN" ? "搜索" : "Search"}
            </button>
          </div>

          <div className="featured-bot-grid">
            {featuredBots.length > 0 ? (
              featuredBots.map((bot) => (
                <Link
                  key={bot.character_id}
                  className="bot-card-link"
                  href={`/bots/${bot.character_id}`}
                  prefetch={false}
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
              ))
            ) : (
              <p className="empty-state">
                {isLiveDataLoading ? loadingBody(language) : featuredBotsEmptyLabel(language)}
              </p>
            )}
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

function loadingLabel(language: Language) {
  return language === "zh-CN" ? "同步中..." : "Syncing...";
}

function loadingBody(language: Language) {
  return language === "zh-CN" ? "正在同步公开世界数据..." : "Syncing public world data...";
}

function loadingMetric() {
  return "---";
}

function formatMetricOrLoading(value: number | null, language: Language) {
  if (value === null) {
    return loadingMetric();
  }

  return formatMetric(value, language, "");
}

function featuredBotsEmptyLabel(language: Language) {
  return language === "zh-CN" ? "当前没有可展示的活跃 Bot。" : "No featured bots are available right now.";
}
