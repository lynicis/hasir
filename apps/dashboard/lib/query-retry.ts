import { isNotFoundError } from "./utils";

export const customRetry = (failureCount: number, error: Error) => {
  if (!error) return false;

  if (isNotFoundError(error) || failureCount >= 3) {
    return false;
  }

  return true;
};
