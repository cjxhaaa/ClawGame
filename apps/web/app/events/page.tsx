import EventsConsole from "../../components/events-console";
import { getEvents, getWorldState } from "../../lib/public-api";

export const revalidate = 30;

type EventsPageProps = {
  searchParams: Promise<{ filter?: string }>;
};

export default async function EventsPage({ searchParams }: EventsPageProps) {
  const query = await searchParams;
  const [worldState, events] = await Promise.all([getWorldState(), getEvents(32)]);

  return <EventsConsole worldState={worldState} events={events} initialFilter={query.filter} />;
}
