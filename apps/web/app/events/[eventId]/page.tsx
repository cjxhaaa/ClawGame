import { notFound } from "next/navigation";

import EventDetailConsole from "../../../components/event-detail-console";
import {
  getBotQuestHistory,
  getDungeonRunDetail,
  getPublicBotDetail,
  getPublicEventDetail,
} from "../../../lib/public-api";

export const revalidate = 30;

type EventDetailPageProps = {
  params: Promise<{ eventId: string }>;
  searchParams: Promise<{ filter?: string; region?: string; focus?: string }>;
};

export default async function EventDetailPage({ params, searchParams }: EventDetailPageProps) {
  const [{ eventId }, query] = await Promise.all([params, searchParams]);
  const event = await getPublicEventDetail(eventId);

  if (!event) {
    notFound();
  }

  const actorCharacterID = event.actor_character_id?.trim() || "";
  const payload = event.payload ?? {};
  const runID = typeof payload.run_id === "string" ? payload.run_id : "";
  const questID = typeof payload.quest_id === "string" ? payload.quest_id : "";

  const [actorDetail, questHistory, runDetail] = await Promise.all([
    actorCharacterID ? getPublicBotDetail(actorCharacterID) : Promise.resolve(null),
    actorCharacterID && questID ? getBotQuestHistory(actorCharacterID, 7) : Promise.resolve([]),
    actorCharacterID && runID ? getDungeonRunDetail(actorCharacterID, runID) : Promise.resolve(null),
  ]);

  const questRecord = questID ? questHistory.find((item) => item.quest_id === questID) ?? null : null;
  const returnFeedHref = `/events${buildQuery(query.filter, query.region, query.focus ?? event.event_id)}`;
  const returnRegionHref = query.region ? `/regions/${query.region}` : event.region_id ? `/regions/${event.region_id}` : null;

  return (
    <EventDetailConsole
      event={event}
      actorDetail={actorDetail}
      questRecord={questRecord}
      runDetail={runDetail}
      returnFeedHref={returnFeedHref}
      returnRegionHref={returnRegionHref}
    />
  );
}

function buildQuery(filter?: string, region?: string, focus?: string) {
  const search = new URLSearchParams();
  if (filter) search.set("filter", filter);
  if (region) search.set("region", region);
  if (focus) search.set("focus", focus);
  const suffix = search.toString();
  return suffix ? `?${suffix}` : "";
}
