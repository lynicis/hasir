"use client";

import { UserService } from "@buf/hasir_hasir.bufbuild_es/user/v1/user_pb";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { toast } from "sonner";
import Link from "next/link";
import { z } from "zod/v4";

import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useClient } from "@/lib/use-client";
import { cn } from "@/lib/utils";

const schema = z
  .object({
    password: z
      .string()
      .min(8, { error: "Password must be at least 8 characters" })
      .max(20, { error: "Password must be at most 20 characters" }),
    confirmPassword: z
      .string()
      .min(8, { error: "Confirm password must be at least 8 characters" })
      .max(20, { error: "Confirm password must be at most 20 characters" }),
  })
  .refine(({ password, confirmPassword }) => password === confirmPassword, {
    error: "Passwords do not match",
    path: ["confirmPassword"],
  });

type ISchema = z.infer<typeof schema>;

interface ResetPasswordFormProps extends React.ComponentProps<"div"> {
  token: string;
}

export function ResetPasswordForm({
  token,
  className,
  ...props
}: ResetPasswordFormProps) {
  const router = useRouter();
  const userApiClient = useClient(UserService);
  const {
    control,
    handleSubmit,
    formState: { isLoading, errors },
    setError,
  } = useForm<ISchema>({
    resolver: zodResolver(schema),
    defaultValues: {
      password: "",
      confirmPassword: "",
    },
  });

  async function onSubmit({ password }: ISchema) {
    try {
      await userApiClient.resetPassword({
        token,
        newPassword: password,
      });

      toast.success("Password reset successful! You can now log in.");
      setTimeout(() => {
        router.push("/login");
      }, 600);
    } catch (error) {
      if (error instanceof Error) {
        setError("root", {
          message: error.message || "Error occurred while resetting password.",
        });
      } else {
        setError("root", {
          message: "Error occurred while resetting password.",
        });
      }
    }
  }

  useEffect(() => {
    if (errors.root) {
      toast.error(errors.root.message);
    }
  }, [errors.root]);

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Reset your password</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)}>
            <FieldGroup>
              <Field>
                <Field className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <Controller
                    name="password"
                    control={control}
                    render={({ field, fieldState }) => (
                      <Field data-invalid={fieldState.invalid}>
                        <FieldLabel htmlFor={field.name}>
                          New Password
                        </FieldLabel>
                        <Input
                          {...field}
                          id={field.name}
                          type="password"
                          required
                          placeholder="********"
                          aria-invalid={fieldState.invalid}
                          disabled={isLoading}
                        />
                      </Field>
                    )}
                  />
                  <Controller
                    name="confirmPassword"
                    control={control}
                    render={({ field, fieldState }) => (
                      <Field data-invalid={fieldState.invalid}>
                        <FieldLabel htmlFor={field.name}>
                          Confirm Password
                        </FieldLabel>
                        <Input
                          {...field}
                          id={field.name}
                          type="password"
                          required
                          placeholder="********"
                          aria-invalid={fieldState.invalid}
                          disabled={isLoading}
                        />
                      </Field>
                    )}
                  />
                </Field>
                <FieldDescription>
                  Must be at least 8 characters long.
                </FieldDescription>
              </Field>
              <Field>
                <Button type="submit" isLoading={isLoading}>
                  Reset Password
                </Button>
                <FieldDescription className="text-center">
                  Remember your password? <Link href="/login">Login</Link>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

export function ResetPasswordFormSkeleton({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Reset your password</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6">
            <div className="grid gap-2">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div className="grid gap-2">
                  <Skeleton className="h-4 w-24" />
                  <Skeleton className="h-10 w-full" />
                </div>
                <div className="grid gap-2">
                  <Skeleton className="h-4 w-28" />
                  <Skeleton className="h-10 w-full" />
                </div>
              </div>
              <Skeleton className="h-4 w-full" />
            </div>
            <div className="grid gap-2">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-4 w-48 mx-auto" />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
