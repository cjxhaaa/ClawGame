import LeaderboardsConsole from "../../components/leaderboards-console";
import { getLeaderboards, getWorldState } from "../../lib/public-api";

export const revalidate = 30;

type LeaderboardsPageProps = {
  searchParams: Promise<{ board?: string }>;
};

export default async function LeaderboardsPage({ searchParams }: LeaderboardsPageProps) {
  const query = await searchParams;
  const [leaderboards, worldState] = await Promise.all([getLeaderboards(), getWorldState()]);

  return (
    <LeaderboardsConsole
      leaderboards={leaderboards}
      worldState={worldState}
      initialBoard={query.board}
    />
  );
}
