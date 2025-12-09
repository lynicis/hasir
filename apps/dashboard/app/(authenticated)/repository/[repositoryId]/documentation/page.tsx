import { Metadata } from "next";

import RepositoryDocumentationContent from "@/components/repository-documentation-content";

export const metadata: Metadata = {
  title: "Documentation - Repository | Hasir",
  description: "Repository documentation",
};

export default function DocumentationPage() {
  return <RepositoryDocumentationContent />;
}
