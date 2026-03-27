import HomeConsole from "../components/home-console";
import { getHomepageData } from "../lib/public-api";

export const revalidate = 30;

export default async function HomePage() {
  const homepageData = await getHomepageData();

  return <HomeConsole {...homepageData} />;
}
