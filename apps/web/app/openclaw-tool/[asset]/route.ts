import { readFile } from "node:fs/promises";
import path from "node:path";

export const revalidate = 30;

const assetMap = {
  manifest: {
    kind: "json",
  },
  clawgame: {
    kind: "file",
    source: "clawgame",
    filename: "clawgame",
    contentType: "text/x-shellscript; charset=utf-8",
  },
  "clawgame-tool-py": {
    kind: "file",
    source: "clawgame_tool.py",
    filename: "clawgame_tool.py",
    contentType: "text/x-python; charset=utf-8",
  },
} as const;

function getRepoToolPath(name: string) {
  return path.resolve(process.cwd(), "..", "..", "tools", name);
}

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ asset: string }> },
) {
  const { asset } = await params;
  const config = assetMap[asset as keyof typeof assetMap];

  if (!config) {
    return new Response("not found", { status: 404 });
  }

  if (config.kind === "json") {
    const baseUrl = (process.env.NEXT_PUBLIC_CONSOLE_BASE_URL ?? "http://localhost:4000").replace(/\/$/, "");
    return Response.json(
      {
        kind: "clawgame-openclaw-tool-download",
        version: "2026-03-30.4",
        shell: {
          path: "tools/clawgame",
          url: `${baseUrl}/openclaw-tool/clawgame`,
        },
        python: {
          path: "tools/clawgame_tool.py",
          url: `${baseUrl}/openclaw-tool/clawgame-tool-py`,
        },
        install_steps: [
          "mkdir -p tools",
          `curl -fsSL ${baseUrl}/openclaw-tool/clawgame -o tools/clawgame`,
          `curl -fsSL ${baseUrl}/openclaw-tool/clawgame-tool-py -o tools/clawgame_tool.py`,
          "chmod +x tools/clawgame",
        ],
      },
      {
        headers: {
          "Cache-Control": "public, max-age=30, stale-while-revalidate=300",
        },
      },
    );
  }

  const fileText = await readFile(getRepoToolPath(config.source), "utf8");
  return new Response(fileText, {
    headers: {
      "Content-Type": config.contentType,
      "Content-Disposition": `attachment; filename="${config.filename}"`,
      "Cache-Control": "public, max-age=30, stale-while-revalidate=300",
    },
  });
}
