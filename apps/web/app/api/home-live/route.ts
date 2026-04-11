import { NextResponse } from "next/server";

import {
  fallbackChatMessages,
  fallbackLeaderboards,
  fallbackWorldState,
  getHomepageLiveData,
} from "../../../lib/public-api";

export const dynamic = "force-dynamic";

export async function GET() {
  try {
    const data = await getHomepageLiveData();

    return NextResponse.json(
      { data },
      {
        status: 200,
        headers: {
          "Cache-Control": "no-store",
        },
      },
    );
  } catch {
    return NextResponse.json(
      {
        data: {
          worldState: fallbackWorldState,
          chatMessages: fallbackChatMessages,
          leaderboards: fallbackLeaderboards,
        },
      },
      {
        status: 200,
        headers: {
          "Cache-Control": "no-store",
        },
      },
    );
  }
}
