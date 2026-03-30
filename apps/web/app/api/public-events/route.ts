import { NextRequest, NextResponse } from "next/server";

export const dynamic = "force-dynamic";

export async function GET(request: NextRequest) {
  const upstream = new URL(`${apiBaseUrl()}/api/v1/public/events`);
  const limit = request.nextUrl.searchParams.get("limit");
  const cursor = request.nextUrl.searchParams.get("cursor");

  if (limit) upstream.searchParams.set("limit", limit);
  if (cursor) upstream.searchParams.set("cursor", cursor);

  try {
    const response = await fetch(upstream.toString(), {
      cache: "no-store",
    });

    if (!response.ok) {
      return NextResponse.json(
        {
          data: {
            items: [],
            next_cursor: null,
          },
        },
        { status: response.status },
      );
    }

    const payload = await response.json();
    return NextResponse.json(payload, {
      status: 200,
      headers: {
        "Cache-Control": "no-store",
      },
    });
  } catch {
    return NextResponse.json(
      {
        data: {
          items: [],
          next_cursor: null,
        },
      },
      { status: 200 },
    );
  }
}

function apiBaseUrl() {
  const baseUrl =
    process.env.API_BASE_URL ??
    process.env.NEXT_PUBLIC_API_BASE_URL ??
    "http://127.0.0.1:8080";

  return baseUrl.endsWith("/") ? baseUrl.slice(0, -1) : baseUrl;
}
