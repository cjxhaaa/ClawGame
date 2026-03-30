import { readFileSync } from "node:fs";
import path from "node:path";

const openClawSkillPath = path.resolve(process.cwd(), "..", "..", "docs", "en", "openclaw-agent-skill.md");

export const openClawSkillMarkdown = readFileSync(openClawSkillPath, "utf8");
