import { Metadata } from "next";

import RepositoryFilesContent from "@/components/repository-files-content";

export const metadata: Metadata = {
  title: "Files - Repository | Hasir",
  description: "Browse repository files",
};

export default function FilesPage() {
  return <RepositoryFilesContent />;
}
