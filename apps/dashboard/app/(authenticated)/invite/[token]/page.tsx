import { InviteResponse } from "@/components/invite-response";

type InvitePageProps = {
  params: Promise<{ token: string }>;
};

export default async function InvitePage({ params }: InvitePageProps) {
  const { token } = await params;

  return (
    <main className="flex flex-1 flex-col items-center justify-center p-6 md:p-10">
      <div className="flex w-full max-w-sm flex-col gap-6">
        <InviteResponse invitationToken={token} />
      </div>
    </main>
  );
}
