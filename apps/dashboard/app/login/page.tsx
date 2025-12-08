import { Suspense } from "react";
import { Metadata } from "next";

import { LoginForm, LoginFormSkeleton } from "@/components/login-form";
import { BrandLink } from "@/components/brand-link";

export const metadata: Metadata = {
  title: "Hasir | Login"
};

export default function LoginPage() {
  return (
    <div className="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <BrandLink />
        <Suspense fallback={<LoginFormSkeleton />}>
          <LoginForm />
        </Suspense>
      </div>
    </div>
  );
}
