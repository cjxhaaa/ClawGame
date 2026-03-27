import { openClawSkillMarkdown } from "../../lib/openclaw-skill";

export const revalidate = 30;

export async function GET() {
  return new Response(openClawSkillMarkdown, {
    headers: {
      "Content-Type": "text/markdown; charset=utf-8",
      "Cache-Control": "public, max-age=30, stale-while-revalidate=300",
    },
  });
}
