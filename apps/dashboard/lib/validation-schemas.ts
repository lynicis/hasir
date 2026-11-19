import { z } from "zod/v4";

export const passwordSchema = z
  .string()
  .min(8, { error: "Password must be at least 8 characters" })
  .max(20, { error: "Password must be at most 20 characters" });

export const confirmPasswordSchema = z
  .string()
  .min(8, { error: "Confirm password must be at least 8 characters" })
  .max(20, { error: "Confirm password must be at most 20 characters" });

export const organizationNameSchema = z
  .string()
  .trim()
  .min(1, { error: "Please enter an organization name." })
  .max(100, { error: "Organization name must be at most 100 characters." })
  .regex(/^[a-zA-Z0-9 _.-]+$/, {
    error: "Only letters, numbers, spaces, and basic punctuation allowed",
  });

export const repositoryNameSchema = z
  .string()
  .min(1, "Repository name is required")
  .max(100, "Repository name must be less than 100 characters")
  .regex(
    /^[a-zA-Z0-9-_]+$/,
    "Repository name can only contain letters, numbers, hyphens and underscores"
  );

export const keyNameSchema = z
  .string()
  .min(1, "Key name is required")
  .max(50, "Key name must be 50 characters or less");

export const sshPublicKeySchema = z
  .string()
  .min(1, "Public key is required")
  .refine(
    (key) => key.startsWith("ssh-") || key.startsWith("ecdsa-"),
    "Must be a valid SSH public key (should start with ssh- or ecdsa-)"
  );
