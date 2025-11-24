"use client";

import { UserService } from "@buf/hasir_hasir.bufbuild_es/user/v1/user_pb";
import { ConnectError, Code } from "@connectrpc/connect";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import Link from "next/link";
import { z } from "zod/v4";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useClient } from "@/lib/use-client";
import { cn } from "@/lib/utils";

const schema = z
  .object({
    email: z.email({ message: "Invalid email address" }),
    password: z
      .string()
      .min(8, { message: "Password must be at least 8 characters" })
      .max(20, { message: "Password must be at most 20 characters" }),
    confirmPassword: z
      .string()
      .min(8, { message: "Confirm password must be at least 8 characters" })
      .max(20, { message: "Confirm password must be at most 20 characters" }),
  })
  .refine(({ password, confirmPassword }) => password === confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type ISchema = z.infer<typeof schema>;

export function RegisterForm({
  className,
  ...props
}: React.ComponentProps<"div">) {
  const router = useRouter();
  const userApiClient = useClient(UserService);
  const {
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<ISchema>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
      password: "",
      confirmPassword: "",
    },
  });

  async function onSubmit({ email, password }: ISchema) {
    try {
      await userApiClient.register({
        email,
        password,
      });

      toast.success("Account successfully created. You may now log in.");
      setTimeout(() => router.push("/login"), 1000);
    } catch (error) {
      if (error instanceof ConnectError && error.code === Code.AlreadyExists) {
        setError("root", {
          message: "An account with this email already exists.",
        });
      }

      setError("root", {
        message: "Error occurred while creating your account.",
      });
    }
  }

  if (errors.root) {
    return toast.error(errors.root.message);
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader className="text-center">
          <CardTitle className="text-xl">Create your account</CardTitle>
          <CardDescription>
            Enter your email below to create your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)}>
            <FieldGroup>
              <Controller
                name="email"
                control={control}
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor={field.name}>Email</FieldLabel>
                    <Input
                      {...field}
                      id={field.name}
                      type="email"
                      placeholder="m@example.com"
                      required
                      aria-invalid={fieldState.invalid}
                      disabled={isSubmitting}
                    />
                  </Field>
                )}
              />
              <Field>
                <Field className="grid grid-cols-2 gap-4">
                  <Controller
                    name="password"
                    control={control}
                    render={({ field, fieldState }) => (
                      <Field data-invalid={fieldState.invalid}>
                        <FieldLabel htmlFor={field.name}>Password</FieldLabel>
                        <Input
                          {...field}
                          id={field.name}
                          type="password"
                          required
                          aria-invalid={fieldState.invalid}
                          disabled={isSubmitting}
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
                          aria-invalid={fieldState.invalid}
                          disabled={isSubmitting}
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
                <Button type="submit" isLoading={isSubmitting}>
                  Create Account
                </Button>
                <FieldDescription className="text-center">
                  Already have an account? <Link href="/sign-in">Sign in</Link>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
      <FieldDescription className="px-6 text-center">
        By clicking continue, you agree to our{" "}
        <Link href="#">Terms of Service</Link> and{" "}
        <Link href="#">Privacy Policy</Link>.
      </FieldDescription>
    </div>
  );
}
