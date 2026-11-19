import {
  confirmPasswordSchema,
  keyNameSchema,
  organizationNameSchema,
  passwordSchema,
  repositoryNameSchema,
  sshPublicKeySchema,
} from "./validation-schemas";

describe("validation-schemas", () => {
  it("validates password and confirmPassword lengths", () => {
    expect(passwordSchema.safeParse("short").success).toBe(false);
    expect(passwordSchema.safeParse("validpass123").success).toBe(true);
    expect(confirmPasswordSchema.safeParse("a".repeat(21)).success).toBe(false);
  });

  it("validates organizationNameSchema", () => {
    expect(organizationNameSchema.safeParse("   ").success).toBe(false);
    expect(organizationNameSchema.safeParse("Acme Corp 1.0").success).toBe(true);
    expect(organizationNameSchema.safeParse("Bad/Name").success).toBe(false);
  });

  it("validates repositoryNameSchema", () => {
    expect(repositoryNameSchema.safeParse("my-repo_1").success).toBe(true);
    expect(repositoryNameSchema.safeParse("my repo").success).toBe(false);
  });

  it("validates keyNameSchema and sshPublicKeySchema", () => {
    expect(keyNameSchema.safeParse("Work Laptop").success).toBe(true);
    expect(keyNameSchema.safeParse("").success).toBe(false);
    expect(sshPublicKeySchema.safeParse("ssh-ed25519 AAAA").success).toBe(true);
    expect(sshPublicKeySchema.safeParse("ecdsa-sha2-nistp256 AAAA").success).toBe(true);
    expect(sshPublicKeySchema.safeParse("invalid-key").success).toBe(false);
  });
});
