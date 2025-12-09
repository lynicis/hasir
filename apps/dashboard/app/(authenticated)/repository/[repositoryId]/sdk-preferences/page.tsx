import { Metadata } from "next";

import RepositorySdkPreferencesContent from "@/components/repository-sdk-preferences-content";

export const metadata: Metadata = {
  title: "SDK Preferences - Repository | Hasir",
  description: "Configure SDK generation preferences",
};

export default function SdkPreferencesPage() {
  return <RepositorySdkPreferencesContent />;
}
