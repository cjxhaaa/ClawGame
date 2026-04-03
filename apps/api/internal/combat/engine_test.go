package combat

import "testing"

func TestHealSkillTriggersWhenLowHP(t *testing.T) {
	player := BaselineCombatant("priest")
	player.EntityID = "player"
	player.Name = "healer"
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = 18
	player.EquippedSkills = []SkillAction{
		{SkillID: "Restore", ActionKind: "heal", CooldownRounds: 1, HealScale: 1.0, Level: 1},
	}
	player.SkillCooldowns = map[string]int{}

	enemy := BaselineCombatant("warrior")
	enemy.EntityID = "enemy"
	enemy.Name = "foe"
	enemy.Team = "b"
	enemy.CurrentHP = enemy.MaxHP

	result := SimulateBattle(BattleConfig{
		BattleType: "dungeon_wave",
		RunID:      "heal-test",
		RoomIndex:  1,
		MaxTurns:   1,
		SideA:      player,
		SideB:      enemy,
	})

	foundHeal := false
	for _, entry := range result.Log {
		if entry["skill_id"] == "Restore" && entry["value_type"] == "heal" {
			foundHeal = true
			break
		}
	}
	if !foundHeal {
		t.Fatal("expected Restore heal action in battle log")
	}
}

func TestSilencePreventsEnemyBurstSkill(t *testing.T) {
	player := BaselineCombatant("mage")
	player.EntityID = "player"
	player.Name = "controller"
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = player.MaxHP
	player.EquippedSkills = []SkillAction{
		{SkillID: "Silencing Prism", ActionKind: "debuff", CooldownRounds: 2, SilenceRounds: 2, Level: 1},
	}
	player.SkillCooldowns = map[string]int{}

	enemy := BaselineCombatant("warrior")
	enemy.EntityID = "enemy"
	enemy.Name = "foe"
	enemy.Team = "b"
	enemy.CurrentHP = enemy.MaxHP

	result := SimulateBattle(BattleConfig{
		BattleType: "dungeon_wave",
		RunID:      "silence-test",
		RoomIndex:  1,
		MaxTurns:   1,
		SideA:      player,
		SideB:      enemy,
	})

	foundSilence := false
	enemyUsedBurst := false
	enemyUsedBasic := false
	for _, entry := range result.Log {
		if entry["skill_id"] == "Silencing Prism" {
			foundSilence = true
		}
		if entry["actor"] == "b" && entry["action"] == "enemy_skill" {
			enemyUsedBurst = true
		}
		if entry["actor"] == "b" && entry["action"] == "enemy_attack" {
			enemyUsedBasic = true
		}
	}
	if !foundSilence {
		t.Fatal("expected Silencing Prism action in battle log")
	}
	if enemyUsedBurst {
		t.Fatal("expected silence to prevent enemy burst skill")
	}
	if !enemyUsedBasic {
		t.Fatal("expected enemy to fall back to basic attack while silenced")
	}
}

func TestSummonSkillProducesSummonTick(t *testing.T) {
	player := BaselineCombatant("priest")
	player.EntityID = "player"
	player.Name = "summoner"
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = player.MaxHP
	player.EquippedSkills = []SkillAction{
		{SkillID: "Choir Invocation", ActionKind: "summon", CooldownRounds: 3, SummonTurns: 3, SummonStrike: 0.7, SummonMend: 0.15, Level: 1},
	}
	player.SkillCooldowns = map[string]int{}

	enemy := BaselineCombatant("warrior")
	enemy.EntityID = "enemy"
	enemy.Name = "foe"
	enemy.Team = "b"
	enemy.CurrentHP = enemy.MaxHP

	result := SimulateBattle(BattleConfig{
		BattleType: "dungeon_wave",
		RunID:      "summon-test",
		RoomIndex:  1,
		MaxTurns:   2,
		SideA:      player,
		SideB:      enemy,
	})

	foundSummonTick := false
	for _, entry := range result.Log {
		if entry["event_type"] == "summon_tick" {
			foundSummonTick = true
			break
		}
	}
	if !foundSummonTick {
		t.Fatal("expected summon tick event in battle log")
	}
}

func TestShieldAbsorbsIncomingDamage(t *testing.T) {
	player := BaselineCombatant("warrior")
	player.EntityID = "player"
	player.Name = "tank"
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = player.MaxHP
	player.EquippedSkills = []SkillAction{
		{SkillID: "Guard Stance", ActionKind: "buff", CooldownRounds: 1, BuffFamily: "def", BuffValue: 0.16, ShieldScale: 0.20, Level: 1},
	}
	player.SkillCooldowns = map[string]int{}

	enemy := BaselineCombatant("warrior")
	enemy.EntityID = "enemy"
	enemy.Name = "foe"
	enemy.Team = "b"
	enemy.CurrentHP = enemy.MaxHP

	result := SimulateBattle(BattleConfig{
		BattleType: "dungeon_wave",
		RunID:      "shield-test",
		RoomIndex:  1,
		MaxTurns:   1,
		SideA:      player,
		SideB:      enemy,
	})

	foundShield := false
	for _, entry := range result.Log {
		if entry["skill_id"] == "Guard Stance" {
			foundShield = true
			if after, ok := entry["target_shield_after"].(int); ok && after <= 0 {
				t.Fatal("expected Guard Stance to grant shield")
			}
		}
	}
	if !foundShield {
		t.Fatal("expected Guard Stance action in battle log")
	}
}

func TestSkirmishAOESkillHitsMultipleEnemies(t *testing.T) {
	player := BaselineCombatant("warrior")
	player.EntityID = "player"
	player.Name = "sweeper"
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = player.MaxHP
	player.PhysAtk = 80
	player.EquippedSkills = []SkillAction{
		{SkillID: "Cleave", ActionKind: "damage", DamageType: "physical", TargetPattern: "all_enemies", CooldownRounds: 1, PowerScale: 1.20, Level: 1},
	}
	player.SkillCooldowns = map[string]int{}

	enemyA := BaselineCombatant("warrior")
	enemyA.EntityID = "enemy_a"
	enemyA.Name = "foe_a"
	enemyA.Team = "b"
	enemyA.Role = "bruiser"
	enemyA.CurrentHP = 20
	enemyA.MaxHP = 20
	enemyA.PhysDef = 4

	enemyB := BaselineCombatant("mage")
	enemyB.EntityID = "enemy_b"
	enemyB.Name = "foe_b"
	enemyB.Team = "b"
	enemyB.Role = "caster"
	enemyB.CurrentHP = 18
	enemyB.MaxHP = 18
	enemyB.PhysDef = 3

	result := SimulateSkirmish(SkirmishConfig{
		BattleType: "dungeon_wave",
		RunID:      "aoe-test",
		RoomIndex:  1,
		MaxTurns:   1,
		Player:     player,
		Enemies:    []Combatant{enemyA, enemyB},
	})

	if !result.PlayerWon {
		t.Fatal("expected player to clear both enemies with aoe skill")
	}

	hitTargets := map[string]bool{}
	for _, entry := range result.Log {
		if entry["skill_id"] == "Cleave" {
			if targetID, ok := entry["target_entity_id"].(string); ok {
				hitTargets[targetID] = true
			}
		}
	}
	if len(hitTargets) != 2 {
		t.Fatalf("expected Cleave to hit both enemies, got %d targets", len(hitTargets))
	}
}

func TestEnemyUsesEquippedSkillInsteadOfFallbackBurst(t *testing.T) {
	player := BaselineCombatant("warrior")
	player.EntityID = "player"
	player.Name = "frontliner"
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = player.MaxHP

	enemy := BaselineCombatant("mage")
	enemy.EntityID = "enemy"
	enemy.Name = "boss"
	enemy.Team = "b"
	enemy.CurrentHP = enemy.MaxHP
	enemy.EquippedSkills = []SkillAction{
		{SkillID: "void_lance", ActionKind: "damage", DamageType: "magic", CooldownRounds: 2, PowerScale: 1.4, Level: 1},
	}
	enemy.SkillCooldowns = map[string]int{}

	result := SimulateBattle(BattleConfig{
		BattleType: "world_boss",
		RunID:      "enemy-skill-test",
		RoomIndex:  1,
		MaxTurns:   1,
		SideA:      player,
		SideB:      enemy,
	})

	foundBossSkill := false
	foundFallback := false
	for _, entry := range result.Log {
		if entry["skill_id"] == "void_lance" {
			foundBossSkill = true
		}
		if entry["skill_id"] == "enemy_burst_skill_mage" {
			foundFallback = true
		}
	}
	if !foundBossSkill {
		t.Fatal("expected enemy equipped skill to execute")
	}
	if foundFallback {
		t.Fatal("expected enemy equipped skill to replace fallback burst skill")
	}
}

func TestHalfHPOnlySkillWaitsUntilBelowHalf(t *testing.T) {
	player := BaselineCombatant("warrior")
	player.EntityID = "player"
	player.Name = "frontliner"
	player.Team = "a"
	player.IsPlayerSide = true
	player.CurrentHP = player.MaxHP

	enemy := BaselineCombatant("mage")
	enemy.EntityID = "enemy"
	enemy.Name = "boss"
	enemy.Team = "b"
	enemy.CurrentHP = enemy.MaxHP
	enemy.EquippedSkills = []SkillAction{
		{SkillID: "spires_end_requiem", ActionKind: "damage", DamageType: "magic", CooldownRounds: 8, PowerScale: 1.9, HalfHPOnly: true, Level: 1},
	}
	enemy.SkillCooldowns = map[string]int{}

	result := SimulateBattle(BattleConfig{
		BattleType: "world_boss",
		RunID:      "half-hp-skill-test",
		RoomIndex:  1,
		MaxTurns:   1,
		SideA:      player,
		SideB:      enemy,
	})

	for _, entry := range result.Log {
		if entry["skill_id"] == "spires_end_requiem" {
			t.Fatal("expected half-hp-only skill to stay unavailable while boss is above 50% hp")
		}
	}
}

func TestRaidBossAOEHitsMultiplePlayersInSharedBattle(t *testing.T) {
	playerA := BaselineCombatant("warrior")
	playerA.EntityID = "player_a"
	playerA.Name = "alpha"
	playerA.Team = "a"
	playerA.IsPlayerSide = true
	playerA.CurrentHP = playerA.MaxHP

	playerB := BaselineCombatant("priest")
	playerB.EntityID = "player_b"
	playerB.Name = "beta"
	playerB.Team = "a"
	playerB.IsPlayerSide = true
	playerB.CurrentHP = playerB.MaxHP

	boss := BaselineCombatant("mage")
	boss.EntityID = "boss"
	boss.Name = "raid boss"
	boss.Team = "b"
	boss.CurrentHP = boss.MaxHP
	boss.Speed = 99
	boss.EquippedSkills = []SkillAction{
		{SkillID: "eclipse_field", ActionKind: "damage", DamageType: "magic", TargetPattern: "all_players", CooldownRounds: 5, PowerScale: 1.0, Level: 1},
	}
	boss.SkillCooldowns = map[string]int{}

	result := SimulateRaidBoss(RaidBossConfig{
		BattleType: "world_boss",
		RunID:      "raid-aoe-test",
		MaxTurns:   1,
		Players:    []Combatant{playerA, playerB},
		Boss:       boss,
	})

	hitTargets := map[string]bool{}
	for _, entry := range result.Log {
		if entry["skill_id"] == "eclipse_field" {
			if targetID, ok := entry["target_entity_id"].(string); ok {
				hitTargets[targetID] = true
			}
		}
	}
	if !hitTargets["player_a"] || !hitTargets["player_b"] {
		t.Fatalf("expected eclipse_field to hit both raid players, got %#v", hitTargets)
	}
}
