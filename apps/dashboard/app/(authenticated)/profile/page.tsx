"use client";

import { UserService } from "@buf/hasir_hasir.bufbuild_es/user/v1/user_pb";
import { ConnectError, Code } from "@connectrpc/connect";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";

import { PasswordConfirmationDialog } from "@/components/password-confirmation-dialog";
import { SshApiKeyPanel } from "@/components/ssh-api-key-panel";
import { ProfileForm } from "@/components/profile-form";
import { DangerZone } from "@/components/danger-zone";
import { useSession } from "@/lib/session-provider";
import { Button } from "@/components/ui/button";
import { useClient } from "@/lib/use-client";

type ProfileFormData = {
  email?: string;
  password?: string;
  confirmPassword?: string;
};

export default function ProfilePage() {
  const router = useRouter();
  const userApiClient = useClient(UserService);
  const { session, refreshSession } = useSession();
  const userId = session?.user?.id;
  const [activeTab, setActiveTab] = useState<"profile" | "access">("profile");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [formData, setFormData] = useState<ProfileFormData | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  async function handleProfileSubmit(data: ProfileFormData) {
    setFormData(data);
    setIsDialogOpen(true);
  }

  async function handlePasswordConfirm(data: { currentPassword: string }) {
    if (!formData) {
      toast.error("Form data is missing.");
      return;
    }

    if (!data.currentPassword.trim()) {
      toast.error("Please enter your current password.");
      return;
    }

    if (!userId) {
      toast.error("Unable to identify user. Please log in again.");
      return;
    }

    const hasEmailUpdate = formData.email && formData.email.trim() !== "";
    const hasPasswordUpdate =
      formData.password && formData.password.trim() !== "";

    if (!hasEmailUpdate && !hasPasswordUpdate) {
      toast.error("Please update at least one field.");
      return;
    }

    try {
      const updatePayload: {
        password: string;
        userId: string;
        email?: string;
        newPassword?: string;
      } = {
        password: data.currentPassword,
        userId,
      };

      if (hasEmailUpdate) {
        updatePayload.email = formData.email;
      }

      if (hasPasswordUpdate) {
        updatePayload.newPassword = formData.password;
      }

      await userApiClient.updateUser(updatePayload);

      await refreshSession();
      setIsDialogOpen(false);
      toast.success("Profile updated successfully!");
      setFormData(null);
    } catch (error) {
      if (error instanceof ConnectError) {
        if (error.code === Code.Unauthenticated) {
          toast.error("Invalid current password. Please try again.");
          return;
        }

        if (error.code === Code.AlreadyExists) {
          toast.error("An account with this email already exists.");
          return;
        }

        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to update this profile.");
          return;
        }
      }

      toast.error("Failed to update profile. Please try again.");
    }
  }

  function handleDialogCancel() {
    setFormData(null);
    setIsDialogOpen(false);
  }

  async function handleDeleteAccount() {
    if (!userId) {
      toast.error("Unable to identify user. Please log in again.");
      return;
    }

    setIsDeleting(true);
    try {
      await userApiClient.deleteAccount({});

      toast.success("Account deleted successfully.");
      setTimeout(() => {
        router.push("/login");
      }, 600);
    } catch (error) {
      setIsDeleting(false);
      if (error instanceof ConnectError) {
        if (error.code === Code.Unauthenticated) {
          toast.error("You are not authorized to delete this account.");
          return;
        }
        if (error.code === Code.PermissionDenied) {
          toast.error("You don't have permission to delete this account.");
          return;
        }
      }

      toast.error("Failed to delete account. Please try again.");
    }
  }

  return (
    <div className="container mx-auto max-w-2xl min-h-screen flex flex-col justify-center gap-6 py-8 px-4">
      <div>
        <div className="flex items-center gap-2 border-b border-border pb-2 mb-4">
          <Button
            type="button"
            onClick={() => setActiveTab("profile")}
            className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              activeTab === "profile"
                ? "bg-background text-foreground border border-border border-b-transparent -mb-px"
                : "text-muted-foreground hover:text-foreground"
            }`}
            aria-selected={activeTab === "profile"}
          >
            Profile
          </Button>
          <Button
            type="button"
            onClick={() => setActiveTab("access")}
            className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              activeTab === "access"
                ? "bg-background text-foreground border border-border border-b-transparent -mb-px"
                : "text-muted-foreground hover:text-foreground"
            }`}
            aria-selected={activeTab === "access"}
          >
            SSH / API key
          </Button>
        </div>

        {activeTab === "profile" ? (
          <ProfileForm onSubmit={handleProfileSubmit} />
        ) : (
          <SshApiKeyPanel />
        )}
      </div>

      <PasswordConfirmationDialog
        open={isDialogOpen}
        onOpenChange={setIsDialogOpen}
        onConfirm={handlePasswordConfirm}
        onCancel={handleDialogCancel}
      />
      <DangerZone onDelete={handleDeleteAccount} isDeleting={isDeleting} />
    </div>
  );
}
