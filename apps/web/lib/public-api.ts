type Envelope<T> = {
  data: T;
};

export type Region = {
  region_id: string;
  name: string;
  type: string;
  min_rank: string;
  travel_cost_gold: number;
};

export type Building = {
  building_id: string;
  name: string;
  type: string;
  actions: string[];
};

export type TravelOption = {
  region_id: string;
  name: string;
  travel_cost_gold: number;
  requires_rank: string;
};

export type EncounterSummary = {
  activity_type: string;
  summary: string;
  highlights: string[];
};

export type RegionDetail = {
  region: Region;
  description: string;
  buildings: Building[];
  travel_options: TravelOption[];
  encounter_summary?: EncounterSummary;
};

export type RegionActivity = Region & {
  population: number;
  recent_event_count: number;
  highlight: string;
  building_count: number;
};

export type ArenaStatus = {
  code: string;
  label: string;
  details: string;
  next_milestone: string;
};

export type PublicWorldState = {
  server_time: string;
  daily_reset_at: string;
  active_bot_count: number;
  bots_in_dungeon_count: number;
  bots_in_arena_count: number;
  quests_completed_today: number;
  dungeon_clears_today: number;
  gold_minted_today: number;
  regions: RegionActivity[];
  current_arena_status: ArenaStatus;
};

export type WorldEvent = {
  event_id: string;
  event_type: string;
  actor_name?: string;
  region_id?: string;
  summary: string;
  occurred_at: string;
};

export type LeaderboardEntry = {
  rank: number;
  character_id: string;
  name: string;
  class: string;
  weapon_style: string;
  region_id: string;
  score: number;
  score_label: string;
  activity_label: string;
};

export type Leaderboards = {
  reputation: LeaderboardEntry[];
  gold: LeaderboardEntry[];
  weekly_arena: LeaderboardEntry[];
  dungeon_clears: LeaderboardEntry[];
};

export const fallbackWorldState: PublicWorldState = {
  server_time: new Date().toISOString(),
  daily_reset_at: new Date(Date.now() + 6 * 60 * 60 * 1000).toISOString(),
  active_bot_count: 0,
  bots_in_dungeon_count: 0,
  bots_in_arena_count: 0,
  quests_completed_today: 0,
  dungeon_clears_today: 0,
  gold_minted_today: 0,
  current_arena_status: {
    code: "offline",
    label: "Offline",
    details: "Public API is unavailable, so the console is showing fallback data.",
    next_milestone: "Retry when API is reachable",
  },
  regions: [],
};

export const fallbackRegions: Region[] = [];
export const fallbackRegionDetails: RegionDetail[] = [];
export const fallbackEvents: WorldEvent[] = [];
export const fallbackLeaderboards: Leaderboards = {
  reputation: [],
  gold: [],
  weekly_arena: [],
  dungeon_clears: [],
};

export async function getHomepageData() {
  const [worldState, regions, events, leaderboards] = await Promise.all([
    getWorldState(),
    getRegions(),
    getEvents(6),
    getLeaderboards(),
  ]);

  const regionDetails = await getRegionDetails(regions);

  return {
    worldState,
    regions,
    regionDetails,
    events,
    leaderboards,
  };
}

export async function getWorldState() {
  return fetchData<PublicWorldState>("/api/v1/public/world-state", fallbackWorldState);
}

export async function getRegions() {
  const payload = await fetchData<{ regions: Region[] }>("/api/v1/world/regions", {
    regions: fallbackRegions,
  });

  return payload.regions;
}

export async function getRegionDetail(regionID: string, fallbackRegion?: Region) {
  const detail = await fetchData<RegionDetail>(`/api/v1/regions/${encodeURIComponent(regionID)}`, {
    region: fallbackRegion ?? {
      region_id: regionID,
      name: regionID,
      type: "field",
      min_rank: "low",
      travel_cost_gold: 0,
    },
    description: "",
    buildings: [],
    travel_options: [],
  });

  return {
    ...detail,
    buildings: detail.buildings ?? [],
    travel_options: detail.travel_options ?? [],
    encounter_summary: detail.encounter_summary
      ? {
          ...detail.encounter_summary,
          highlights: detail.encounter_summary.highlights ?? [],
        }
      : undefined,
  };
}

export async function getRegionDetails(regions: Region[]) {
  return Promise.all(regions.map((region) => getRegionDetail(region.region_id, region)));
}

export async function getEvents(limit = 6) {
  const payload = await fetchData<{ items: WorldEvent[] }>(`/api/v1/public/events?limit=${limit}`, {
    items: fallbackEvents,
  });

  return payload.items;
}

export async function getLeaderboards() {
  return fetchData<Leaderboards>("/api/v1/public/leaderboards", fallbackLeaderboards);
}

async function fetchData<T>(path: string, fallback: T): Promise<T> {
  try {
    const response = await fetch(`${apiBaseUrl()}${path}`, {
      next: { revalidate: 30 },
    });

    if (!response.ok) {
      throw new Error(`request failed with ${response.status}`);
    }

    const payload = (await response.json()) as Envelope<T>;
    return payload.data;
  } catch {
    return fallback;
  }
}

function apiBaseUrl() {
  const baseUrl =
    process.env.API_BASE_URL ??
    process.env.NEXT_PUBLIC_API_BASE_URL ??
    "http://127.0.0.1:8080";

  return baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
}
