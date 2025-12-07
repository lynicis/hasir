"use client";

import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Settings } from "lucide-react";
import { useEffect } from "react";
import { z } from "zod/v4";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

const schema = z.object({
  name: z
    .string()
    .min(1, "Repository name is required")
    .max(100, "Repository name must be less than 100 characters")
    .regex(
      /^[a-zA-Z0-9-_]+$/,
      "Repository name can only contain letters, numbers, hyphens and underscores"
    ),
  visibility: z.enum(["private", "public", "internal"]),
});

type RepositorySettingsFormData = z.infer<typeof schema>;

export interface RepositorySettingsFormProps {
  initialData?: Partial<RepositorySettingsFormData>;
  onSubmit: (data: RepositorySettingsFormData) => Promise<void>;
  isLoading?: boolean;
}

export function RepositorySettingsForm({
  initialData,
  onSubmit,
  isLoading = false,
}: RepositorySettingsFormProps) {
  const { control, handleSubmit, formState, reset } =
    useForm<RepositorySettingsFormData>({
      resolver: zodResolver(schema),
      defaultValues: {
        name: initialData?.name || "",
        visibility: initialData?.visibility || "private",
      },
    });

  useEffect(() => {
    if (initialData) {
      reset({
        name: initialData.name || "",
        visibility: initialData.visibility || "private",
      });
    }
  }, [initialData, reset]);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-2">
          <Settings className="size-5 text-muted-foreground" />
          <CardTitle>General Settings</CardTitle>
        </div>
        <CardDescription>Configure basic repository settings</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="max-w-2xl space-y-6">
          <FieldGroup>
            <Controller
              control={control}
              name="name"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>Repository Name</FieldLabel>
                  <Input
                    {...field}
                    id={field.name}
                    placeholder="my-repository"
                    disabled={isLoading || formState.isSubmitting}
                  />
                  <FieldDescription>
                    The name of your repository. This will be used in URLs and
                    API references.
                  </FieldDescription>
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />

            <Controller
              control={control}
              name="visibility"
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={field.name}>Visibility</FieldLabel>
                  <Select
                    value={field.value}
                    onValueChange={field.onChange}
                    disabled={isLoading || formState.isSubmitting}
                  >
                    <SelectTrigger id={field.name}>
                      <SelectValue placeholder="Select visibility" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="private">Private</SelectItem>
                      <SelectItem value="public">Public</SelectItem>
                      <SelectItem value="internal">Internal</SelectItem>
                    </SelectContent>
                  </Select>
                  <FieldDescription>
                    Control who can see and access this repository.
                  </FieldDescription>
                  {fieldState.error && (
                    <FieldError errors={[fieldState.error]} />
                  )}
                </Field>
              )}
            />
          </FieldGroup>

          <div className="flex items-center gap-3">
            <Button
              type="submit"
              isLoading={isLoading || formState.isSubmitting}
            >
              Save Changes
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => reset()}
              disabled={isLoading || formState.isSubmitting}
            >
              Cancel
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
