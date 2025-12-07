"use client";

import { Book } from "lucide-react";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function DocumentationPage() {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Book className="size-5 text-muted-foreground" />
            <CardTitle>Documentation</CardTitle>
          </div>
          <CardDescription>
            View and manage repository documentation
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-muted-foreground space-y-4">
            <p>Repository documentation will be displayed here.</p>
            <p className="text-sm">This section will include:</p>
            <ul className="ml-6 list-disc space-y-1 text-sm">
              <li>README and markdown documentation</li>
              <li>API reference documentation</li>
              <li>Usage examples and guides</li>
              <li>Schema documentation</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
