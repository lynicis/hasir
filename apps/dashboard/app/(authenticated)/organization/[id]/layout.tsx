"use client";

import { Building2, Settings, Users } from "lucide-react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { use } from "react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const tabs = [
  { id: "users", label: "Users", icon: Users, path: "users" },
  { id: "settings", label: "Organization", icon: Building2, path: "settings" },
  {
    id: "repositories",
    label: "Repositories",
    icon: Settings,
    path: "repositories",
  },
];

export default function OrganizationLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ id: string }>;
}) {
  const pathname = usePathname();
  const { id: organizationId } = use(params);

  return (
    <div className="h-[calc(100vh-4.5rem)] bg-background px-6 py-6 overflow-hidden">
      <div className="mx-auto flex h-full w-full max-w-6xl gap-6">
        <Card className="h-full w-64 gap-0 overflow-hidden rounded-2xl border border-border/60 py-0 shadow-sm">
          <CardHeader className="flex items-center bg-primary px-6 py-4">
            <CardTitle className="text-sm font-medium text-secondary">
              Settings
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 py-4">
            <nav className="space-y-1">
              {tabs.map((tab) => {
                const Icon = tab.icon;
                const href = `/organization/${organizationId}/${tab.path}`;
                const isActive =
                  pathname === href ||
                  (tab.path === "users" &&
                    pathname === `/organization/${organizationId}`);
                return (
                  <Link
                    key={tab.id}
                    href={href}
                    className={`flex w-full items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors ${
                      isActive
                        ? "bg-accent text-accent-foreground"
                        : "hover:bg-accent hover:text-accent-foreground"
                    }`}
                  >
                    <Icon className="size-4" />
                    <span>{tab.label}</span>
                  </Link>
                );
              })}
            </nav>
          </CardContent>
        </Card>
        <main className="flex-1 overflow-y-auto">{children}</main>
      </div>
    </div>
  );
}
