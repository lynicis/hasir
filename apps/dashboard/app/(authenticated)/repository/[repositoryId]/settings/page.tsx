import { Metadata } from "next";

import RepositorySettingsContent from "@/components/repository-settings-content";

export const metadata: Metadata = {
  title: "Settings - Repository | Hasir",
  description: "Repository settings",
};

export default function SettingsPage() {
  return <RepositorySettingsContent />;
}
