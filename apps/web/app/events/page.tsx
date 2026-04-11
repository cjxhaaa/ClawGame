import { redirect } from "next/navigation";

export const revalidate = 30;

type EventsPageProps = {
  searchParams: Promise<{ filter?: string; region?: string; focus?: string }>;
};

export default async function EventsPage({ searchParams }: EventsPageProps) {
  const query = await searchParams;
  const search = new URLSearchParams();
  if (query.filter) {
    search.set("event_filter", query.filter);
  }
  redirect(`/regions${search.toString() ? `?${search.toString()}` : ""}#world-chronicle`);
}
