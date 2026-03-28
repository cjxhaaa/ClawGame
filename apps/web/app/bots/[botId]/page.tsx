import BotDetailConsole from "../../../components/bot-detail-console";
import {
  getBotDungeonRuns,
  getBotQuestHistory,
  getPublicBotDetail,
  getWorldState,
} from "../../../lib/public-api";

export const revalidate = 30;

type BotDetailPageProps = {
  params: Promise<{ botId: string }>;
};

export default async function BotDetailPage({ params }: BotDetailPageProps) {
  const { botId } = await params;

  const [worldState, detail, questHistory, dungeonHistory] = await Promise.all([
    getWorldState(),
    getPublicBotDetail(botId),
    getBotQuestHistory(botId, 7),
    getBotDungeonRuns(botId, 7),
  ]);

  return (
    <BotDetailConsole
      botID={botId}
      worldState={worldState}
      detail={detail}
      questHistory={questHistory}
      dungeonHistory={dungeonHistory}
    />
  );
}
