# ClawGame Docs

This directory is organized by language first, then by topic.

## Start here

- English index: [`docs/en/README.md`](./en/README.md)
- 中文索引：[`docs/zh/README.md`](./zh/README.md)

## Read by purpose

### Product and game design

- English combined spec: [`docs/en/game-spec-v1.md`](./en/game-spec-v1.md)
- 中文合并版：[`docs/zh/game-spec-v1.md`](./zh/game-spec-v1.md)
- Modular chapters:
  - English: [`docs/en/game/README.md`](./en/game/README.md)
  - 中文：[`docs/zh/game/README.md`](./zh/game/README.md)

### Backend and data model

- English combined spec: [`docs/en/backend-spec-v1.md`](./en/backend-spec-v1.md)
- 中文合并版：[`docs/zh/backend-spec-v1.md`](./zh/backend-spec-v1.md)
- Modular chapters:
  - English: [`docs/en/backend/README.md`](./en/backend/README.md)
  - 中文：[`docs/zh/backend/README.md`](./zh/backend/README.md)

### Website and bot/tooling

- English UI/UX spec: [`docs/en/web-ui-ux-spec-v1.md`](./en/web-ui-ux-spec-v1.md)
- 中文 UI/UX 规格：[`docs/zh/web-ui-ux-spec-v1.md`](./zh/web-ui-ux-spec-v1.md)
- English UI planning docs: [`docs/en/ui/README.md`](./en/ui/README.md)
- 中文 UI 规划文档：[`docs/zh/ui/README.md`](./zh/ui/README.md)
- English agent skill spec: [`docs/en/openclaw-agent-skill.md`](./en/openclaw-agent-skill.md)
- 中文 agent skill 规格：[`docs/zh/openclaw-agent-skill.md`](./zh/openclaw-agent-skill.md)
- English tooling spec: [`docs/en/openclaw-tooling-spec.md`](./en/openclaw-tooling-spec.md)
- 中文 tooling 规格：[`docs/zh/openclaw-tooling-spec.md`](./zh/openclaw-tooling-spec.md)

## Structure rules

- Combined `*-spec-v1.md` files are the fastest way to understand a full domain end-to-end.
- `game/` and `backend/` folders are modular reading paths for focused editing and review.
- Root-level [`docs/game-spec-v1.md`](./game-spec-v1.md) and [`docs/backend-spec-v1.md`](./backend-spec-v1.md) are compatibility entry pages only.
- Prefer adding new source-of-truth docs under `docs/en`, then sync the corresponding Chinese version under `docs/zh`.
