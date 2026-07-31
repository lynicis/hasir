import { Code, ConnectError } from "@connectrpc/connect";

export function isNotFoundError(error: unknown): error is ConnectError {
  return error instanceof ConnectError && error.code === Code.NotFound;
}

export function isUnauthenticatedError(error: unknown): error is ConnectError {
  return error instanceof ConnectError && error.code === Code.Unauthenticated
}
