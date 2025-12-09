import { Metadata } from "next";

import OrganizationUsersContent from "@/components/organization-users-content";

export const metadata: Metadata = {
  title: "Users - Organization | Hasir",
  description: "Manage organization members",
};

export default function UsersPage() {
  return <OrganizationUsersContent />;
}
