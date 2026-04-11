import ChatConsole from "../../components/chat-console";
import { getPublicRegionChatPage, getPublicWorldChatPage, getRegions, getWorldState } from "../../lib/public-api";

export const revalidate = 30;

type ChatPageProps = {
  searchParams: Promise<{
    scope?: string;
    channel?: string;
    region?: string;
    message_type?: string;
    cursor?: string;
  }>;
};

function toMessageType(value?: string) {
  if (value === "free_text" || value === "friend_recruit" || value === "assist_ad" || value === "system_notice") {
    return value;
  }
  return "all";
}

export default async function ChatPage({ searchParams }: ChatPageProps) {
  const query = await searchParams;
  const messageType = toMessageType(query.message_type);
  const cursor = query.cursor;
  const regions = await getRegions();
  const legacyScope = query.channel === "region" ? query.region || "main_city" : "world";
  const requestedScope = query.scope || legacyScope;
  const activeScope =
    requestedScope === "world" || regions.some((region) => region.region_id === requestedScope)
      ? requestedScope
      : "world";
  const activeRegionID = activeScope === "world" ? undefined : activeScope;

  const [worldState, chatPage] = await Promise.all([
    getWorldState(),
    activeRegionID
      ? getPublicRegionChatPage({
          regionID: activeRegionID,
          limit: 30,
          cursor,
          messageType: messageType === "all" ? undefined : messageType,
        })
      : getPublicWorldChatPage({
          limit: 30,
          cursor,
          messageType: messageType === "all" ? undefined : messageType,
        }),
  ]);

  return (
    <ChatConsole
      worldState={worldState}
      regions={regions}
      activeScope={activeScope}
      messageType={messageType}
      messages={chatPage.items}
      nextCursor={chatPage.next_cursor ?? undefined}
      cursor={cursor}
    />
  );
}
