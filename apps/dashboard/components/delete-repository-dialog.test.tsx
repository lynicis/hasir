import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { DeleteRepositoryDialog } from "./delete-repository-dialog";
import { type OrganizationRepository } from "./repository-item";

describe("DeleteRepositoryDialog", () => {
  const mockRepository: OrganizationRepository = {
    id: "1",
    name: "frontend-app",
    visibility: "public",
  };

  const mockOnOpenChange = vi.fn();
  const mockOnConfirm = vi.fn();

  it("renders dialog when open", () => { console.log("STARTED: renders dialog when open");  console.log("STARTED: renders dialog when open");  console.log("STARTED: renders dialog when open");  console.log("STARTED: renders dialog when open"); 
    render(
      <DeleteRepositoryDialog
        open={true}
        repository={mockRepository}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByRole("heading", { name: /delete repository/i })
    ).toBeInTheDocument();
    expect(screen.getByText(/frontend-app/i)).toBeInTheDocument();
  });

  it("does not render when closed", () => { console.log("STARTED: does not render when closed");  console.log("STARTED: does not render when closed");  console.log("STARTED: does not render when closed");  console.log("STARTED: does not render when closed"); 
    render(
      <DeleteRepositoryDialog
        open={false}
        repository={mockRepository}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.queryByRole("heading", { name: /delete repository/i })
    ).not.toBeInTheDocument();
  });

  it("calls onConfirm when confirm button is clicked", async () => { console.log("STARTED: calls onConfirm when confirm button is clicked");  console.log("STARTED: calls onConfirm when confirm button is clicked");  console.log("STARTED: calls onConfirm when confirm button is clicked");  console.log("STARTED: calls onConfirm when confirm button is clicked"); 
    const user = userEvent.setup();
    render(
      <DeleteRepositoryDialog
        open={true}
        repository={mockRepository}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    const confirmButton = screen.getByRole("button", {
      name: /delete repository/i,
    });
    await user.click(confirmButton);

    console.log("ASSERTING"); console.log("ASSERTING"); console.log("ASSERTING"); console.log("ASSERTING"); expect(mockOnConfirm).toHaveBeenCalledTimes(1);
  });

  it("calls onOpenChange when cancel button is clicked", async () => { console.log("STARTED: calls onOpenChange when cancel button is clicked");  console.log("STARTED: calls onOpenChange when cancel button is clicked");  console.log("STARTED: calls onOpenChange when cancel button is clicked");  console.log("STARTED: calls onOpenChange when cancel button is clicked"); 
    const user = userEvent.setup();
    render(
      <DeleteRepositoryDialog
        open={true}
        repository={mockRepository}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    const cancelButton = screen.getByRole("button", { name: /cancel/i });
    await user.click(cancelButton);

    console.log("ASSERTING"); console.log("ASSERTING"); console.log("ASSERTING"); console.log("ASSERTING"); expect(mockOnOpenChange).toHaveBeenCalled();
  });

  it("handles null repository gracefully", () => { console.log("STARTED: handles null repository gracefully");  console.log("STARTED: handles null repository gracefully");  console.log("STARTED: handles null repository gracefully");  console.log("STARTED: handles null repository gracefully"); 
    render(
      <DeleteRepositoryDialog
        open={true}
        repository={null}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByRole("heading", { name: /delete repository/i })
    ).toBeInTheDocument();
  });

  it("displays warning message about permanent deletion", () => { console.log("STARTED: displays warning message about permanent deletion");  console.log("STARTED: displays warning message about permanent deletion");  console.log("STARTED: displays warning message about permanent deletion");  console.log("STARTED: displays warning message about permanent deletion"); 
    render(
      <DeleteRepositoryDialog
        open={true}
        repository={mockRepository}
        onOpenChange={mockOnOpenChange}
        onConfirm={mockOnConfirm}
      />
    );

    expect(
      screen.getByText(
        /cannot be undone and all data will be permanently deleted/i
      )
    ).toBeInTheDocument();
  });
});
