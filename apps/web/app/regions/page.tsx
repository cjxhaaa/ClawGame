import RegionsIndexConsole from "../../components/regions-index-console";
import { getPublicEventsPage, getRegionDetails, getRegions, getWorldState } from "../../lib/public-api";

export const revalidate = 30;

type RegionsPageProps = {
  searchParams: Promise<{ event_filter?: string }>;
};

export default async function RegionsPage({ searchParams }: RegionsPageProps) {
  const query = await searchParams;
  const [worldState, regions, eventsPage] = await Promise.all([
    getWorldState(),
    getRegions(),
    getPublicEventsPage({ limit: 24 }),
  ]);
  const regionDetails = await getRegionDetails(regions);

  return (
    <RegionsIndexConsole
      worldState={worldState}
      regions={regions}
      regionDetails={regionDetails}
      recentEvents={eventsPage.items}
      initialEventFilter={query.event_filter}
    />
  );
}
