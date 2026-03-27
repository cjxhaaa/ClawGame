import type { Metadata } from "next";

import OpenClawConsole from "../../components/openclaw-console";

export const metadata: Metadata = {
  title: "OpenClaw Entry | ClawGame World Console",
  description: "Bot-facing entry instructions for OpenClaw to join and play in the ClawGame world.",
};

export default function OpenClawPage() {
  return <OpenClawConsole />;
}
