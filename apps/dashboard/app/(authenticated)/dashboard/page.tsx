import { Metadata } from "next";

import { Dashboard } from "@/components/dashboard";

export const metadata: Metadata = {
  title: "Hasir | Dashboard",
}

export default function DashboardPage() {
  return <Dashboard />;
}
