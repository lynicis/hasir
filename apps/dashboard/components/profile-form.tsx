"use client";

import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { z } from "zod/v4";

import {
  FieldDescription,
  FieldLabel,
  FieldGroup,
  Field,
  FieldError,
} from "@/components/ui/field";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { useSession } from "@/lib/session-provider";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const schema = z
  .object({
    email: z
      .email({ error: "Please enter a valid email address." })
      .optional()
      .or(z.literal("")),
    password: z
      .string()
      .optional()
      .or(z.literal(""))
      .refine(
        (val) =>
          !val || val.trim() === "" || (val.length >= 8 && val.length <= 20),
        {
          error: "Password must be between 8 and 20 characters.",
        }
      ),
    confirmPassword: z
      .string()
      .optional()
      .or(z.literal(""))
      .refine(
        (val) =>
          !val || val.trim() === "" || (val.length >= 8 && val.length <= 20),
        {
          error: "Confirm password must be between 8 and 20 characters.",
        }
      ),
  })
  .refine(
    ({ password, confirmPassword }) => {
      if (!password && !confirmPassword) return true;
      if (!password || !confirmPassword) return false;
      return password === confirmPassword;
    },
    {
      error: "Passwords do not match",
      path: ["confirmPassword"],
    }
  );

type ISchema = z.infer<typeof schema>;

interface ProfileFormProps {
  onSubmit: (data: ISchema) => void;
  resetTrigger?: number;
}

export function ProfileForm({ onSubmit, resetTrigger }: ProfileFormProps) {
  const { session } = useSession();
  const {
    control,
    handleSubmit,
    formState: { isLoading, isSubmitting },
    reset,
  } = useForm<ISchema>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: session?.user?.email || "",
      password: "",
      confirmPassword: "",
    },
  });

  useEffect(() => {
    reset({
      email: session?.user?.email || "",
      password: "",
      confirmPassword: "",
    });
  }, [session?.user?.email, reset, resetTrigger]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-2xl">Profile Settings</CardTitle>
        <CardDescription>
          Update your personal information and account details.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)}>
          <FieldGroup>
            <Controller
              control={control}
              name="email"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>Email</FieldLabel>
                  <Input
                    {...field}
                    id={field.name}
                    type="email"
                    placeholder="m@example.com"
                    disabled={isLoading || isSubmitting}
                    aria-invalid={fieldState.invalid}
                  />
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
            <Controller
              control={control}
              name="password"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>New Password</FieldLabel>
                  <Input
                    {...field}
                    id={field.name}
                    type="password"
                    placeholder="Enter new password"
                    disabled={isLoading || isSubmitting}
                    aria-invalid={fieldState.invalid}
                  />
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                  <FieldDescription>
                    Must be at least 8 characters long.
                  </FieldDescription>
                </Field>
              )}
            />
            <Controller
              control={control}
              name="confirmPassword"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>
                    Confirm New Password
                  </FieldLabel>
                  <Input
                    {...field}
                    id={field.name}
                    type="password"
                    placeholder="Confirm new password"
                    disabled={isLoading || isSubmitting}
                    aria-invalid={fieldState.invalid}
                  />
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
            <Field>
              <Button
                type="submit"
                isLoading={isSubmitting}
                disabled={isLoading}
              >
                Save Changes
              </Button>
            </Field>
          </FieldGroup>
        </form>
      </CardContent>
    </Card>
  );
}
