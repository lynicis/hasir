import { Metadata } from "next";

import OrganizationRepositoriesContent from "@/components/organization-repositories-content";

export const metadata: Metadata = {
  title: "Repositories - Organization | Hasir",
  description: "Manage organization repositories",
};

export default function RepositoriesPage() {
  return <OrganizationRepositoriesContent />;
}
