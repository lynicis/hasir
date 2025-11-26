import type { Mock } from "vitest";

import { Visibility } from "@buf/hasir_hasir.bufbuild_es/repository/v1/repository_pb";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, beforeEach, expect, it, vi } from "vitest";
import { ConnectError, Code } from "@connectrpc/connect";
import userEvent from "@testing-library/user-event";

import { useClient } from "@/lib/use-client";

import { RepositoryDialogForm } from "./repository-dialog-form";

// eslint-disable-next-line no-var
var toastSuccess: ReturnType<typeof vi.fn>;
// eslint-disable-next-line no-var
var toastError: ReturnType<typeof vi.fn>;

vi.mock("@/lib/use-client", () => ({
  useClient: vi.fn(),
}));

vi.mock("sonner", () => {
  toastSuccess = vi.fn();
  toastError = vi.fn();
  return {
    toast: {
      success: toastSuccess,
      error: toastError,
    },
  };
});

const mockCreateRepository = vi.fn();
const mockedUseClient = useClient as unknown as Mock;

describe("RepositoryDialogForm", () => {
  beforeEach(() => {
    mockedUseClient.mockReturnValue({
      createRepository: mockCreateRepository,
    } as never);
    mockCreateRepository.mockReset();
    toastSuccess.mockReset();
    toastError.mockReset();
  });

  function setup(open = true) {
    const onOpenChange = vi.fn();
    const onCancel = vi.fn();

    render(
      <RepositoryDialogForm
        open={open}
        onOpenChange={onOpenChange}
        onCancel={onCancel}
      />
    );

    return { onOpenChange, onCancel };
  }

  it("renders dialog with name and visibility fields when open", () => {
    setup(true);

    expect(
      screen.getByRole("heading", { name: /create repository/i })
    ).toBeInTheDocument();

    expect(screen.getByLabelText(/name/i)).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /public/i })).toBeInTheDocument();
    expect(screen.getByRole("radio", { name: /private/i })).toBeInTheDocument();
  });

  it("creates a public repository when submitted with valid data", async () => {
    const user = userEvent.setup();
    const { onOpenChange } = setup(true);

    await user.type(
      screen.getByLabelText(/name/i),
      "awesome-repository-example"
    );
    // public is selected by default, so we can submit directly
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(mockCreateRepository).toHaveBeenCalledWith({
        name: "awesome-repository-example",
        visibility: Visibility.PUBLIC,
      })
    );

    await waitFor(() =>
      expect(toastSuccess).toHaveBeenCalledWith(
        "Repository created successfully."
      )
    );
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("creates a private repository when private visibility is selected", async () => {
    const user = userEvent.setup();

    setup(true);

    await user.type(screen.getByLabelText(/name/i), "secret-repository");
    await user.click(screen.getByRole("radio", { name: /private/i }));
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(mockCreateRepository).toHaveBeenCalledWith({
        name: "secret-repository",
        visibility: Visibility.PRIVATE,
      })
    );
  });

  it("shows an error toast when a ConnectError occurs", async () => {
    const user = userEvent.setup();

    mockCreateRepository.mockRejectedValueOnce(
      new ConnectError("invalid name", Code.InvalidArgument)
    );

    setup(true);

    await user.type(screen.getByLabelText(/name/i), "bad name");
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith("[invalid_argument] invalid name")
    );
  });

  it("shows a generic error toast when an unexpected error occurs", async () => {
    const user = userEvent.setup();

    mockCreateRepository.mockRejectedValueOnce(new Error("network error"));

    setup(true);

    await user.type(screen.getByLabelText(/name/i), "any-name");
    await user.click(screen.getByRole("button", { name: /create/i }));

    await waitFor(() =>
      expect(toastError).toHaveBeenCalledWith("Failed to create repository.")
    );
  });

  it("calls onCancel and resets form when Cancel is clicked", async () => {
    const user = userEvent.setup();
    const { onCancel } = setup(true);

    const nameInput = screen.getByLabelText(/name/i) as HTMLInputElement;
    await user.type(nameInput, "temp-name");

    await user.click(screen.getByRole("button", { name: /cancel/i }));

    expect(onCancel).toHaveBeenCalled();
    // After cancel, the form should be reset back to default values
    expect(nameInput.value).toBe("");
  });
});

