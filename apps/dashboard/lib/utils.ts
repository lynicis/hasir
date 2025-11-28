import { Code, ConnectError } from "@connectrpc/connect";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function isNotFoundError(error: Error): error is ConnectError {
  return error instanceof ConnectError && error.code === Code.NotFound;
}
