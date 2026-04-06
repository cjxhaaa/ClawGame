package quests

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed configs/**/*.yaml
var questConfigFS embed.FS

type questTemplateCatalog struct {
	all          []questTemplateDefinition
	daily        []questTemplateDefinition
	supplemental []questTemplateDefinition
}

type questTemplateConfigFile struct {
	Pool             string                 `yaml:"pool"`
	Order            int                    `yaml:"order"`
	TemplateType     string                 `yaml:"template_type"`
	Difficulty       string                 `yaml:"difficulty"`
	FlowKind         string                 `yaml:"flow_kind"`
	Rarity           string                 `yaml:"rarity"`
	Title            string                 `yaml:"title"`
	Description      string                 `yaml:"description"`
	TargetRegionID   string                 `yaml:"target_region_id"`
	ProgressTarget   int                    `yaml:"progress_target"`
	RewardGold       int                    `yaml:"reward_gold"`
	RewardReputation int                    `yaml:"reward_reputation"`
	Runtime          questRuntimeConfigFile `yaml:"runtime"`
}

type questRuntimeConfigFile struct {
	InitialStepKey       string                 `yaml:"initial_step_key"`
	CompletionStepKey    string                 `yaml:"completion_step_key"`
	ChoiceStepKey        string                 `yaml:"choice_step_key"`
	InspectStepKey       string                 `yaml:"inspect_step_key"`
	ProgressTriggerType  string                 `yaml:"progress_trigger_type"`
	ProgressSource       string                 `yaml:"progress_source"`
	RequiresChoice       bool                   `yaml:"requires_choice"`
	RequiresInspection   bool                   `yaml:"requires_inspection"`
	RequiresRouteConfirm bool                   `yaml:"requires_route_confirm"`
	BaseClues            []QuestClue            `yaml:"base_clues"`
	Steps                []questStepSpec        `yaml:"steps"`
	InteractionSpecs     []questInteractionSpec `yaml:"interaction_specs"`
	ChoiceSpecs          []questChoiceSpec      `yaml:"choice_specs"`
}

var (
	questCatalogOnce sync.Once
	questCatalogData questTemplateCatalog
	questCatalogErr  error
)

func defaultQuestTemplateCatalog() (questTemplateCatalog, error) {
	questCatalogOnce.Do(func() {
		questCatalogData, questCatalogErr = loadQuestTemplateCatalog()
	})
	return questCatalogData, questCatalogErr
}

func mustQuestTemplateCatalog() questTemplateCatalog {
	catalog, err := defaultQuestTemplateCatalog()
	if err != nil {
		panic(err)
	}
	return catalog
}

func loadQuestTemplateCatalog() (questTemplateCatalog, error) {
	entries := make([]questTemplateDefinition, 0, 8)
	err := fs.WalkDir(questConfigFS, "configs", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		content, err := fs.ReadFile(questConfigFS, path)
		if err != nil {
			return fmt.Errorf("read quest config %s: %w", path, err)
		}

		var file questTemplateConfigFile
		if err := yaml.Unmarshal(content, &file); err != nil {
			return fmt.Errorf("decode quest config %s: %w", path, err)
		}

		definition, err := buildQuestTemplateDefinition(path, file)
		if err != nil {
			return err
		}
		entries = append(entries, definition)
		return nil
	})
	if err != nil {
		return questTemplateCatalog{}, err
	}
	if err := validateQuestTemplateDefinitions(entries); err != nil {
		return questTemplateCatalog{}, err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Pool == entries[j].Pool {
			if entries[i].Order == entries[j].Order {
				if entries[i].TemplateType == entries[j].TemplateType {
					return entries[i].Difficulty < entries[j].Difficulty
				}
				return entries[i].TemplateType < entries[j].TemplateType
			}
			return entries[i].Order < entries[j].Order
		}
		return entries[i].Pool < entries[j].Pool
	})

	catalog := questTemplateCatalog{
		all: make([]questTemplateDefinition, 0, len(entries)),
	}
	for _, definition := range entries {
		catalog.all = append(catalog.all, definition)
		switch definition.Pool {
		case "daily":
			catalog.daily = append(catalog.daily, definition)
		case "supplemental":
			catalog.supplemental = append(catalog.supplemental, definition)
		default:
			return questTemplateCatalog{}, fmt.Errorf("quest template %s uses unsupported pool %q", definition.TemplateType, definition.Pool)
		}
	}
	return catalog, nil
}

func buildQuestTemplateDefinition(path string, file questTemplateConfigFile) (questTemplateDefinition, error) {
	if strings.TrimSpace(file.Pool) == "" {
		return questTemplateDefinition{}, fmt.Errorf("quest config %s missing pool", path)
	}
	if strings.TrimSpace(file.TemplateType) == "" {
		return questTemplateDefinition{}, fmt.Errorf("quest config %s missing template_type", path)
	}
	if strings.TrimSpace(file.Difficulty) == "" {
		return questTemplateDefinition{}, fmt.Errorf("quest config %s missing difficulty", path)
	}
	if strings.TrimSpace(file.FlowKind) == "" {
		return questTemplateDefinition{}, fmt.Errorf("quest config %s missing flow_kind", path)
	}
	if len(file.Runtime.Steps) == 0 {
		return questTemplateDefinition{}, fmt.Errorf("quest config %s missing runtime.steps", path)
	}
	if strings.TrimSpace(file.Runtime.InitialStepKey) == "" || strings.TrimSpace(file.Runtime.CompletionStepKey) == "" {
		return questTemplateDefinition{}, fmt.Errorf("quest config %s missing runtime initial/completion step", path)
	}

	return questTemplateDefinition{
		TemplateType:     strings.TrimSpace(file.TemplateType),
		Pool:             strings.TrimSpace(file.Pool),
		Order:            file.Order,
		Difficulty:       strings.TrimSpace(file.Difficulty),
		FlowKind:         strings.TrimSpace(file.FlowKind),
		Rarity:           strings.TrimSpace(file.Rarity),
		Title:            file.Title,
		Description:      file.Description,
		TargetRegionID:   strings.TrimSpace(file.TargetRegionID),
		ProgressTarget:   file.ProgressTarget,
		RewardGold:       file.RewardGold,
		RewardReputation: file.RewardReputation,
		Spec: questRuntimeSpec{
			InitialStepKey:       strings.TrimSpace(file.Runtime.InitialStepKey),
			CompletionStepKey:    strings.TrimSpace(file.Runtime.CompletionStepKey),
			ChoiceStepKey:        strings.TrimSpace(file.Runtime.ChoiceStepKey),
			InspectStepKey:       strings.TrimSpace(file.Runtime.InspectStepKey),
			ProgressTriggerType:  strings.TrimSpace(file.Runtime.ProgressTriggerType),
			ProgressSource:       strings.TrimSpace(file.Runtime.ProgressSource),
			RequiresChoice:       file.Runtime.RequiresChoice,
			RequiresInspection:   file.Runtime.RequiresInspection,
			RequiresRouteConfirm: file.Runtime.RequiresRouteConfirm,
			ChoiceSpecs:          file.Runtime.ChoiceSpecs,
			InteractionSpecs:     file.Runtime.InteractionSpecs,
			BaseClues:            file.Runtime.BaseClues,
			Steps:                file.Runtime.Steps,
		},
	}, nil
}

func validateQuestTemplateDefinitions(definitions []questTemplateDefinition) error {
	seen := make(map[string]string, len(definitions))
	for _, definition := range definitions {
		key := strings.Join([]string{
			strings.TrimSpace(definition.Pool),
			strings.TrimSpace(definition.TemplateType),
			strings.TrimSpace(definition.Difficulty),
			strings.TrimSpace(definition.FlowKind),
		}, "::")
		if existing, ok := seen[key]; ok {
			return fmt.Errorf("duplicate quest template definition %s and %s share key %s", existing, definition.TemplateType, key)
		}
		seen[key] = definition.TemplateType
		if err := validateQuestTemplateDefinition(definition); err != nil {
			return err
		}
	}
	return nil
}

func validateQuestTemplateDefinition(definition questTemplateDefinition) error {
	stepKeys := make(map[string]struct{}, len(definition.Spec.Steps))
	for _, step := range definition.Spec.Steps {
		key := strings.TrimSpace(step.Key)
		if key == "" {
			return fmt.Errorf("quest template %s has empty step key", definition.TemplateType)
		}
		if _, ok := stepKeys[key]; ok {
			return fmt.Errorf("quest template %s repeats step key %s", definition.TemplateType, key)
		}
		stepKeys[key] = struct{}{}
	}
	for _, required := range []string{definition.Spec.InitialStepKey, definition.Spec.CompletionStepKey} {
		if strings.TrimSpace(required) == "" {
			continue
		}
		if _, ok := stepKeys[strings.TrimSpace(required)]; !ok {
			return fmt.Errorf("quest template %s references missing step %s", definition.TemplateType, required)
		}
	}
	if strings.TrimSpace(definition.Spec.ChoiceStepKey) != "" {
		if _, ok := stepKeys[strings.TrimSpace(definition.Spec.ChoiceStepKey)]; !ok {
			return fmt.Errorf("quest template %s choice_step_key %s is missing from steps", definition.TemplateType, definition.Spec.ChoiceStepKey)
		}
	}
	if strings.TrimSpace(definition.Spec.InspectStepKey) != "" {
		if _, ok := stepKeys[strings.TrimSpace(definition.Spec.InspectStepKey)]; !ok {
			return fmt.Errorf("quest template %s inspect_step_key %s is missing from steps", definition.TemplateType, definition.Spec.InspectStepKey)
		}
	}
	if definition.Spec.RequiresChoice && len(definition.Spec.ChoiceSpecs) == 0 {
		return fmt.Errorf("quest template %s requires choice but defines no choice_specs", definition.TemplateType)
	}
	for _, interaction := range definition.Spec.InteractionSpecs {
		if _, ok := stepKeys[strings.TrimSpace(interaction.StepKey)]; !ok {
			return fmt.Errorf("quest template %s interaction step %s is missing from steps", definition.TemplateType, interaction.StepKey)
		}
		if strings.TrimSpace(interaction.NextStepKey) != "" {
			if _, ok := stepKeys[strings.TrimSpace(interaction.NextStepKey)]; !ok {
				return fmt.Errorf("quest template %s interaction %s points to missing next step %s", definition.TemplateType, interaction.StepKey, interaction.NextStepKey)
			}
		}
	}
	for _, choice := range definition.Spec.ChoiceSpecs {
		if strings.TrimSpace(choice.ChoiceKey) == "" {
			return fmt.Errorf("quest template %s has choice with empty choice_key", definition.TemplateType)
		}
		if strings.TrimSpace(choice.NextStepKey) != "" {
			if _, ok := stepKeys[strings.TrimSpace(choice.NextStepKey)]; !ok {
				return fmt.Errorf("quest template %s choice %s points to missing next step %s", definition.TemplateType, choice.ChoiceKey, choice.NextStepKey)
			}
		}
	}
	return nil
}
