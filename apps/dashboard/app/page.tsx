import { Metadata } from "next";

import HomePageContent from "@/components/home-page-content";

export const metadata: Metadata = {
  title: "Hasir - Schema Registry",
  description: "Modern Proto Schema Registry dashboard",
};

export default function HomePage() {
  return <HomePageContent />;
}
