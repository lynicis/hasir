"use client";

import * as React from "react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const MOCK_ORGANIZATIONS = [
  { id: "org-1", name: "Acme Corp" },
  { id: "org-2", name: "Hasir Labs" },
  { id: "org-3", name: "Proto Systems" },
];

const MOCK_REPOSITORIES = [
  { id: "repo-1", name: "payment-protos", organization: "Acme Corp" },
  { id: "repo-2", name: "analytics-protos", organization: "Hasir Labs" },
  { id: "repo-3", name: "core-registry", organization: "Proto Systems" },
  { id: "repo-4", name: "internal-tools", organization: "Hasir Labs" },
];

export function Dashboard() {
  const [activeOrgId, setActiveOrgId] = React.useState<string | "all">("all");

  const filteredRepositories =
    activeOrgId === "all"
      ? MOCK_REPOSITORIES
      : MOCK_REPOSITORIES.filter((repo) => {
          const org = MOCK_ORGANIZATIONS.find(
            (organization) => organization.id === activeOrgId
          );
          return org ? repo.organization === org.name : true;
        });

  return (
    <div className="h-[calc(100vh-4.5rem)] bg-background px-6 py-6 overflow-hidden">
      <div className="mx-auto flex h-full w-full max-w-6xl flex-col gap-6">
        <main className="grid flex-1 grid-cols-[260px_minmax(0,1fr)] gap-6 pt-2">
          <Card className="h-full gap-0 overflow-hidden rounded-2xl border border-border/60 py-0 shadow-sm">
            <CardHeader className="flex items-center bg-primary px-6 py-4">
              <CardTitle className="text-sm font-medium text-white">
                Your organizations
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 py-4">
              <button
                type="button"
                onClick={() => setActiveOrgId("all")}
                className={`flex w-full items-center justify-between rounded-md px-3 py-2.5 text-sm transition-colors ${
                  activeOrgId === "all"
                    ? "bg-accent text-accent-foreground"
                    : "hover:bg-accent hover:text-accent-foreground"
                }`}
              >
                <span>All organizations</span>
              </button>
              {MOCK_ORGANIZATIONS.map((org) => {
                const isActive = activeOrgId === org.id;
                return (
                  <button
                    key={org.id}
                    type="button"
                    onClick={() => setActiveOrgId(org.id)}
                    className={`flex w-full items-center justify-between rounded-md px-3 py-2.5 text-sm transition-colors ${
                      isActive
                        ? "bg-accent text-accent-foreground"
                        : "hover:bg-accent hover:text-accent-foreground"
                    }`}
                  >
                    <span>{org.name}</span>
                  </button>
                );
              })}
            </CardContent>
          </Card>
          <Card className="h-full gap-0 overflow-hidden rounded-2xl border border-border/60 py-0 shadow-sm">
            <CardHeader className="flex items-center justify-between bg-primary px-6 py-4">
              <div className="space-y-0.5">
                <CardTitle className="text-sm font-medium text-white">
                  Repositories
                </CardTitle>
                {activeOrgId !== "all" && (
                  <p className="text-xs text-white/70">
                    Showing repositories in{" "}
                    {
                      MOCK_ORGANIZATIONS.find(
                        (organization) => organization.id === activeOrgId
                      )?.name
                    }
                  </p>
                )}
              </div>
              <span className="text-xs text-white/70">
                {filteredRepositories.length} repos
              </span>
            </CardHeader>
            <CardContent className="space-y-2.5 py-4">
              {filteredRepositories.map((repo) => (
                <div
                  key={repo.id}
                  className="hover:bg-accent/60 flex items-center justify-between rounded-xl border border-border/60 bg-card px-4 py-3 text-sm transition-colors"
                >
                  <div className="flex flex-col">
                    <span className="font-medium">{repo.name}</span>
                    <span className="text-muted-foreground text-xs">
                      {repo.organization}
                    </span>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </main>
      </div>
    </div>
  );
}
