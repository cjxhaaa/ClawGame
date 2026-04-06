package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"clawgame/apps/api/internal/auth"
	"clawgame/apps/api/internal/characters"
	"clawgame/apps/api/internal/quests"
	"clawgame/apps/api/internal/world"

	_ "github.com/lib/pq"
)

const businessTimezone = "Asia/Shanghai"

type PostgresStore struct {
	db  *sql.DB
	loc *time.Location
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	store := &PostgresStore{
		db:  db,
		loc: mustLocation(businessTimezone),
	}
	if err := store.ensureSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *PostgresStore) LoadAccounts() ([]auth.StoredAccount, error) {
	rows, err := s.db.Query(`
		SELECT id, bot_name, password_hash, created_at
		FROM accounts
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]auth.StoredAccount, 0)
	for rows.Next() {
		var accountID string
		var botName string
		var password string
		var createdAt time.Time

		if err := rows.Scan(&accountID, &botName, &password, &createdAt); err != nil {
			return nil, err
		}

		accounts = append(accounts, auth.StoredAccount{
			Account: auth.Account{
				AccountID: accountID,
				BotName:   botName,
				CreatedAt: createdAt.In(s.loc).Format(time.RFC3339),
			},
			Password: password,
		})
	}

	return accounts, rows.Err()
}

func (s *PostgresStore) SaveAccount(stored auth.StoredAccount) error {
	createdAt := parseRFC3339InLocation(stored.Account.CreatedAt, s.loc)
	now := time.Now().In(s.loc)

	_, err := s.db.Exec(`
		INSERT INTO accounts (id, bot_name, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'active', $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET bot_name = EXCLUDED.bot_name,
		    password_hash = EXCLUDED.password_hash,
		    updated_at = EXCLUDED.updated_at
	`, stored.Account.AccountID, stored.Account.BotName, stored.Password, createdAt, now)
	return err
}

func (s *PostgresStore) LoadSessions(now time.Time) ([]auth.StoredSession, error) {
	rows, err := s.db.Query(`
		SELECT id, account_id, access_token, access_token_expires_at, refresh_token_hash, expires_at
		FROM auth_sessions
		WHERE revoked_at IS NULL
		  AND access_token IS NOT NULL
		  AND expires_at > $1
		ORDER BY created_at ASC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]auth.StoredSession, 0)
	for rows.Next() {
		var session auth.StoredSession
		if err := rows.Scan(
			&session.SessionID,
			&session.AccountID,
			&session.AccessToken,
			&session.AccessTokenExpiresAt,
			&session.RefreshToken,
			&session.RefreshTokenExpiresAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (s *PostgresStore) SaveSession(stored auth.StoredSession) error {
	now := time.Now().In(s.loc)

	_, err := s.db.Exec(`
		INSERT INTO auth_sessions (
			id, account_id, refresh_token_hash, user_agent, ip_address, expires_at, revoked_at, created_at,
			access_token, access_token_expires_at, updated_at
		)
		VALUES ($1, $2, $3, NULL, NULL, $4, NULL, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET refresh_token_hash = EXCLUDED.refresh_token_hash,
		    expires_at = EXCLUDED.expires_at,
		    revoked_at = NULL,
		    access_token = EXCLUDED.access_token,
		    access_token_expires_at = EXCLUDED.access_token_expires_at,
		    updated_at = EXCLUDED.updated_at
	`, stored.SessionID, stored.AccountID, stored.RefreshToken, stored.RefreshTokenExpiresAt, now,
		stored.AccessToken, stored.AccessTokenExpiresAt, now)
	return err
}

func (s *PostgresStore) DeleteSession(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM auth_sessions WHERE id = $1`, sessionID)
	return err
}

func (s *PostgresStore) LoadChallenges(now time.Time) ([]auth.StoredChallenge, error) {
	rows, err := s.db.Query(`
		SELECT id, prompt_text, answer_format, expected_answer, expires_at, used_at
		FROM auth_challenges
		WHERE expires_at > $1
		ORDER BY created_at ASC
	`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	challenges := make([]auth.StoredChallenge, 0)
	for rows.Next() {
		var challenge auth.StoredChallenge
		var usedAt sql.NullTime
		if err := rows.Scan(
			&challenge.ChallengeID,
			&challenge.PromptText,
			&challenge.AnswerFormat,
			&challenge.ExpectedAnswer,
			&challenge.ExpiresAt,
			&usedAt,
		); err != nil {
			return nil, err
		}
		if usedAt.Valid {
			used := usedAt.Time
			challenge.UsedAt = &used
		}
		challenges = append(challenges, challenge)
	}

	return challenges, rows.Err()
}

func (s *PostgresStore) SaveChallenge(stored auth.StoredChallenge) error {
	now := time.Now().In(s.loc)
	_, err := s.db.Exec(`
		INSERT INTO auth_challenges (
			id, prompt_text, answer_format, expected_answer, expires_at, used_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET prompt_text = EXCLUDED.prompt_text,
		    answer_format = EXCLUDED.answer_format,
		    expected_answer = EXCLUDED.expected_answer,
		    expires_at = EXCLUDED.expires_at,
		    used_at = EXCLUDED.used_at,
		    updated_at = EXCLUDED.updated_at
	`, stored.ChallengeID, stored.PromptText, stored.AnswerFormat, stored.ExpectedAnswer, stored.ExpiresAt, stored.UsedAt, now, now)
	return err
}

func (s *PostgresStore) MarkChallengeUsed(challengeID string, usedAt time.Time) error {
	_, err := s.db.Exec(`
		UPDATE auth_challenges
		SET used_at = $2, updated_at = $2
		WHERE id = $1
	`, challengeID, usedAt)
	return err
}

func (s *PostgresStore) LoadCharacters() ([]characters.StoredCharacter, error) {
	rows, err := s.db.Query(`
		SELECT
			c.account_id,
			c.id,
			c.name,
			c.class,
			COALESCE(c.profession_route_id, ''),
			COALESCE(c.weapon_style, ''),
			COALESCE(c.season_level, 1),
			COALESCE(c.season_xp, 0),
			COALESCE(c.skill_levels_json, '{}'::jsonb),
			COALESCE(c.skill_loadout_json, '[]'::jsonb),
			c.rank,
			c.reputation,
			c.gold,
			c.status,
			c.location_region_id,
			COALESCE(bs.max_hp, 0),
			COALESCE(bs.max_mp, 0),
			COALESCE(bs.physical_attack, 0),
			COALESCE(bs.magic_attack, 0),
			COALESCE(bs.physical_defense, 0),
			COALESCE(bs.magic_defense, 0),
			COALESCE(bs.speed, 0),
			COALESCE(bs.healing_power, 0),
			COALESCE(dl.reset_date::text, ''),
			COALESCE(dl.quest_completion_used, 0),
			COALESCE(dl.dungeon_entry_used, 0),
			COALESCE(dl.dungeon_bonus_purchased, 0)
		FROM characters c
		LEFT JOIN character_base_stats bs ON bs.character_id = c.id
		LEFT JOIN character_daily_limits dl ON dl.character_id = c.id
		ORDER BY c.created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]characters.StoredCharacter, 0)
	for rows.Next() {
		var item characters.StoredCharacter
		var ignoredMaxMP int
		var ignoredRank string
		var skillLevelsJSON []byte
		var skillLoadoutJSON []byte
		if err := rows.Scan(
			&item.AccountID,
			&item.Summary.CharacterID,
			&item.Summary.Name,
			&item.Summary.Class,
			&item.Summary.ProfessionRoute,
			&item.Summary.WeaponStyle,
			&item.Summary.SeasonLevel,
			&item.Summary.SeasonXP,
			&skillLevelsJSON,
			&skillLoadoutJSON,
			&ignoredRank,
			&item.Summary.Reputation,
			&item.Summary.Gold,
			&item.Summary.Status,
			&item.Summary.LocationRegionID,
			&item.Stats.MaxHP,
			&ignoredMaxMP,
			&item.Stats.PhysicalAttack,
			&item.Stats.MagicAttack,
			&item.Stats.PhysicalDefense,
			&item.Stats.MagicDefense,
			&item.Stats.Speed,
			&item.Stats.HealingPower,
			&item.DailyLimitsResetDate,
			&item.QuestCompletionUsed,
			&item.DungeonEntryUsed,
			&item.DungeonBonusPurchased,
		); err != nil {
			return nil, err
		}
		if len(skillLevelsJSON) > 0 {
			if err := json.Unmarshal(skillLevelsJSON, &item.SkillLevels); err != nil {
				return nil, err
			}
		}
		if len(skillLoadoutJSON) > 0 {
			if err := json.Unmarshal(skillLoadoutJSON, &item.SkillLoadout); err != nil {
				return nil, err
			}
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *PostgresStore) LoadRecentEvents(limitPerCharacter int) ([]world.WorldEvent, error) {
	rows, err := s.db.Query(`
		SELECT id, event_type, visibility, COALESCE(actor_character_id, ''), COALESCE(actor_name, ''),
		       COALESCE(region_id, ''), summary, payload_json, occurred_at
		FROM world_events
		ORDER BY occurred_at DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]world.WorldEvent, 0)
	perCharacterCount := make(map[string]int)
	for rows.Next() {
		var event world.WorldEvent
		var payloadBytes []byte
		var occurredAt time.Time

		if err := rows.Scan(
			&event.EventID,
			&event.EventType,
			&event.Visibility,
			&event.ActorCharacterID,
			&event.ActorName,
			&event.RegionID,
			&event.Summary,
			&payloadBytes,
			&occurredAt,
		); err != nil {
			return nil, err
		}

		if event.ActorCharacterID != "" && limitPerCharacter > 0 {
			if perCharacterCount[event.ActorCharacterID] >= limitPerCharacter {
				continue
			}
			perCharacterCount[event.ActorCharacterID]++
		}

		if len(payloadBytes) > 0 {
			if err := json.Unmarshal(payloadBytes, &event.Payload); err != nil {
				return nil, err
			}
		} else {
			event.Payload = map[string]any{}
		}
		event.OccurredAt = occurredAt.In(s.loc).Format(time.RFC3339)
		events = append(events, event)
	}

	return events, rows.Err()
}

func (s *PostgresStore) SaveCharacter(stored characters.StoredCharacter) error {
	now := time.Now().In(s.loc)
	resetDate := stored.DailyLimitsResetDate
	if resetDate == "" {
		resetDate = businessDate(now)
	}
	limits := characters.BuildDailyLimits(nextDailyReset(now), stored.QuestCompletionUsed, stored.DungeonEntryUsed, stored.DungeonBonusPurchased)
	skillLevelsJSON, err := json.Marshal(stored.SkillLevels)
	if err != nil {
		return err
	}
	skillLoadoutJSON, err := json.Marshal(stored.SkillLoadout)
	if err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO characters (
			id, account_id, name, class, profession_route_id, weapon_style, season_level, season_xp, skill_levels_json, skill_loadout_json,
			rank, reputation, gold, status, location_region_id, hp_current, mp_current, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		ON CONFLICT (id) DO UPDATE
		SET name = EXCLUDED.name,
		    class = EXCLUDED.class,
		    profession_route_id = EXCLUDED.profession_route_id,
		    weapon_style = EXCLUDED.weapon_style,
		    season_level = EXCLUDED.season_level,
		    season_xp = EXCLUDED.season_xp,
		    skill_levels_json = EXCLUDED.skill_levels_json,
		    skill_loadout_json = EXCLUDED.skill_loadout_json,
		    rank = EXCLUDED.rank,
		    reputation = EXCLUDED.reputation,
		    gold = EXCLUDED.gold,
		    status = EXCLUDED.status,
		    location_region_id = EXCLUDED.location_region_id,
		    hp_current = EXCLUDED.hp_current,
		    mp_current = EXCLUDED.mp_current,
		    updated_at = EXCLUDED.updated_at
	`, stored.Summary.CharacterID, stored.AccountID, stored.Summary.Name, stored.Summary.Class, nullableString(stored.Summary.ProfessionRoute), nullableString(stored.Summary.WeaponStyle), stored.Summary.SeasonLevel, stored.Summary.SeasonXP,
		skillLevelsJSON, skillLoadoutJSON, "", stored.Summary.Reputation, stored.Summary.Gold, stored.Summary.Status, stored.Summary.LocationRegionID,
		stored.Stats.MaxHP, 0, now, now); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO character_base_stats (
			character_id, max_hp, max_mp, physical_attack, magic_attack,
			physical_defense, magic_defense, speed, healing_power, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (character_id) DO UPDATE
		SET max_hp = EXCLUDED.max_hp,
		    max_mp = EXCLUDED.max_mp,
		    physical_attack = EXCLUDED.physical_attack,
		    magic_attack = EXCLUDED.magic_attack,
		    physical_defense = EXCLUDED.physical_defense,
		    magic_defense = EXCLUDED.magic_defense,
		    speed = EXCLUDED.speed,
		    healing_power = EXCLUDED.healing_power,
		    updated_at = EXCLUDED.updated_at
	`, stored.Summary.CharacterID, stored.Stats.MaxHP, 0, stored.Stats.PhysicalAttack, stored.Stats.MagicAttack,
		stored.Stats.PhysicalDefense, stored.Stats.MagicDefense, stored.Stats.Speed, stored.Stats.HealingPower, now); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO character_daily_limits (
			character_id, reset_date, quest_completion_cap, quest_completion_used,
			dungeon_entry_cap, dungeon_entry_used, dungeon_bonus_purchased, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (character_id) DO UPDATE
		SET reset_date = EXCLUDED.reset_date,
		    quest_completion_cap = EXCLUDED.quest_completion_cap,
		    quest_completion_used = EXCLUDED.quest_completion_used,
		    dungeon_entry_cap = EXCLUDED.dungeon_entry_cap,
		    dungeon_entry_used = EXCLUDED.dungeon_entry_used,
		    dungeon_bonus_purchased = EXCLUDED.dungeon_bonus_purchased,
		    updated_at = EXCLUDED.updated_at
	`, stored.Summary.CharacterID, resetDate, limits.QuestCompletionCap, stored.QuestCompletionUsed,
		limits.DungeonEntryCap, stored.DungeonEntryUsed, stored.DungeonBonusPurchased, now); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStore) AppendWorldEvents(accountID string, characterID string, events []world.WorldEvent) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, event := range events {
		payload := event.Payload
		if payload == nil {
			payload = map[string]any{}
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		occurredAt := parseRFC3339InLocation(event.OccurredAt, s.loc)
		if _, err := tx.Exec(`
			INSERT INTO world_events (
				id, event_type, visibility, actor_account_id, actor_character_id, actor_name,
				region_id, related_entity_type, related_entity_id, summary, payload_json, occurred_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), NULL, NULL, $8, $9, $10)
			ON CONFLICT (id) DO NOTHING
		`, event.EventID, event.EventType, event.Visibility, accountID, characterID, event.ActorName,
			event.RegionID, event.Summary, payloadBytes, occurredAt); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStore) LoadBoards() ([]quests.StoredBoard, error) {
	rows, err := s.db.Query(`
		SELECT id, character_id, reset_date::text, status, reroll_count
		FROM quest_boards
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boardsByID := make(map[string]quests.StoredBoard)
	boardOrder := make([]string, 0)
	for rows.Next() {
		var board quests.StoredBoard
		if err := rows.Scan(&board.BoardID, &board.CharacterID, &board.ResetDate, &board.Status, &board.RerollCount); err != nil {
			return nil, err
		}
		board.Quests = []characters.QuestSummary{}
		boardsByID[board.BoardID] = board
		boardOrder = append(boardOrder, board.BoardID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	questRows, err := s.db.Query(`
		SELECT id, board_id, template_type, COALESCE(difficulty, ''), COALESCE(flow_kind, ''),
		       rarity, status, title, description, COALESCE(target_region_id, ''),
		       progress_current, progress_target, reward_gold, reward_reputation,
		       COALESCE(runtime_state_json, '{}'::jsonb)
		FROM quests
		ORDER BY board_id ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer questRows.Close()

	for questRows.Next() {
		var quest characters.QuestSummary
		var boardID string
		var runtimeStateJSON []byte
		if err := questRows.Scan(
			&quest.QuestID,
			&boardID,
			&quest.TemplateType,
			&quest.Difficulty,
			&quest.FlowKind,
			&quest.Rarity,
			&quest.Status,
			&quest.Title,
			&quest.Description,
			&quest.TargetRegionID,
			&quest.ProgressCurrent,
			&quest.ProgressTarget,
			&quest.RewardGold,
			&quest.RewardReputation,
			&runtimeStateJSON,
		); err != nil {
			return nil, err
		}

		board, ok := boardsByID[boardID]
		if !ok {
			continue
		}
		quest.BoardID = boardID
		board.Quests = append(board.Quests, quest)
		if len(runtimeStateJSON) > 0 && string(runtimeStateJSON) != "null" && string(runtimeStateJSON) != "{}" {
			if board.RuntimeByQuest == nil {
				board.RuntimeByQuest = make(map[string]quests.StoredQuestRuntime)
			}
			var runtime quests.StoredQuestRuntime
			if err := json.Unmarshal(runtimeStateJSON, &runtime); err != nil {
				return nil, err
			}
			board.RuntimeByQuest[quest.QuestID] = runtime
		}
		boardsByID[boardID] = board
	}
	if err := questRows.Err(); err != nil {
		return nil, err
	}

	items := make([]quests.StoredBoard, 0, len(boardOrder))
	for _, boardID := range boardOrder {
		items = append(items, boardsByID[boardID])
	}
	return items, nil
}

func (s *PostgresStore) SaveBoard(board quests.StoredBoard) error {
	now := time.Now().In(s.loc)
	expiresAt := nextResetFromBusinessDate(board.ResetDate, s.loc)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO quest_boards (id, character_id, reset_date, status, reroll_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET status = EXCLUDED.status,
		    reroll_count = EXCLUDED.reroll_count,
		    updated_at = EXCLUDED.updated_at
	`, board.BoardID, board.CharacterID, board.ResetDate, board.Status, board.RerollCount, now, now); err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM quests WHERE board_id = $1`, board.BoardID); err != nil {
		return err
	}

	for _, quest := range board.Quests {
		runtimeStateJSON := []byte(`{}`)
		if runtime, ok := board.RuntimeByQuest[quest.QuestID]; ok {
			payload, err := json.Marshal(runtime)
			if err != nil {
				return err
			}
			runtimeStateJSON = payload
		}
		if _, err := tx.Exec(`
			INSERT INTO quests (
				id, board_id, character_id, template_type, difficulty, flow_kind, rarity, status, title, description,
				target_region_id, target_dungeon_id, target_enemy_key, progress_current, progress_target,
				reward_gold, reward_reputation, reward_item_catalog_id, accepted_at, completed_at,
				submitted_at, expires_at, runtime_state_json
			)
			VALUES (
				$1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9, $10,
				NULLIF($11, ''), NULL, NULL, $12, $13,
				$14, $15, NULL, NULL, NULL, NULL, $16, $17
			)
		`, quest.QuestID, board.BoardID, board.CharacterID, quest.TemplateType, quest.Difficulty, quest.FlowKind,
			quest.Rarity, quest.Status, quest.Title, quest.Description, quest.TargetRegionID, quest.ProgressCurrent,
			quest.ProgressTarget, quest.RewardGold, quest.RewardReputation, expiresAt, runtimeStateJSON); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func parseRFC3339InLocation(raw string, loc *time.Location) time.Time {
	if raw == "" {
		return time.Now().In(loc)
	}

	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Now().In(loc)
	}

	return parsed.In(loc)
}

func businessDate(now time.Time) string {
	if now.Hour() < 4 {
		now = now.Add(-24 * time.Hour)
	}

	return now.Format("2006-01-02")
}

func nextDailyReset(now time.Time) time.Time {
	resetToday := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	if now.Before(resetToday) {
		return resetToday
	}

	return resetToday.Add(24 * time.Hour)
}

func nextResetFromBusinessDate(resetDate string, loc *time.Location) time.Time {
	date, err := time.ParseInLocation("2006-01-02", resetDate, loc)
	if err != nil {
		return nextDailyReset(time.Now().In(loc))
	}

	return time.Date(date.Year(), date.Month(), date.Day(), 4, 0, 0, 0, loc).Add(24 * time.Hour)
}

func mustLocation(name string) *time.Location {
	location, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}

	return location
}

func (s *PostgresStore) String() string {
	return fmt.Sprintf("postgres:%p", s.db)
}

func (s *PostgresStore) ensureSchema() error {
	statements := []string{
		`ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS access_token text`,
		`ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS access_token_expires_at timestamptz`,
		`ALTER TABLE auth_sessions ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT NOW()`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS profession_route_id text`,
		`ALTER TABLE characters ALTER COLUMN weapon_style DROP NOT NULL`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS season_level integer NOT NULL DEFAULT 1`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS season_xp integer NOT NULL DEFAULT 0`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS skill_levels_json jsonb NOT NULL DEFAULT '{}'::jsonb`,
		`ALTER TABLE characters ADD COLUMN IF NOT EXISTS skill_loadout_json jsonb NOT NULL DEFAULT '[]'::jsonb`,
		`ALTER TABLE quests ADD COLUMN IF NOT EXISTS difficulty text`,
		`ALTER TABLE quests ADD COLUMN IF NOT EXISTS flow_kind text`,
		`ALTER TABLE quests ADD COLUMN IF NOT EXISTS runtime_state_json jsonb NOT NULL DEFAULT '{}'::jsonb`,
		`ALTER TABLE character_daily_limits ADD COLUMN IF NOT EXISTS dungeon_bonus_purchased integer NOT NULL DEFAULT 0`,
		`CREATE INDEX IF NOT EXISTS idx_auth_sessions_access_token ON auth_sessions(access_token)`,
		`CREATE TABLE IF NOT EXISTS auth_challenges (
			id text PRIMARY KEY,
			prompt_text text NOT NULL,
			answer_format text NOT NULL,
			expected_answer text NOT NULL,
			expires_at timestamptz NOT NULL,
			used_at timestamptz,
			created_at timestamptz NOT NULL,
			updated_at timestamptz NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_challenges_expires_at ON auth_challenges(expires_at)`,
	}

	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}

	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
