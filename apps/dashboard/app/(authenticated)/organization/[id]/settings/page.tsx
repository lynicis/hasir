"use client";

import { OrganizationSettingsForm } from "@/components/organization-settings-form";

export default function OrganizationSettingsPage() {
  return (
    <div className="h-full space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Organization Settings</h1>
        <p className="text-muted-foreground mt-1 text-sm">
          Manage your organization details and preferences
        </p>
      </div>
      <OrganizationSettingsForm />
    </div>
  );
}
