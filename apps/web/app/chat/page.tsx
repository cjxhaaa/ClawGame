import ChatConsole from "../../components/chat-console";
import { getPublicRegionChatPage, getPublicWorldChatPage, getRegions, getWorldState } from "../../lib/public-api";

export const revalidate = 30;

type ChatPageProps = {
  searchParams: Promise<{
    channel?: string;
    region?: string;
    message_type?: string;
  }>;
};

function toMessageType(value?: string) {
  if (value === "free_text" || value === "friend_recruit" || value === "assist_ad") {
    return value;
  }
  return "all";
}

export default async function ChatPage({ searchParams }: ChatPageProps) {
  const query = await searchParams;
  const channel = query.channel === "region" ? "region" : "world";
  const messageType = toMessageType(query.message_type);
  const regionID = query.region || "main_city";

  const [worldState, regions, chatPage] = await Promise.all([
    getWorldState(),
    getRegions(),
    channel === "region"
      ? getPublicRegionChatPage({
          regionID,
          limit: 30,
          messageType: messageType === "all" ? undefined : messageType,
        })
      : getPublicWorldChatPage({
          limit: 30,
          messageType: messageType === "all" ? undefined : messageType,
        }),
  ]);

  return (
    <ChatConsole
      worldState={worldState}
      regions={regions}
      channel={channel}
      regionID={channel === "region" ? regionID : undefined}
      messageType={messageType}
      messages={chatPage.items}
    />
  );
}
