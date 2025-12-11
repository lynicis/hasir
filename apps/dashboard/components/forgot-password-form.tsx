"use client";

import { UserService } from "@buf/hasir_hasir.bufbuild_es/user/v1/user_pb";
import { zodResolver } from "@hookform/resolvers/zod";
import { Controller, useForm } from "react-hook-form";
import { useEffect, useState } from "react";
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

const schema = z.object({
  email: z.email({ error: "Please enter a valid email address." }),
});

type ISchema = z.infer<typeof schema>;

export function ForgotPasswordForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const [isSubmitted, setIsSubmitted] = useState(false);
  const userApiClient = useClient(UserService);
  const {
    control,
    handleSubmit,
    formState: { isLoading, errors },
    setError,
  } = useForm<ISchema>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
    },
  });

  async function onSubmit({ email }: ISchema) {
    try {
      await userApiClient.forgotPassword({
        email,
      });

      setIsSubmitted(true);
      toast.success("Password reset email sent!");
    } catch (error) {
      if (error instanceof Error) {
        setError("root", {
          message:
            error.message || "Error occurred while sending reset email.",
        });
      } else {
        setError("root", {
          message: "Error occurred while sending reset email.",
        });
      }
    }
  }

  useEffect(() => {
    if (errors.root) {
      toast.error(errors.root.message);
    }
  }, [errors.root]);

  if (isSubmitted) {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card>
          <CardHeader className="text-center">
            <CardTitle className="text-xl">Check your email</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-muted-foreground space-y-4 text-center text-sm">
              <p>
                We&apos;ve sent you an email with instructions to reset your
                password.
              </p>
              <p>
                If you don&apos;t see the email, check your spam folder or{" "}
                <button
                  onClick={() => setIsSubmitted(false)}
                  className="text-primary hover:underline"
                >
                  try again
                </button>
                .
              </p>
            </div>
            <div className="mt-6">
              <FieldDescription className="text-center">
                Remember your password? <Link href="/login">Login</Link>
              </FieldDescription>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Forgot your password?</CardTitle>
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
                      required
                      disabled={isLoading}
                      aria-invalid={fieldState.invalid}
                    />
                    <FieldDescription>
                      Enter your email address and we&apos;ll send you a link to
                      reset your password.
                    </FieldDescription>
                  </Field>
                )}
              />
              <Field>
                <Button type="submit" isLoading={isLoading}>
                  Send Reset Link
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

export function ForgotPasswordFormSkeleton({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Forgot your password?</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6">
            <div className="grid gap-2">
              <Skeleton className="h-4 w-10" />
              <Skeleton className="h-10 w-full" />
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
