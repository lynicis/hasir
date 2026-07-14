import type { Metadata } from "next";

import { redirect } from "next/navigation";

export const metadata: Metadata = {
  title: "Hasir | Organization Settings",
};

export default async function OrganizationPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  redirect(`/organization/${id}/users`);
}
