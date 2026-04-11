import HomeConsole from "../components/home-console";
import { getHomepageStaticData } from "../lib/public-api";

export const revalidate = 30;

export default async function HomePage() {
  const homepageData = await getHomepageStaticData();

  return <HomeConsole {...homepageData} />;
}
