CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS regions (
    id text PRIMARY KEY,
    name text NOT NULL,
    type text NOT NULL,
    min_rank text NOT NULL DEFAULT 'low',
    travel_cost_gold integer NOT NULL DEFAULT 0,
    sort_order integer NOT NULL,
    is_active boolean NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS buildings (
    id text PRIMARY KEY,
    region_id text NOT NULL REFERENCES regions(id),
    name text NOT NULL,
    type text NOT NULL,
    sort_order integer NOT NULL,
    is_active boolean NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_buildings_region_id ON buildings(region_id);

CREATE TABLE IF NOT EXISTS accounts (
    id text PRIMARY KEY,
    bot_name citext NOT NULL UNIQUE,
    password_hash text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id text PRIMARY KEY,
    account_id text NOT NULL REFERENCES accounts(id),
    refresh_token_hash text NOT NULL,
    user_agent text,
    ip_address inet,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_account_id ON auth_sessions(account_id);
CREATE INDEX IF NOT EXISTS idx_auth_sessions_expires_at ON auth_sessions(expires_at);

CREATE TABLE IF NOT EXISTS auth_challenges (
    id text PRIMARY KEY,
    prompt_text text NOT NULL,
    answer_format text NOT NULL,
    expected_answer text NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_challenges_expires_at ON auth_challenges(expires_at);

CREATE TABLE IF NOT EXISTS characters (
    id text PRIMARY KEY,
    account_id text NOT NULL UNIQUE REFERENCES accounts(id),
    name text NOT NULL UNIQUE,
    class text NOT NULL,
    weapon_style text NOT NULL,
    rank text NOT NULL DEFAULT 'low',
    reputation integer NOT NULL DEFAULT 0,
    gold bigint NOT NULL DEFAULT 0,
    status text NOT NULL DEFAULT 'active',
    location_region_id text NOT NULL REFERENCES regions(id),
    hp_current integer NOT NULL,
    mp_current integer NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_characters_rank ON characters(rank);
CREATE INDEX IF NOT EXISTS idx_characters_location_region_id ON characters(location_region_id);

CREATE TABLE IF NOT EXISTS character_base_stats (
    character_id text PRIMARY KEY REFERENCES characters(id),
    max_hp integer NOT NULL,
    max_mp integer NOT NULL,
    physical_attack integer NOT NULL,
    magic_attack integer NOT NULL,
    physical_defense integer NOT NULL,
    magic_defense integer NOT NULL,
    speed integer NOT NULL,
    healing_power integer NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS character_daily_limits (
    character_id text PRIMARY KEY REFERENCES characters(id),
    reset_date date NOT NULL,
    quest_completion_cap integer NOT NULL,
    quest_completion_used integer NOT NULL DEFAULT 0,
    dungeon_entry_cap integer NOT NULL,
    dungeon_entry_used integer NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS items_catalog (
    id text PRIMARY KEY,
    name text NOT NULL,
    slot text NOT NULL,
    rarity text NOT NULL,
    required_class text,
    required_weapon_style text,
    base_stats_json jsonb NOT NULL,
    passive_affix_json jsonb,
    sell_price_gold integer NOT NULL,
    enhanceable boolean NOT NULL DEFAULT false,
    max_enhancement_level integer NOT NULL DEFAULT 0,
    is_active boolean NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS item_instances (
    id text PRIMARY KEY,
    owner_character_id text NOT NULL REFERENCES characters(id),
    catalog_id text NOT NULL REFERENCES items_catalog(id),
    state text NOT NULL,
    slot text NOT NULL,
    enhancement_level integer NOT NULL DEFAULT 0,
    durability integer NOT NULL DEFAULT 100,
    obtained_at timestamptz NOT NULL,
    sold_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_item_instances_owner_character_id ON item_instances(owner_character_id);
CREATE INDEX IF NOT EXISTS idx_item_instances_owner_state ON item_instances(owner_character_id, state);

CREATE TABLE IF NOT EXISTS character_equipment (
    character_id text NOT NULL REFERENCES characters(id),
    slot text NOT NULL,
    item_id text NOT NULL UNIQUE REFERENCES item_instances(id),
    equipped_at timestamptz NOT NULL,
    PRIMARY KEY (character_id, slot)
);

CREATE TABLE IF NOT EXISTS quest_boards (
    id text PRIMARY KEY,
    character_id text NOT NULL REFERENCES characters(id),
    reset_date date NOT NULL,
    status text NOT NULL,
    reroll_count integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    UNIQUE (character_id, reset_date)
);

CREATE TABLE IF NOT EXISTS quests (
    id text PRIMARY KEY,
    board_id text NOT NULL REFERENCES quest_boards(id),
    character_id text NOT NULL REFERENCES characters(id),
    template_type text NOT NULL,
    rarity text NOT NULL,
    status text NOT NULL,
    title text NOT NULL,
    description text NOT NULL,
    target_region_id text REFERENCES regions(id),
    target_dungeon_id text,
    target_enemy_key text,
    progress_current integer NOT NULL DEFAULT 0,
    progress_target integer NOT NULL,
    reward_gold integer NOT NULL,
    reward_reputation integer NOT NULL,
    reward_item_catalog_id text REFERENCES items_catalog(id),
    accepted_at timestamptz,
    completed_at timestamptz,
    submitted_at timestamptz,
    expires_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_quests_character_id ON quests(character_id);
CREATE INDEX IF NOT EXISTS idx_quests_character_status ON quests(character_id, status);
CREATE INDEX IF NOT EXISTS idx_quests_board_id ON quests(board_id);

CREATE TABLE IF NOT EXISTS dungeon_definitions (
    id text PRIMARY KEY,
    name text NOT NULL,
    min_rank text NOT NULL,
    region_id text NOT NULL REFERENCES regions(id),
    room_count integer NOT NULL,
    boss_room_index integer NOT NULL,
    rating_reward_profile_id text NOT NULL DEFAULT '',
    room_config_json jsonb NOT NULL DEFAULT '{}',
    is_active boolean NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS dungeon_runs (
    id text PRIMARY KEY,
    character_id text NOT NULL REFERENCES characters(id),
    dungeon_id text NOT NULL REFERENCES dungeon_definitions(id),
    status text NOT NULL,
    runtime_phase text NOT NULL DEFAULT 'queued',
    current_room_index integer NOT NULL DEFAULT 1,
    highest_room_cleared integer NOT NULL DEFAULT 0,
    current_rating text,
    seed bigint NOT NULL,
    party_snapshot_json jsonb NOT NULL DEFAULT '{}',
    run_summary_json jsonb NOT NULL DEFAULT '{}',
    started_at timestamptz NOT NULL,
    finished_at timestamptz,
    last_action_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dungeon_runs_character_id ON dungeon_runs(character_id);
CREATE INDEX IF NOT EXISTS idx_dungeon_runs_character_status ON dungeon_runs(character_id, status);

CREATE TABLE IF NOT EXISTS dungeon_run_states (
    run_id text PRIMARY KEY REFERENCES dungeon_runs(id),
    state_version integer NOT NULL DEFAULT 1,
    state_json jsonb NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS arena_tournaments (
    id text PRIMARY KEY,
    week_key text NOT NULL UNIQUE,
    status text NOT NULL,
    signup_opens_at timestamptz NOT NULL,
    signup_closes_at timestamptz NOT NULL,
    starts_at timestamptz NOT NULL,
    completed_at timestamptz,
    bracket_size integer NOT NULL,
    snapshot_json jsonb
);

CREATE TABLE IF NOT EXISTS arena_entries (
    id text PRIMARY KEY,
    tournament_id text NOT NULL REFERENCES arena_tournaments(id),
    character_id text NOT NULL REFERENCES characters(id),
    status text NOT NULL,
    seed_number integer,
    equipment_score integer NOT NULL,
    signed_up_at timestamptz NOT NULL,
    final_rank integer,
    UNIQUE (tournament_id, character_id)
);

CREATE INDEX IF NOT EXISTS idx_arena_entries_tournament_id ON arena_entries(tournament_id);

CREATE TABLE IF NOT EXISTS arena_matches (
    id text PRIMARY KEY,
    tournament_id text NOT NULL REFERENCES arena_tournaments(id),
    round_number integer NOT NULL,
    match_number integer NOT NULL,
    left_character_id text REFERENCES characters(id),
    right_character_id text REFERENCES characters(id),
    winner_character_id text REFERENCES characters(id),
    status text NOT NULL,
    battle_log_json jsonb,
    scheduled_at timestamptz NOT NULL,
    resolved_at timestamptz,
    UNIQUE (tournament_id, round_number, match_number)
);

CREATE TABLE IF NOT EXISTS leaderboard_snapshots (
    id text PRIMARY KEY,
    leaderboard_type text NOT NULL,
    scope_key text NOT NULL,
    generated_at timestamptz NOT NULL,
    payload_json jsonb NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_leaderboard_snapshots_type_scope
    ON leaderboard_snapshots(leaderboard_type, scope_key);

CREATE TABLE IF NOT EXISTS world_events (
    id text PRIMARY KEY,
    event_type text NOT NULL,
    visibility text NOT NULL,
    actor_account_id text REFERENCES accounts(id),
    actor_character_id text REFERENCES characters(id),
    actor_name text,
    region_id text REFERENCES regions(id),
    related_entity_type text,
    related_entity_id text,
    summary text NOT NULL,
    payload_json jsonb NOT NULL,
    occurred_at timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_world_events_occurred_at_desc ON world_events(occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_world_events_visibility_occurred_at_desc
    ON world_events(visibility, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_world_events_actor_character_id ON world_events(actor_character_id);
CREATE INDEX IF NOT EXISTS idx_world_events_region_id ON world_events(region_id);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    idempotency_key text NOT NULL,
    account_id text NOT NULL REFERENCES accounts(id),
    route_key text NOT NULL,
    request_hash text NOT NULL,
    response_status_code integer NOT NULL,
    response_body_json jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    PRIMARY KEY (idempotency_key, account_id, route_key)
);

INSERT INTO regions (id, name, type, min_rank, travel_cost_gold, sort_order, is_active)
VALUES
    ('main_city', 'Main City', 'safe_hub', 'low', 0, 1, true),
    ('greenfield_village', 'Greenfield Village', 'safe_hub', 'low', 0, 2, true),
    ('whispering_forest', 'Whispering Forest', 'field', 'low', 10, 3, true),
    ('ancient_catacomb', 'Ancient Catacomb', 'dungeon', 'low', 15, 4, true),
    ('sunscar_desert_outskirts', 'Sunscar Desert Outskirts', 'field', 'mid', 30, 5, true),
    ('sandworm_den', 'Sandworm Den', 'dungeon', 'high', 50, 6, true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO buildings (id, region_id, name, type, sort_order, is_active)
VALUES
    ('guild_main_city', 'main_city', 'Adventurers Guild', 'guild', 1, true),
    ('weapon_shop_main_city', 'main_city', 'Weapon Shop', 'weapon_shop', 2, true),
    ('armor_shop_main_city', 'main_city', 'Armor Shop', 'armor_shop', 3, true),
    ('temple_main_city', 'main_city', 'Temple', 'temple', 4, true),
    ('blacksmith_main_city', 'main_city', 'Blacksmith', 'blacksmith', 5, true),
    ('arena_hall_main_city', 'main_city', 'Arena Hall', 'arena_hall', 6, true),
    ('warehouse_main_city', 'main_city', 'Warehouse', 'warehouse', 7, true),
    ('quest_outpost_village', 'greenfield_village', 'Quest Outpost', 'quest_outpost', 1, true),
    ('general_store_village', 'greenfield_village', 'General Store', 'general_store', 2, true),
    ('field_healer_village', 'greenfield_village', 'Field Healer', 'healer', 3, true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO dungeon_definitions (id, name, min_rank, region_id, encounter_count, boss_encounter_key, reward_table_json, is_active)
VALUES
    ('ancient_catacomb_v1', 'Ancient Catacomb', 'low', 'ancient_catacomb', 4, 'catacomb_boss_necromancer', '{"gold_min":180,"gold_max":260,"drop_table":"catacomb_v1"}', true),
    ('sandworm_den_v1', 'Sandworm Den', 'high', 'sandworm_den', 5, 'sandworm_boss_matriarch', '{"gold_min":320,"gold_max":460,"drop_table":"sandworm_v1"}', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO items_catalog (
    id, name, slot, rarity, required_class, required_weapon_style, base_stats_json,
    passive_affix_json, sell_price_gold, enhanceable, max_enhancement_level, is_active
)
VALUES
    ('warrior_sword_starter', 'Recruit Sword', 'weapon', 'common', 'warrior', 'sword_shield', '{"physical_attack":6}', NULL, 20, true, 5, true),
    ('warrior_axe_starter', 'Recruit Axe', 'weapon', 'common', 'warrior', 'great_axe', '{"physical_attack":7}', NULL, 20, true, 5, true),
    ('mage_staff_starter', 'Ashwood Staff', 'weapon', 'common', 'mage', 'staff', '{"magic_attack":8}', NULL, 20, true, 5, true),
    ('mage_spellbook_starter', 'Trainee Spellbook', 'weapon', 'common', 'mage', 'spellbook', '{"magic_attack":8}', NULL, 20, true, 5, true),
    ('priest_scepter_starter', 'Pilgrim Scepter', 'weapon', 'common', 'priest', 'scepter', '{"healing_power":6,"magic_attack":4}', NULL, 20, true, 5, true),
    ('priest_tome_starter', 'Pilgrim Tome', 'weapon', 'common', 'priest', 'holy_tome', '{"healing_power":5,"magic_attack":5}', NULL, 20, true, 5, true),
    ('starter_chest_cloth', 'Novice Robe', 'chest', 'common', NULL, NULL, '{"magic_defense":3}', NULL, 15, true, 5, true),
    ('starter_chest_armor', 'Novice Armor', 'chest', 'common', NULL, NULL, '{"physical_defense":4,"max_hp":12}', NULL, 15, true, 5, true),
    ('starter_boots', 'Trail Boots', 'boots', 'common', NULL, NULL, '{"speed":2}', NULL, 10, false, 0, true)
ON CONFLICT (id) DO NOTHING;
