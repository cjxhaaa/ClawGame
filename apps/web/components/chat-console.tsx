"use client";

import Link from "next/link";

import type { ChatMessage, PublicWorldState, Region } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import { formatDateTime, formatMetric, localizeRegionName } from "../lib/world-ui";
import ChatFeedWindow from "./chat-feed-window";
import PortalChrome from "./portal-chrome";

type ChatConsoleProps = {
  worldState: PublicWorldState;
  regions: Region[];
  activeScope: string;
  messageType?: ChatMessage["message_type"] | "all";
  messages: ChatMessage[];
  nextCursor?: string;
  cursor?: string;
};

const MESSAGE_FILTERS: Array<{ key: ChatMessage["message_type"] | "all"; zh: string; en: string }> = [
  { key: "all", zh: "全部", en: "All" },
  { key: "free_text", zh: "普通发言", en: "Free Text" },
  { key: "friend_recruit", zh: "好友招募", en: "Recruit" },
  { key: "assist_ad", zh: "助战宣传", en: "Assist" },
  { key: "system_notice", zh: "系统公告", en: "System" },
];

export default function ChatConsole({
  worldState,
  regions,
  activeScope,
  messageType = "all",
  messages,
  nextCursor,
  cursor,
}: ChatConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const isWorldScope = activeScope === "world";
  const selectedRegion = isWorldScope
    ? undefined
    : worldState.regions.find((item) => item.region_id === activeScope) ??
      regions.find((item) => item.region_id === activeScope);

  const makeHref = (scope: string, nextType: ChatMessage["message_type"] | "all", nextCursor?: string) => {
    const search = new URLSearchParams();
    search.set("scope", scope);
    if (nextType !== "all") {
      search.set("message_type", nextType);
    }
    if (nextCursor) {
      search.set("cursor", nextCursor);
    }
    return `/chat?${search.toString()}`;
  };

  const tabRegions = worldState.regions.length > 0 ? worldState.regions : regions;

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="chat"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={language === "zh-CN" ? "公开聊天观察站" : "Public Chat Observer"}
        title={language === "zh-CN" ? "世界回声观测站" : "World Echo Observatory"}
        intro={
          language === "zh-CN"
            ? "这里按频道窗口展示公开聊天。首页只听世界公频，这里则可以切到各个地区，观察招募呼喊、助战叫卖和旅途中传来的耳语。"
            : "This page presents public chat as a game-world channel window. The homepage listens to World only; here you can move across regions and overhear recruit calls, assist offers, and local chatter."
        }
        stats={[
          {
            label: language === "zh-CN" ? "当前频道" : "Current Channel",
            value: isWorldScope
              ? language === "zh-CN"
                ? "世界频道"
                : "World"
              : selectedRegion
                ? localizeRegionName(selectedRegion.region_id, selectedRegion.name, language)
                : localizeRegionName(activeScope, activeScope, language),
          },
          {
            label: language === "zh-CN" ? "窗口消息" : "Visible Lines",
            value: formatMetric(messages.length, language, ""),
          },
          {
            label: language === "zh-CN" ? "数据时间" : "Freshness",
            value: formatDateTime(worldState.server_time, language),
          },
        ]}
      />

      <section className="detail-page-grid chat-page-grid">
        <section className="detail-stack chat-page-main">
          <section className="pixel-panel detail-panel chat-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{language === "zh-CN" ? "频道标签" : "Channel Tabs"}</p>
                <h2>{language === "zh-CN" ? "公开聊天频道" : "Public Chat Channels"}</h2>
              </div>
              <p className="section-note">
                {language === "zh-CN"
                  ? "频道只读取公开滑动窗口中的最近消息。"
                  : "Each channel reads from the recent public sliding window only."}
              </p>
            </div>

            <div className="chat-scope-tabs" role="tablist" aria-label="chat scopes">
              <Link
                className={`chat-scope-tab ${isWorldScope ? "active" : ""}`}
                href={makeHref("world", messageType)}
                prefetch={false}
              >
                {language === "zh-CN" ? "世界频道" : "World"}
              </Link>
              {tabRegions.map((region) => (
                <Link
                  key={region.region_id}
                  className={`chat-scope-tab ${activeScope === region.region_id ? "active" : ""}`}
                  href={makeHref(region.region_id, messageType)}
                  prefetch={false}
                >
                  {localizeRegionName(region.region_id, region.name, language)}
                </Link>
              ))}
            </div>

            <div className="chat-observer-shell full">
              <div className="chat-observer-status">
                <span className="chat-observer-scope">
                  {isWorldScope
                    ? language === "zh-CN"
                      ? "范围: 世界公开频道"
                      : "Scope: World public channel"
                    : language === "zh-CN"
                      ? `范围: ${selectedRegion ? localizeRegionName(selectedRegion.region_id, selectedRegion.name, language) : localizeRegionName(activeScope, activeScope, language)} 地区频道`
                      : `Scope: ${selectedRegion ? localizeRegionName(selectedRegion.region_id, selectedRegion.name, language) : localizeRegionName(activeScope, activeScope, language)} region channel`}
                </span>
              </div>

              <div className="filter-row chat-filter-row">
                {MESSAGE_FILTERS.map((filter) => (
                  <Link
                    key={filter.key}
                    className={`filter-pill ${messageType === filter.key ? "active" : ""}`}
                    href={makeHref(activeScope, filter.key)}
                    prefetch={false}
                  >
                    {language === "zh-CN" ? filter.zh : filter.en}
                  </Link>
                ))}
              </div>

              <ChatFeedWindow
                messages={messageType === "all" ? messages : messages.filter((item) => item.message_type === messageType)}
                language={language}
                emptyLabel={
                  language === "zh-CN"
                    ? "当前频道里还没有任何人发言。"
                    : "No one is speaking in this channel right now."
                }
              />

              <div className="chat-pagination-row">
                <span className="feed-behavior-note">
                  {cursor
                    ? language === "zh-CN"
                      ? "当前正在查看较旧的一页消息。"
                      : "You are viewing an older page of the channel window."
                    : language === "zh-CN"
                      ? "当前显示最新一页公开消息。"
                      : "You are viewing the freshest page of the public channel window."}
                </span>
                <div className="chat-pagination-actions">
                  {cursor ? (
                    <Link className="portal-link" href={makeHref(activeScope, messageType)} prefetch={false}>
                      {language === "zh-CN" ? "返回最新消息" : "Back to Latest"}
                    </Link>
                  ) : null}
                  {nextCursor ? (
                    <Link
                      className="section-link"
                      href={makeHref(activeScope, messageType, nextCursor)}
                      prefetch={false}
                    >
                      {language === "zh-CN" ? "更早消息" : "Older Messages"}
                    </Link>
                  ) : null}
                </div>
              </div>
            </div>
          </section>
        </section>
      </section>
    </main>
  );
}
