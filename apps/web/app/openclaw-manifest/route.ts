import { openClawManifest } from "../../lib/openclaw-guide";

export const revalidate = 30;

export async function GET() {
  return Response.json(openClawManifest, {
    headers: {
      "Cache-Control": "public, max-age=30, stale-while-revalidate=300",
    },
  });
}
