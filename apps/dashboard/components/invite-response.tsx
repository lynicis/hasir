"use client";

import { OrganizationService } from "@buf/hasir_hasir.bufbuild_es/organization/v1/organization_pb";
import { CheckIcon, XIcon, MailIcon, AlertCircleIcon } from "lucide-react";
import { ConnectError, Code } from "@connectrpc/connect";
import { useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { toast } from "sonner";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { useClient } from "@/lib/use-client";
import { cn } from "@/lib/utils";

type InviteStatus =
  | "loading"
  | "valid"
  | "invalid"
  | "accepted"
  | "rejected"
  | "error";

type InviteResponseProps = React.ComponentProps<"div"> & {
  invitationToken: string;
};

export function InviteResponse({
  className,
  invitationToken,
  ...props
}: InviteResponseProps) {
  const router = useRouter();
  const organizationApiClient = useClient(OrganizationService);

  const [status, setStatus] = useState<InviteStatus>("loading");
  const [isAccepting, setIsAccepting] = useState(false);
  const [isRejecting, setIsRejecting] = useState(false);

  useEffect(() => {
    async function validateToken() {
      if (!invitationToken) {
        setStatus("invalid");
        return;
      }

      try {
        await organizationApiClient.isInvitationValid({
          token: invitationToken,
        });
        setStatus("valid");
      } catch (error) {
        if (error instanceof ConnectError) {
          switch (error.code) {
            case Code.NotFound:
              setStatus("invalid");
              break;
            case Code.PermissionDenied:
              setStatus("invalid");
              break;
            default:
              setStatus("error");
          }
        } else {
          setStatus("error");
        }
      }
    }

    validateToken();
  }, [invitationToken, organizationApiClient]);

  async function handleResponse(accept: boolean) {
    const setLoading = accept ? setIsAccepting : setIsRejecting;
    setLoading(true);

    try {
      await organizationApiClient.respondToInvitation({
        invitationId: invitationToken,
        accept,
      });

      setStatus(accept ? "accepted" : "rejected");
      toast.success(
        accept
          ? "You have joined the organization!"
          : "Invitation declined successfully."
      );

      if (accept) {
        setTimeout(() => router.push("/dashboard"), 1500);
      }
    } catch (error) {
      setStatus("error");

      if (error instanceof ConnectError) {
        switch (error.code) {
          case Code.NotFound:
            toast.error("This invitation is no longer valid or has expired.");
            break;
          case Code.AlreadyExists:
            toast.error("You are already a member of this organization.");
            break;
          case Code.PermissionDenied:
            toast.error(
              "You don't have permission to respond to this invitation."
            );
            break;
          default:
            toast.error("Failed to process your response. Please try again.");
        }
      } else {
        toast.error("An unexpected error occurred. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  }

  if (status === "loading") {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4">
              <Skeleton className="size-16 rounded-full" />
            </div>
            <Skeleton className="mx-auto h-6 w-40" />
            <Skeleton className="mx-auto mt-2 h-4 w-60" />
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-3">
              <Skeleton className="h-9 w-full" />
              <Skeleton className="h-9 w-full" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (status === "invalid") {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card className="border-destructive/50">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircleIcon className="size-8 text-destructive" />
            </div>
            <CardTitle className="text-xl">Invalid Invitation</CardTitle>
            <CardDescription>
              This invitation link is invalid, has expired, or has already been
              used.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex justify-center">
            <Button variant="outline" onClick={() => router.push("/login")}>
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (status === "accepted") {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card className="border-green-500/50">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-green-500/10">
              <CheckIcon className="size-8 text-green-500" />
            </div>
            <CardTitle className="text-xl">Welcome to the team!</CardTitle>
            <CardDescription>
              You have successfully joined the organization. Redirecting you to
              the dashboard...
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (status === "rejected") {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card>
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-muted">
              <XIcon className="size-8 text-muted-foreground" />
            </div>
            <CardTitle className="text-xl">Invitation Declined</CardTitle>
            <CardDescription>You have declined the invitation.</CardDescription>
          </CardHeader>
          <CardContent className="flex justify-center">
            <Button variant="outline" onClick={() => router.push("/login")}>
              Go to Login
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (status === "error") {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card className="border-destructive/50">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-destructive/10">
              <AlertCircleIcon className="size-8 text-destructive" />
            </div>
            <CardTitle className="text-xl">Something went wrong</CardTitle>
            <CardDescription>
              We couldn&apos;t process your request. Please try again later.
            </CardDescription>
          </CardHeader>
          <CardContent className="flex justify-center gap-3">
            <Button variant="outline" onClick={() => router.push("/login")}>
              Go to Login
            </Button>
            <Button onClick={() => setStatus("loading")}>Try Again</Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex size-16 items-center justify-center rounded-full bg-primary/10">
            <MailIcon className="size-8 text-primary" />
          </div>
          <CardTitle className="text-xl">You&apos;re Invited!</CardTitle>
          <CardDescription>
            You have been invited to join an organization.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-3">
            <Button
              onClick={() => handleResponse(true)}
              isLoading={isAccepting}
              disabled={isRejecting}
              className="w-full"
            >
              <CheckIcon className="size-4" />
              Accept Invitation
            </Button>
            <Button
              variant="outline"
              onClick={() => handleResponse(false)}
              isLoading={isRejecting}
              disabled={isAccepting}
              className="w-full"
            >
              <XIcon className="size-4" />
              Decline
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
