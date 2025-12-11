import { Suspense } from "react";
import { Metadata } from "next";

import {
  ResetPasswordForm,
  ResetPasswordFormSkeleton,
} from "@/components/reset-password-form";
import { BrandLink } from "@/components/brand-link";

export const metadata: Metadata = {
  title: "Reset Password | Hasir",
};

interface ResetPasswordPageProps {
  params: Promise<{
    token: string;
  }>;
}

export default async function ResetPasswordPage({
  params,
}: ResetPasswordPageProps) {
  const { token } = await params;

  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <BrandLink />
        <Suspense fallback={<ResetPasswordFormSkeleton />}>
          <ResetPasswordForm token={token} />
        </Suspense>
      </div>
    </div>
  );
}
