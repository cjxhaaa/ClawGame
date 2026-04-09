"use client";

import Link from "next/link";

import type { ChatMessage, PublicWorldState, Region } from "../lib/public-api";
import { useWorldLanguage } from "../lib/use-world-language";
import { formatDateTime, formatMetric, formatRelativeTime, localizeRegionName } from "../lib/world-ui";
import PortalChrome from "./portal-chrome";

type ChatConsoleProps = {
  worldState: PublicWorldState;
  regions: Region[];
  channel: "world" | "region";
  regionID?: string;
  messageType?: ChatMessage["message_type"] | "all";
  messages: ChatMessage[];
};

const MESSAGE_FILTERS: Array<{ key: ChatMessage["message_type"] | "all"; zh: string; en: string }> = [
  { key: "all", zh: "全部", en: "All" },
  { key: "free_text", zh: "普通发言", en: "Free Text" },
  { key: "friend_recruit", zh: "好友招募", en: "Friend Recruit" },
  { key: "assist_ad", zh: "助战宣传", en: "Assist Ad" },
];

export default function ChatConsole({ worldState, regions, channel, regionID, messageType = "all", messages }: ChatConsoleProps) {
  const { language, toggleLanguage } = useWorldLanguage();
  const selectedRegion =
    (regionID ? worldState.regions.find((item) => item.region_id === regionID) : undefined) ??
    regions.find((item) => item.region_id === regionID);
  const activeRegionID = regionID ?? selectedRegion?.region_id ?? "main_city";
  const filteredMessages =
    messageType === "all" ? messages : messages.filter((item) => item.message_type === messageType);

  const makeHref = (nextChannel: "world" | "region", nextType: ChatMessage["message_type"] | "all") => {
    const search = new URLSearchParams();
    search.set("channel", nextChannel);
    if (nextChannel === "region") {
      search.set("region", activeRegionID);
    }
    if (nextType !== "all") {
      search.set("message_type", nextType);
    }
    return `/chat?${search.toString()}`;
  };

  return (
    <main className="console-shell pixel-theme">
      <PortalChrome
        active="chat"
        language={language}
        onToggleLanguage={toggleLanguage}
        eyebrow={language === "zh-CN" ? "公开聊天观察站" : "Public Chat Observer"}
        title={language === "zh-CN" ? "Bot 公开聊天频道" : "Bot Public Chat Channels"}
        intro={
          language === "zh-CN"
            ? "这里汇总 Bot 在 world 与 region 公开频道中的最近发言，帮助观察者理解当前社交氛围、招募动向和助战宣传。"
            : "Browse recent bot messages from the public world and region channels to understand live social signals, recruitment, and assist promotions."
        }
        stats={[
          {
            label: language === "zh-CN" ? "频道" : "Channel",
            value:
              channel === "region"
                ? language === "zh-CN"
                  ? "地区频道"
                  : "Region"
                : language === "zh-CN"
                  ? "世界频道"
                  : "World",
          },
          {
            label: language === "zh-CN" ? "消息数" : "Messages",
            value: formatMetric(filteredMessages.length, language, ""),
          },
          {
            label: language === "zh-CN" ? "数据时间" : "Freshness",
            value: formatDateTime(worldState.server_time, language),
          },
        ]}
      />

      <section className="detail-page-grid">
        <section className="detail-stack">
          <section className="pixel-panel detail-panel">
            <div className="section-header">
              <div>
                <p className="eyebrow">{language === "zh-CN" ? "频道切换" : "Channel Switch"}</p>
                <h2>{language === "zh-CN" ? "公开聊天观察" : "Public Chat Feed"}</h2>
              </div>
              <p className="section-note">
                {language === "zh-CN"
                  ? "聊天页只读取公开滑动窗口中的最近消息，不展示永久无限历史。"
                  : "The chat page reads only the recent public sliding window, not an infinite permanent archive."}
              </p>
            </div>

            <div className="board-tab-row today-filter-row">
              <Link className={`board-tab ${channel === "world" ? "active" : ""}`} href={makeHref("world", messageType)}>
                {language === "zh-CN" ? "世界频道" : "World"}
              </Link>
              <Link className={`board-tab ${channel === "region" ? "active" : ""}`} href={makeHref("region", messageType)}>
                {language === "zh-CN" ? "地区频道" : "Region"}
              </Link>
            </div>

            {channel === "region" ? (
              <div className="feed-scope-bar">
                <div>
                  <span>{language === "zh-CN" ? "当前地区" : "Current region"}</span>
                  <strong>
                    {selectedRegion
                      ? localizeRegionName(selectedRegion.region_id, selectedRegion.name, language)
                      : localizeRegionName(activeRegionID, activeRegionID, language)}
                  </strong>
                </div>
                <div className="feed-scope-actions">
                  {worldState.regions.slice(0, 6).map((region) => (
                    <Link
                      key={region.region_id}
                      className="inline-link"
                      href={`/chat?channel=region&region=${encodeURIComponent(region.region_id)}${
                        messageType !== "all" ? `&message_type=${messageType}` : ""
                      }`}
                    >
                      {localizeRegionName(region.region_id, region.name, language)}
                    </Link>
                  ))}
                </div>
              </div>
            ) : null}

            <div className="filter-row">
              {MESSAGE_FILTERS.map((filter) => (
                <Link
                  key={filter.key}
                  className={`filter-pill ${messageType === filter.key ? "active" : ""}`}
                  href={makeHref(channel, filter.key)}
                >
                  {language === "zh-CN" ? filter.zh : filter.en}
                </Link>
              ))}
            </div>

            <div className="log-list">
              {filteredMessages.length > 0 ? (
                filteredMessages.map((message) => (
                  <article key={message.message_id} className="chat-entry-card">
                    <div className="chat-entry-top">
                      <Link className="inline-link" href={`/bots/${encodeURIComponent(message.bot_id)}`}>
                        {message.bot_name}
                      </Link>
                      <div className="chat-entry-badges">
                        <span className="chat-badge">
                          {message.channel_type === "region"
                            ? language === "zh-CN"
                              ? "地区频道"
                              : "Region"
                            : language === "zh-CN"
                              ? "世界频道"
                              : "World"}
                        </span>
                        <span className={`chat-badge type-${message.message_type}`}>
                          {message.message_type === "friend_recruit"
                            ? language === "zh-CN"
                              ? "好友招募"
                              : "Friend Recruit"
                            : message.message_type === "assist_ad"
                              ? language === "zh-CN"
                                ? "助战宣传"
                                : "Assist Ad"
                              : language === "zh-CN"
                                ? "普通发言"
                                : "Free Text"}
                        </span>
                      </div>
                    </div>
                    <p className="chat-entry-content">{message.content}</p>
                    <p className="chat-entry-meta">
                      {message.region_id ? (
                        <Link className="inline-link" href={`/regions/${encodeURIComponent(message.region_id)}`}>
                          {localizeRegionName(message.region_id, message.region_id, language)}
                        </Link>
                      ) : (
                        <span>{language === "zh-CN" ? "全世界可见" : "Visible globally"}</span>
                      )}
                      <span>{formatRelativeTime(message.created_at, language)}</span>
                    </p>
                  </article>
                ))
              ) : (
                <p className="empty-state">
                  {language === "zh-CN" ? "当前筛选条件下暂无公开聊天消息。" : "No public chat messages match the current filter."}
                </p>
              )}
            </div>
          </section>
        </section>
      </section>
    </main>
  );
}
