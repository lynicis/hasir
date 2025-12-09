import { Metadata } from "next";

import OrganizationSettingsContent from "@/components/organization-settings-content";

export const metadata: Metadata = {
  title: "Settings - Organization | Hasir",
  description: "Manage organization settings",
};

export default function OrganizationSettingsPage() {
  return <OrganizationSettingsContent />;
}
