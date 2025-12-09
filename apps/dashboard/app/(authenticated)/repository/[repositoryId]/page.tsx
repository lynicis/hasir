import { redirect } from "next/navigation";

export default async function RepositoryDetailPage({
  params,
}: {
  params: Promise<{ repositoryId: string }>;
}) {
  const { repositoryId } = await params;
  redirect(`/repository/${repositoryId}/documentation`);
}
