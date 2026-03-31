import ArenaConsole from "../../components/arena-console";
import { getArenaCurrent, getEvents, getLeaderboards, getWorldState } from "../../lib/public-api";

export const revalidate = 30;

export default async function ArenaPage() {
  const [worldState, leaderboards, events, arenaCurrent] = await Promise.all([
    getWorldState(),
    getLeaderboards(),
    getEvents(24),
    getArenaCurrent(),
  ]);

  return <ArenaConsole worldState={worldState} leaderboards={leaderboards} events={events} arenaCurrent={arenaCurrent} />;
}
