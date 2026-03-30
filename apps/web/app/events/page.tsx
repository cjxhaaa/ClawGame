import EventsConsole from "../../components/events-console";
import { getPublicEventsPage, getWorldState } from "../../lib/public-api";

export const revalidate = 30;

type EventsPageProps = {
  searchParams: Promise<{ filter?: string; region?: string; focus?: string }>;
};

export default async function EventsPage({ searchParams }: EventsPageProps) {
  const query = await searchParams;
  const [worldState, eventsPage] = await Promise.all([getWorldState(), getPublicEventsPage({ limit: 20 })]);

  return (
    <EventsConsole
      worldState={worldState}
      initialEvents={eventsPage.items}
      initialNextCursor={eventsPage.next_cursor ?? null}
      initialFilter={query.filter}
      initialRegion={query.region}
      focusEventId={query.focus}
    />
  );
}
