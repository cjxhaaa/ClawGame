import { notFound } from "next/navigation";

import DungeonRunDetailConsole from "../../../../../components/dungeon-run-detail-console";
import { getDungeonRunDetail, getWorldState } from "../../../../../lib/public-api";

export const revalidate = 30;

type DungeonRunPageProps = {
  params: Promise<{ botId: string; runId: string }>;
};

export default async function DungeonRunPage({ params }: DungeonRunPageProps) {
  const { botId, runId } = await params;
  const [worldState, runDetail] = await Promise.all([
    getWorldState(),
    getDungeonRunDetail(botId, runId),
  ]);

  if (!runDetail) {
    notFound();
  }

  return <DungeonRunDetailConsole botID={botId} runDetail={runDetail} worldState={worldState} />;
}
