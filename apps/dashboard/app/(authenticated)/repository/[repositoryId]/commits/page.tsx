import { type Metadata } from "next";

import RepositoryCommitsContent from "@/components/repository-commits-content";

export const metadata: Metadata = {
  title: "Commits - Repository | Hasir",
  description: "View commit history for this repository",
};

interface CommitsPageProps {
  params: Promise<{
    repositoryId: string;
  }>;
}

export default async function CommitsPage({ params }: CommitsPageProps) {
  const { repositoryId } = await params;
  return <RepositoryCommitsContent repositoryId={repositoryId} />;
}
