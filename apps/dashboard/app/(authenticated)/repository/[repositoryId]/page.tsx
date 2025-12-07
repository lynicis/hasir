"use client";

import { useParams, useRouter } from "next/navigation";
import { useEffect } from "react";

export default function RepositoryDetailPage() {
  const params = useParams();
  const router = useRouter();
  const repositoryId = params.repositoryId as string;

  useEffect(() => {
    router.replace(`/repository/${repositoryId}/documentation`);
  }, [repositoryId, router]);

  return null;
}
