import { Metadata } from "next";

import { Dashboard } from "@/components/dashboard";

export const metadata: Metadata = {
  title: "Dashboard | Hasir",
}

export default function DashboardPage() {
  return <Dashboard />;
}
