---
name: clawgame-docs-governance
description: Use this skill when reading, writing, translating, or updating ClawGame documentation. Prefer English docs under docs/en as the source of truth, update English first, then sync the corresponding Chinese docs under docs/zh. Also use this skill for requests about 文档, docs, 英文优先, 中文翻译, or documentation consistency.
---

# ClawGame Documentation Governance

Use this skill for any task that touches project documentation.

## Source Of Truth

- Treat `docs/en/` as the primary documentation set.
- Read English documentation first unless the user explicitly asks to inspect Chinese wording first.
- Treat `docs/zh/` as the synchronized translation set, not the primary source for new product or technical decisions.
- Treat root-level files in `docs/` as compatibility or legacy copies unless the task explicitly requires them.

## Required Workflow

1. Find the relevant English file in `docs/en/` and use it as the baseline for understanding the topic.
2. When updating docs, edit the English file first.
3. After the English version is correct, update the matching Chinese file under `docs/zh/`.
4. Keep both language versions aligned in structure, headings, code blocks, tables, and file references unless a language-specific wording adjustment is necessary.
5. If only one language version exists, add the missing counterpart when the task clearly calls for a maintained bilingual document set.

## Translation Rules

- Translate meaning, not word-for-word phrasing.
- Keep product names, API paths, commands, env vars, and schema fields unchanged.
- Preserve examples, lists, and section ordering unless the English source was intentionally reorganized.
- Keep terminology consistent across the whole docs tree. If an existing Chinese term is already used broadly, reuse it unless it is clearly wrong.
- If a sentence is ambiguous in Chinese, fix the English source first if the ambiguity comes from the source text.

## Editing Guidance

- Prefer adding new documentation under `docs/en/` and `docs/zh/`.
- When a request references a root-level doc in `docs/`, check whether an equivalent canonical file already exists in `docs/en/` and `docs/zh/` before editing.
- Avoid letting English and Chinese versions drift. Do not leave a completed English doc unsynchronized if the Chinese counterpart is part of the maintained set.

## Completion Checklist

- Confirm the English file was reviewed first.
- Confirm English was updated before Chinese.
- Confirm the Chinese file matches the final English structure and intent.
- Mention any intentionally unsynchronized files in the final handoff.
