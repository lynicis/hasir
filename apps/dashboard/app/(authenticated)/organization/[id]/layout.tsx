"use client";

import { getMembers } from "@buf/hasir_hasir.connectrpc_query-es/organization/v1/organization-OrganizationService_connectquery";
import { Role } from "@buf/hasir_hasir.bufbuild_es/shared/role_pb";
import { Building2, Settings, Users } from "lucide-react";
import { useQuery } from "@connectrpc/connect-query";
import { usePathname } from "next/navigation";
import { use, useMemo } from "react";
import Link from "next/link";

import type { Permission } from "@/components/member-item";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useSession } from "@/lib/session-provider";
import { customRetry } from "@/lib/query-retry";

const tabs = [
  { id: "users", label: "Users", icon: Users, path: "users" as const },
  {
    id: "settings",
    label: "Organization",
    icon: Building2,
    path: "settings" as const,
  },
  {
    id: "repositories",
    label: "Repositories",
    icon: Settings,
    path: "repositories" as const,
  },
] as const;

const memberRoleMapper = new Map<Role, Permission>([
  [Role.OWNER, "owner"],
  [Role.AUTHOR, "author"],
  [Role.READER, "reader"],
]);

export default function OrganizationLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: Promise<{ id: string }>;
}) {
  const pathname = usePathname();
  const { id: organizationId } = use(params);
  const { session } = useSession();

  const { data: membersData } = useQuery(
    getMembers,
    { id: organizationId },
    { retry: customRetry }
  );

  const currentMemberPermission = useMemo<Permission | null>(() => {
    if (!membersData?.members || !session?.user?.email) {
      return null;
    }

    const currentMember = membersData.members.find(
      (member) => member.email === session.user?.email
    );

    if (!currentMember) {
      return null;
    }

    return memberRoleMapper.get(currentMember.role) as Permission;
  }, [membersData, session]);

  const canSeeOrganizationTab = currentMemberPermission === "owner";

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
                if (tab.id === "settings" && !canSeeOrganizationTab) {
                  return null;
                }
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
