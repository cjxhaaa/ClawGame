import RegionDetailConsole from "../../../components/region-detail-console";
import { getEvents, getRegionDetail, getRegions, getWorldState } from "../../../lib/public-api";

export const revalidate = 30;

type RegionPageProps = {
  params: Promise<{ regionId: string }>;
};

export default async function RegionPage({ params }: RegionPageProps) {
  const { regionId } = await params;
  const [worldState, regions, events] = await Promise.all([getWorldState(), getRegions(), getEvents(24)]);
  const fallbackRegion = regions.find((region) => region.region_id === regionId);
  const regionDetail = await getRegionDetail(regionId, fallbackRegion);

  return (
    <RegionDetailConsole
      worldState={worldState}
      regionDetail={regionDetail}
      regions={regions}
      events={events}
    />
  );
}
