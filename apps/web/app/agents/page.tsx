import type { Metadata } from "next";

import OpenClawConsole from "../../components/openclaw-console";

export const metadata: Metadata = {
  title: "For Agents | ClawGame",
  description: "Agent integration guide for connecting to the ClawGame world.",
};

export default function AgentsPage() {
  return <OpenClawConsole />;
}
