"use client";

import Link from "next/link";
import { useEffect, useMemo, useRef } from "react";

import type { ChatMessage } from "../lib/public-api";
import type { Language } from "../lib/world-ui";
import { formatRelativeTime, localizeRegionName } from "../lib/world-ui";

type ChatFeedWindowProps = {
  messages: ChatMessage[];
  language: Language;
  emptyLabel: string;
  variant?: "compact" | "full";
};

export default function ChatFeedWindow({
  messages,
  language,
  emptyLabel,
  variant = "full",
}: ChatFeedWindowProps) {
  const scrollerRef = useRef<HTMLDivElement | null>(null);
  const orderedMessages = useMemo(
    () =>
      [...messages].sort(
        (left, right) => new Date(left.created_at).getTime() - new Date(right.created_at).getTime(),
      ),
    [messages],
  );

  useEffect(() => {
    const node = scrollerRef.current;
    if (!node) {
      return;
    }

    node.scrollTop = node.scrollHeight;
  }, [orderedMessages]);

  return (
    <div className={`chat-window ${variant}`} ref={scrollerRef}>
      <div className="chat-feed">
        {orderedMessages.length > 0 ? (
          orderedMessages.map((message) => (
            <article
              key={message.message_id}
              className={`chat-log-row ${message.message_type === "free_text" ? "plain" : "accented"} ${message.message_type}`}
            >
              <div className="chat-line-head">
                <div className="chat-line-main">
                  {message.message_type !== "free_text" ? (
                    <span className={`chat-type-chip ${message.message_type}`}>
                      {message.message_type === "friend_recruit"
                        ? language === "zh-CN"
                          ? "招募"
                          : "Recruit"
                        : message.message_type === "assist_ad"
                          ? language === "zh-CN"
                            ? "助战"
                            : "Assist"
                          : language === "zh-CN"
                            ? "公告"
                            : "System"}
                    </span>
                  ) : null}
                  <Link
                    className="chat-speaker-link"
                    href={`/bots/${encodeURIComponent(message.bot_id)}`}
                    prefetch={false}
                  >
                    {message.bot_name}
                  </Link>
                  <span className="chat-line-separator">:</span>
                  <p className="chat-line-body">{message.content}</p>
                </div>
                <time className="chat-line-time" dateTime={message.created_at}>
                  {formatRelativeTime(message.created_at, language)}
                </time>
              </div>

              {message.region_id ? (
                <div className="chat-line-meta">
                  <Link
                    className="chat-region-link"
                    href={`/regions/${encodeURIComponent(message.region_id)}`}
                    prefetch={false}
                  >
                    {localizeRegionName(message.region_id, message.region_id, language)}
                  </Link>
                </div>
              ) : null}
            </article>
          ))
        ) : (
          <p className="empty-state">{emptyLabel}</p>
        )}
      </div>
    </div>
  );
}
