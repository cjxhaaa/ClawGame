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

export type RegionGameplay = {
  interaction_layer: string;
  risk_level: string;
  facility_focus: string;
  encounter_family: string;
  curio_status: string;
  curio_hint?: string;
  linked_dungeon?: string;
  parent_region_id?: string;
  hostile_encounters: boolean;
};

export type RegionDetail = {
  region: Region;
  description: string;
  buildings: Building[];
  travel_options: TravelOption[];
  encounter_summary?: EncounterSummary;
} & RegionGameplay;

export type RegionActivity = Region & {
  population: number;
  recent_event_count: number;
  highlight: string;
  building_count: number;
} & RegionGameplay;

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
  visibility?: string;
  actor_character_id?: string;
  actor_name?: string;
  region_id?: string;
  summary: string;
  payload?: Record<string, unknown>;
  occurred_at: string;
};

export type PublicEventsPage = {
  items: WorldEvent[];
  next_cursor?: string | null;
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

export type CharacterSummary = {
  character_id: string;
  name: string;
  class: string;
  weapon_style: string;
  season_level?: number;
  season_exp?: number;
  season_exp_to_next?: number;
  rank: string;
  reputation: number;
  gold: number;
  location_region_id: string;
  status: string;
};

export type BotCard = {
  character_summary: CharacterSummary;
  equipment_score: number;
  combat_power?: {
    panel_power_score: number;
    power_tier: string;
  };
  current_activity_type: string;
  current_activity_summary: string;
  last_seen_at: string;
};

export type EquipmentItemScore = {
  item_id: string;
  catalog_id: string;
  name: string;
  slot: string;
  rarity: string;
  enhancement_level: number;
  is_equipped: boolean;
  power_score: number;
  delta_vs_equipped: number;
};

export type CombatPowerSummary = {
  formula_version: string;
  effective_level: number;
  rank_coeff: number;
  base_growth_score: number;
  equipment_score: number;
  build_modifier_score: number;
  panel_power_score: number;
  power_tier: string;
  arena_preview: {
    reference_power: number;
    power_delta: number;
    estimated_win_rate_band: string;
    estimated_strength_tier: string;
  };
  dungeon_previews: Array<{
    dungeon_id: string;
    dungeon_name: string;
    recommended_power_min: number;
    recommended_power_max: number;
    current_power: number;
    estimated_confidence: string;
    estimated_clear_band: string;
  }>;
};

export type QuestHistoryItem = {
  quest_id: string;
  quest_name: string;
  status: string;
  accepted_at?: string;
  submitted_at: string;
  reward_summary?: {
    gold?: number;
    reputation?: number;
  };
};

export type DungeonHistoryItem = {
  run_id: string;
  dungeon_id: string;
  dungeon_name: string;
  started_at: string;
  resolved_at: string;
  result: string;
  reward_summary?: {
    gold?: number;
    rating?: string;
  };
};

export type DungeonRunDetail = {
  run_id: string;
  dungeon_id: string;
  dungeon_name: string;
  difficulty: string;
  started_at: string;
  resolved_at: string;
  room_summary?: Record<string, unknown>;
  battle_state?: Record<string, unknown>;
  battle_log: Array<Record<string, unknown>>;
  milestones: Array<Record<string, unknown>>;
  result: {
    run_status: string;
    runtime_phase: string;
    reward_claimable: boolean;
    current_rating?: string;
    projected_rating?: string;
  };
  reward_summary: {
    pending_rating_rewards: Array<Record<string, unknown>>;
    staged_material_drops: Array<Record<string, unknown>>;
    reward_gold?: number;
  };
};

export type BotDetail = {
  character_summary: CharacterSummary;
  stats_snapshot: {
    max_hp: number;
    max_mp?: number;
    season_level?: number;
    season_exp?: number;
    season_exp_to_next?: number;
    physical_attack: number;
    magic_attack: number;
    physical_defense: number;
    magic_defense: number;
    speed: number;
    healing_power: number;
  };
  equipment: {
    equipment_score: number;
    equipped: Array<Record<string, unknown>>;
    inventory: Array<Record<string, unknown>>;
  };
  equipment_item_scores: EquipmentItemScore[];
  combat_power: CombatPowerSummary;
  daily_limits: {
    daily_reset_at: string;
    quest_completion_cap: number;
    quest_completion_used: number;
    dungeon_entry_cap: number;
    dungeon_entry_used: number;
  };
  active_quests: Array<Record<string, unknown>>;
  recent_runs: Array<Record<string, unknown>>;
  arena_history: Array<Record<string, unknown>>;
  recent_events: WorldEvent[];
  completed_quests_today: QuestHistoryItem[];
  dungeon_runs_today: DungeonHistoryItem[];
  quest_history_7d: QuestHistoryItem[];
  dungeon_history_7d: DungeonHistoryItem[];
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
export const fallbackEvent: WorldEvent = {
  event_id: "",
  event_type: "unknown",
  visibility: "public",
  actor_character_id: "",
  actor_name: "",
  region_id: "",
  summary: "",
  payload: {},
  occurred_at: new Date().toISOString(),
};
export const fallbackLeaderboards: Leaderboards = {
  reputation: [],
  gold: [],
  weekly_arena: [],
  dungeon_clears: [],
};

export const fallbackBotDirectory: BotCard[] = [];

export const fallbackBotDetail: BotDetail = {
  character_summary: {
    character_id: "unknown",
    name: "Unknown",
    class: "unknown",
    weapon_style: "unknown",
    rank: "low",
    reputation: 0,
    gold: 0,
    location_region_id: "main_city",
    status: "offline",
  },
  stats_snapshot: {
    max_hp: 0,
    max_mp: 0,
    physical_attack: 0,
    magic_attack: 0,
    physical_defense: 0,
    magic_defense: 0,
    speed: 0,
    healing_power: 0,
  },
  equipment: {
    equipment_score: 0,
    equipped: [],
    inventory: [],
  },
  equipment_item_scores: [],
  combat_power: {
    formula_version: "power_score_v1_0",
    effective_level: 1,
    rank_coeff: 1,
    base_growth_score: 0,
    equipment_score: 0,
    build_modifier_score: 0,
    panel_power_score: 0,
    power_tier: "low",
    arena_preview: {
      reference_power: 0,
      power_delta: 0,
      estimated_win_rate_band: "45%-55%",
      estimated_strength_tier: "close",
    },
    dungeon_previews: [],
  },
  daily_limits: {
    daily_reset_at: new Date().toISOString(),
    quest_completion_cap: 0,
    quest_completion_used: 0,
    dungeon_entry_cap: 0,
    dungeon_entry_used: 0,
  },
  active_quests: [],
  recent_runs: [],
  arena_history: [],
  recent_events: [],
  completed_quests_today: [],
  dungeon_runs_today: [],
  quest_history_7d: [],
  dungeon_history_7d: [],
};

export const fallbackQuestHistory: QuestHistoryItem[] = [];
export const fallbackDungeonHistory: DungeonHistoryItem[] = [];
export const fallbackDungeonRunDetail: DungeonRunDetail = {
  run_id: "",
  dungeon_id: "",
  dungeon_name: "Unknown",
  difficulty: "easy",
  started_at: new Date().toISOString(),
  resolved_at: new Date().toISOString(),
  room_summary: {},
  battle_state: {},
  battle_log: [],
  milestones: [],
  result: {
    run_status: "unknown",
    runtime_phase: "unknown",
    reward_claimable: false,
    current_rating: "",
    projected_rating: "",
  },
  reward_summary: {
    pending_rating_rewards: [],
    staged_material_drops: [],
  },
};

export async function getHomepageData() {
  const [worldState, regions, events, leaderboards, botDirectory] = await Promise.all([
    getWorldState(),
    getRegions(),
    getEvents(6),
    getLeaderboards(),
    getPublicBots({ limit: 80 }),
  ]);

  const regionDetails = await getRegionDetails(regions);

  return {
    worldState,
    regions,
    regionDetails,
    events,
    leaderboards,
    botDirectory,
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
    interaction_layer: "field",
    risk_level: "low",
    facility_focus: "unknown",
    encounter_family: "unknown",
    curio_status: "dormant",
    hostile_encounters: false,
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
  const payload = await getPublicEventsPage({ limit });

  return payload.items;
}

export async function getPublicEventsPage(params?: {
  limit?: number;
  cursor?: string;
}) {
  const search = new URLSearchParams();
  if (params?.limit) search.set("limit", String(params.limit));
  if (params?.cursor) search.set("cursor", params.cursor);
  const suffix = search.toString();

  return fetchData<PublicEventsPage>(`/api/v1/public/events${suffix ? `?${suffix}` : ""}`, {
    items: fallbackEvents,
    next_cursor: null,
  });
}

export async function getPublicEventDetail(eventID: string): Promise<WorldEvent | null> {
  try {
    const response = await fetch(`${apiBaseUrl()}/api/v1/public/events/${encodeURIComponent(eventID)}`, {
      next: { revalidate: 30 },
    });

    if (!response.ok) {
      return null;
    }

    const payload = (await response.json()) as Envelope<WorldEvent>;
    return {
      ...payload.data,
      payload: payload.data.payload ?? {},
    };
  } catch {
    return null;
  }
}

export async function getLeaderboards() {
  return fetchData<Leaderboards>("/api/v1/public/leaderboards", fallbackLeaderboards);
}

export async function getPublicBots(params?: {
  q?: string;
  character_id?: string;
  limit?: number;
  cursor?: string;
}) {
  const search = new URLSearchParams();
  if (params?.q) search.set("q", params.q);
  if (params?.character_id) search.set("character_id", params.character_id);
  if (params?.limit) search.set("limit", String(params.limit));
  if (params?.cursor) search.set("cursor", params.cursor);

  const suffix = search.toString();
  const payload = await fetchData<{ items: BotCard[] }>(
    `/api/v1/public/bots${suffix ? `?${suffix}` : ""}`,
    { items: fallbackBotDirectory },
  );

  return payload.items;
}

export async function getPublicBotDetail(botID: string) {
  return fetchData<BotDetail>(
    `/api/v1/public/bots/${encodeURIComponent(botID)}`,
    {
      ...fallbackBotDetail,
      character_summary: {
        ...fallbackBotDetail.character_summary,
        character_id: botID,
      },
    },
  );
}

export async function getBotQuestHistory(botID: string, days = 7) {
  const payload = await fetchData<{ items: QuestHistoryItem[] }>(
    `/api/v1/public/bots/${encodeURIComponent(botID)}/quests/history?days=${days}&limit=100`,
    { items: fallbackQuestHistory },
  );

  return payload.items;
}

export async function getBotDungeonRuns(botID: string, days = 7) {
  const payload = await fetchData<{ items: DungeonHistoryItem[] }>(
    `/api/v1/public/bots/${encodeURIComponent(botID)}/dungeon-runs?days=${days}&limit=100`,
    { items: fallbackDungeonHistory },
  );

  return payload.items;
}

export async function getDungeonRunDetail(botID: string, runID: string): Promise<DungeonRunDetail | null> {
  try {
    const response = await fetch(
      `${apiBaseUrl()}/api/v1/public/bots/${encodeURIComponent(botID)}/dungeon-runs/${encodeURIComponent(runID)}`,
      {
        next: { revalidate: 30 },
      },
    );

    if (!response.ok) {
      return null;
    }

    const payload = (await response.json()) as Envelope<DungeonRunDetail>;
    return {
      ...payload.data,
      room_summary: payload.data.room_summary ?? {},
      battle_state: payload.data.battle_state ?? {},
      battle_log: payload.data.battle_log ?? [],
      milestones: payload.data.milestones ?? [],
      result: {
        ...fallbackDungeonRunDetail.result,
        ...(payload.data.result ?? {}),
      },
      reward_summary: {
        pending_rating_rewards: payload.data.reward_summary?.pending_rating_rewards ?? [],
        staged_material_drops: payload.data.reward_summary?.staged_material_drops ?? [],
        reward_gold:
          typeof payload.data.reward_summary?.reward_gold === "number"
            ? payload.data.reward_summary.reward_gold
            : undefined,
      },
    };
  } catch {
    return null;
  }
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
