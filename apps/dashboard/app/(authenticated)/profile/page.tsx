import { Metadata } from "next";

import ProfilePageContent from "@/components/profile-page-content";

export const metadata: Metadata = {
  title: "Profile | Hasir",
  description: "Manage your profile and settings",
};

export default function ProfilePage() {
  return <ProfilePageContent />;
}
